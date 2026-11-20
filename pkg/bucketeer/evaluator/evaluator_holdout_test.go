package evaluator

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	evaluation "github.com/bucketeer-io/bucketeer/v2/evaluation/go"
	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	userproto "github.com/bucketeer-io/bucketeer/v2/proto/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

const (
	holdoutSegmentID = "segment-id-holdout"
)

var (
	// The HASH group below holds this user out
	userHeldOut = &user.User{ID: "user-id-24"}
	// The HASH group below doesn't hold this user out
	userNotHeldOut = &user.User{ID: "user-id-1"}

	holdoutSegment = &ftproto.SegmentUsers{
		SegmentId: holdoutSegmentID,
		Users: []*ftproto.SegmentUser{
			{
				SegmentId: holdoutSegmentID,
				State:     ftproto.SegmentUser_INCLUDED,
				UserId:    "user-id-1",
			},
		},
	}
)

// The group serves the given variation of the given features to the users it holds out.
func newHoldoutGroup(
	id string,
	assignmentType ftproto.HoldoutGroup_AssignmentType,
	holdoutVariationID string,
	featureIDs ...string,
) *ftproto.HoldoutGroup {
	group := &ftproto.HoldoutGroup{
		Id:             id,
		EnvironmentId:  "environment-id",
		AssignmentType: assignmentType,
		Status:         ftproto.HoldoutGroup_RUNNING,
		UpdatedAt:      1763500000,
	}
	switch assignmentType {
	case ftproto.HoldoutGroup_HASH:
		// It wraps: [97, 100) and [0, 2)
		group.SamplingSeed = "sampling-seed"
		group.BucketStart = 97
		group.Percentage = 5
	case ftproto.HoldoutGroup_SEGMENT:
		group.SegmentId = holdoutSegmentID
	}
	for _, featureID := range featureIDs {
		group.Features = append(group.Features, &ftproto.HoldoutGroupFeature{
			Id:                 id + ":" + featureID,
			HoldoutGroupId:     id,
			FeatureId:          featureID,
			EnvironmentId:      "environment-id",
			HoldoutVariationId: holdoutVariationID,
			AddedAt:            1763400000,
		})
	}
	return group
}

// errorLogs keeps the error logs of the SDK
type errorLogs struct {
	mu   sync.Mutex
	logs []string
}

func (l *errorLogs) Print(values ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, fmt.Sprint(values...))
}

func (l *errorLogs) Printf(format string, values ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, fmt.Sprintf(format, values...))
}

func (l *errorLogs) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string{}, l.logs...)
}

func newHoldoutEvaluator(
	t *testing.T,
	controller *gomock.Controller,
	groups []*ftproto.HoldoutGroup,
) (*evaluator, *errorLogs) {
	t.Helper()
	holdoutGroupsCache := cache.NewHoldoutGroupsCache(cache.NewInMemoryCache())
	require.NoError(t, holdoutGroupsCache.Put(groups))
	logs := &errorLogs{}
	return &evaluator{
		tag:                "server",
		featuresCache:      mock.NewMockFeaturesCache(controller),
		segmentUsersCache:  mock.NewMockSegmentUsersCache(controller),
		holdoutGroupsCache: holdoutGroupsCache,
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    logs,
		}),
	}, logs
}

// The users of the tests are fixed IDs. It fails when the HASH group stops matching them.
func TestHoldoutUsers(t *testing.T) {
	t.Parallel()
	group := newHoldoutGroup("holdout-group-id", ftproto.HoldoutGroup_HASH, "variation-false-id", ft2.Id)
	assert.True(t, evaluation.IsHeldOut(group, &userproto.User{Id: userHeldOut.ID}, nil))
	assert.False(t, evaluation.IsHeldOut(group, &userproto.User{Id: userNotHeldOut.ID}, nil))
}

func TestEvaluateHoldout(t *testing.T) {
	t.Parallel()

	const (
		hashGroupID    = "holdout-group-id-hash"
		segmentGroupID = "holdout-group-id-segment"
	)
	// ft2 is enabled and serves the true variation by default.
	// The groups serve the false variation to the users they hold out.
	hashGroup := newHoldoutGroup(hashGroupID, ftproto.HoldoutGroup_HASH, "variation-false-id", ft2.Id)
	segmentGroup := newHoldoutGroup(segmentGroupID, ftproto.HoldoutGroup_SEGMENT, "variation-false-id", ft2.Id)

	newEvaluation := func(
		featureID, userID, variationID string,
		reason *model.Reason,
	) *model.Evaluation {
		evaluation := &model.Evaluation{
			ID:          fmt.Sprintf("%s:0:%s", featureID, userID),
			FeatureID:   featureID,
			UserID:      userID,
			VariationID: variationID,
			Reason:      reason,
		}
		switch variationID {
		case "variation-true-id":
			evaluation.VariationValue = "true"
			evaluation.VariationName = "true-name"
		case "variation-false-id":
			evaluation.VariationValue = "false"
			evaluation.VariationName = "false-name"
		}
		return evaluation
	}

	patterns := []struct {
		desc        string
		groups      []*ftproto.HoldoutGroup
		setup       func(*evaluator)
		user        *user.User
		featureID   string
		expected    *model.Evaluation
		expectedErr error
		// Substrings of the only error log. Empty when nothing must be logged
		expectedLog []string
	}{
		{
			desc:   "hash: the held-out user gets the holdout variation with the reason and the group ID",
			groups: []*ftproto.HoldoutGroup{hashGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
			},
			user:      userHeldOut,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, userHeldOut.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: hashGroupID},
			),
		},
		{
			desc:   "hash: the user who is not held out is evaluated as without the group",
			groups: []*ftproto.HoldoutGroup{hashGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
			},
			user:      userNotHeldOut,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, userNotHeldOut.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonDefault},
			),
		},
		{
			desc:   "hash: the held-out user is evaluated as without the group for a flag the group doesn't target",
			groups: []*ftproto.HoldoutGroup{hashGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil)
			},
			user:      userHeldOut,
			featureID: ft5.Id,
			expected: newEvaluation(
				ft5.Id, userHeldOut.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonDefault},
			),
		},
		{
			desc: "hash: the holdout is evaluated before the flag is checked for being disabled",
			groups: []*ftproto.HoldoutGroup{
				// ft4 is disabled, and its off variation is the false one
				newHoldoutGroup(hashGroupID, ftproto.HoldoutGroup_HASH, "variation-true-id", ft4.Id),
			},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(ft4, nil)
			},
			user:      userHeldOut,
			featureID: ft4.Id,
			expected: newEvaluation(
				ft4.Id, userHeldOut.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: hashGroupID},
			),
		},
		{
			desc:   "hash: the holdout variation of a prerequisite flag is what the dependent flag sees",
			groups: []*ftproto.HoldoutGroup{hashGroup},
			setup: func(e *evaluator) {
				// ft1 requires the true variation of ft2
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft1.Id).Return(ft1, nil)
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
			},
			user:      userHeldOut,
			featureID: ft1.Id,
			expected: newEvaluation(
				ft1.Id, userHeldOut.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonPrerequisite},
			),
		},
		{
			desc:   "segment: the user in the cached segment gets the holdout variation with the reason and the group ID",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(holdoutSegment, nil)
			},
			user:      user1,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, user1.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: segmentGroupID},
			),
		},
		{
			desc:   "segment: the user who is not in the cached segment is evaluated as without the group",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(holdoutSegment, nil)
			},
			user:      user2,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, user2.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonDefault},
			),
		},
		{
			desc:   "segment: the segment is not cached yet. The user is evaluated as without the group, and it is logged",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(nil, cache.ErrNotFound)
			},
			// The user is in the segment on the server
			user:      user1,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, user1.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonDefault},
			),
			expectedLog: []string{"not cached yet", holdoutSegmentID, ft2.Id, user1.ID},
		},
		{
			desc:   "segment: err: failed to get the segment from cache",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(nil, internalErr)
			},
			user:        user1,
			featureID:   ft2.Id,
			expectedErr: internalErr,
		},
		{
			desc:   "segment: the segment is not read for a flag the group doesn't target",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				// Only the segment of the flag rule is read
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil)
			},
			user:      user1,
			featureID: ft5.Id,
			expected: newEvaluation(
				ft5.Id, user1.ID, "variation-true-id",
				&model.Reason{Type: model.ReasonDefault},
			),
		},
		{
			desc:   "segment: the segment is read when the group targets a prerequisite flag",
			groups: []*ftproto.HoldoutGroup{segmentGroup},
			setup: func(e *evaluator) {
				// ft1 requires the true variation of ft2
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft1.Id).Return(ft1, nil)
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(holdoutSegment, nil)
			},
			user:      user1,
			featureID: ft1.Id,
			expected: newEvaluation(
				ft1.Id, user1.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonPrerequisite},
			),
		},
		{
			desc: "segment: the segment is read once when a rule of the flag references it too",
			groups: []*ftproto.HoldoutGroup{
				{
					Id:             segmentGroupID,
					EnvironmentId:  "environment-id",
					AssignmentType: ftproto.HoldoutGroup_SEGMENT,
					SegmentId:      segment2.SegmentId,
					Status:         ftproto.HoldoutGroup_RUNNING,
					Features: []*ftproto.HoldoutGroupFeature{
						{FeatureId: ft5.Id, HoldoutVariationId: "variation-false-id"},
					},
				},
			},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil).Times(1)
			},
			// The user matches the rule of ft5, which serves the true variation
			user:      user2,
			featureID: ft5.Id,
			expected: newEvaluation(
				ft5.Id, user2.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: segmentGroupID},
			),
		},
		{
			desc: "all the groups of the list are evaluated",
			groups: []*ftproto.HoldoutGroup{
				newHoldoutGroup(hashGroupID, ftproto.HoldoutGroup_HASH, "variation-false-id", ft4.Id),
				segmentGroup,
			},
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(holdoutSegmentID).Return(holdoutSegment, nil)
			},
			user:      user1,
			featureID: ft2.Id,
			expected: newEvaluation(
				ft2.Id, user1.ID, "variation-false-id",
				&model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: segmentGroupID},
			),
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			// The controller belongs to the subtest, and Finish is not deferred in the parent test.
			// The parallel subtests run after the parent test function returns, so a Finish deferred
			// there would run before the expectations are set, and the missing calls would pass.
			evaluator, logs := newHoldoutEvaluator(t, gomock.NewController(t), p.groups)
			p.setup(evaluator)
			actual, err := evaluator.Evaluate(p.user, p.featureID)
			assert.Equal(t, p.expected, actual)
			assert.Equal(t, p.expectedErr, err)
			if len(p.expectedLog) == 0 {
				assert.Empty(t, logs.all())
				return
			}
			require.Len(t, logs.all(), 1)
			for _, expected := range p.expectedLog {
				assert.Contains(t, logs.all()[0], expected)
			}
		})
	}
}

func TestEvaluateHoldoutFailedToGetHoldoutGroups(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	evaluator := newEvaluator(t, "server", controller)
	holdoutGroupsCache := mock.NewMockHoldoutGroupsCache(controller)
	evaluator.holdoutGroupsCache = holdoutGroupsCache

	evaluator.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
	holdoutGroupsCache.EXPECT().Get().Return(nil, internalErr)

	actual, err := evaluator.Evaluate(userHeldOut, ft2.Id)
	assert.Nil(t, actual)
	assert.Equal(t, internalErr, err)
}

// The cached groups are shared by all the evaluations. They are never copied.
// The race detector fails this test if an evaluation modifies them.
func TestEvaluateHoldoutConcurrently(t *testing.T) {
	t.Parallel()
	// The caches encode what they store, so the features and the segment are not the ones
	// the other tests share.
	newFeature := func(id string, enabled bool, prerequisiteIDs ...string) *ftproto.Feature {
		feature := &ftproto.Feature{
			Id:      id,
			Enabled: enabled,
			Tags:    []string{"server"},
			Variations: []*ftproto.Variation{
				{Id: "variation-true-id", Name: "true-name", Value: "true"},
				{Id: "variation-false-id", Name: "false-name", Value: "false"},
			},
			DefaultStrategy: &ftproto.Strategy{
				FixedStrategy: &ftproto.FixedStrategy{Variation: "variation-true-id"},
			},
			OffVariation: "variation-false-id",
		}
		for _, prerequisiteID := range prerequisiteIDs {
			feature.Prerequisites = append(feature.Prerequisites, &ftproto.Prerequisite{
				FeatureId:   prerequisiteID,
				VariationId: "variation-true-id",
			})
		}
		return feature
	}
	const (
		enabledID   = "feature-id-enabled"
		disabledID  = "feature-id-disabled"
		dependentID = "feature-id-dependent"
		segmentID   = "segment-id-concurrent"
	)
	inMemoryCache := cache.NewInMemoryCache()
	featuresCache := cache.NewFeaturesCache(inMemoryCache)
	segmentUsersCache := cache.NewSegmentUsersCache(inMemoryCache)
	holdoutGroupsCache := cache.NewHoldoutGroupsCache(inMemoryCache)
	require.NoError(t, featuresCache.Put(newFeature(enabledID, true)))
	require.NoError(t, featuresCache.Put(newFeature(disabledID, false)))
	require.NoError(t, featuresCache.Put(newFeature(dependentID, true, enabledID)))
	require.NoError(t, segmentUsersCache.Put(&ftproto.SegmentUsers{
		SegmentId: segmentID,
		Users: []*ftproto.SegmentUser{
			{SegmentId: segmentID, State: ftproto.SegmentUser_INCLUDED, UserId: user1.ID},
		},
	}))
	segmentGroup := newHoldoutGroup(
		"holdout-group-id-segment", ftproto.HoldoutGroup_SEGMENT, "variation-false-id", enabledID,
	)
	segmentGroup.SegmentId = segmentID
	require.NoError(t, holdoutGroupsCache.Put([]*ftproto.HoldoutGroup{
		newHoldoutGroup("holdout-group-id-hash", ftproto.HoldoutGroup_HASH, "variation-true-id", disabledID),
		segmentGroup,
	}))
	evaluator := NewEvaluatorWithHoldoutGroups("server", featuresCache, segmentUsersCache, holdoutGroupsCache, log.NewLoggers(
		&log.LoggersConfig{ErrorLogger: log.DiscardErrorLogger},
	))

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				// The HASH group serves the true variation of the disabled flag
				actual, err := evaluator.Evaluate(userHeldOut, disabledID)
				if assert.NoError(t, err) {
					assert.Equal(t, "variation-true-id", actual.VariationID)
					assert.Equal(t, model.ReasonHoldout, actual.Reason.Type)
					assert.Equal(t, "holdout-group-id-hash", actual.Reason.HoldoutGroupID)
				}
				// The SEGMENT group serves the false variation of the enabled flag
				actual, err = evaluator.Evaluate(user1, enabledID)
				if assert.NoError(t, err) {
					assert.Equal(t, "variation-false-id", actual.VariationID)
					assert.Equal(t, model.ReasonHoldout, actual.Reason.Type)
					assert.Equal(t, "holdout-group-id-segment", actual.Reason.HoldoutGroupID)
				}
				// The dependent flag requires the true variation of the enabled flag
				actual, err = evaluator.Evaluate(user1, dependentID)
				if assert.NoError(t, err) {
					assert.Equal(t, "variation-false-id", actual.VariationID)
					assert.Equal(t, model.ReasonPrerequisite, actual.Reason.Type)
					assert.Empty(t, actual.Reason.HoldoutGroupID)
				}
				// The user is in no group
				actual, err = evaluator.Evaluate(user2, enabledID)
				if assert.NoError(t, err) {
					assert.Equal(t, "variation-true-id", actual.VariationID)
					assert.Equal(t, model.ReasonDefault, actual.Reason.Type)
					assert.Empty(t, actual.Reason.HoldoutGroupID)
				}
			}
		}()
	}
	wg.Wait()
}

func TestHoldoutSegmentIDs(t *testing.T) {
	t.Parallel()
	hashGroup := newHoldoutGroup("holdout-group-id-hash", ftproto.HoldoutGroup_HASH, "variation-false-id", ft1.Id)
	segmentGroup := newHoldoutGroup("holdout-group-id-segment", ftproto.HoldoutGroup_SEGMENT, "variation-false-id", ft2.Id)
	removedTarget := newHoldoutGroup("holdout-group-id-removed", ftproto.HoldoutGroup_SEGMENT, "variation-false-id", ft2.Id)
	removedTarget.SegmentId = "segment-id-removed"
	removedTarget.Features[0].RemovedAt = 1763500000

	patterns := []struct {
		desc     string
		groups   []*ftproto.HoldoutGroup
		features []*ftproto.Feature
		expected []string
	}{
		{
			desc:     "no group",
			groups:   nil,
			features: []*ftproto.Feature{ft1, ft2},
			expected: nil,
		},
		{
			desc:     "a hash group has no segment",
			groups:   []*ftproto.HoldoutGroup{hashGroup},
			features: []*ftproto.Feature{ft1, ft2},
			expected: nil,
		},
		{
			desc:     "a segment group that targets one of the features",
			groups:   []*ftproto.HoldoutGroup{hashGroup, segmentGroup},
			features: []*ftproto.Feature{ft1, ft2},
			expected: []string{holdoutSegmentID},
		},
		{
			desc:     "a segment group that targets none of the features",
			groups:   []*ftproto.HoldoutGroup{hashGroup, segmentGroup},
			features: []*ftproto.Feature{ft1, ft4},
			expected: nil,
		},
		{
			desc:     "a segment group without targets",
			groups:   []*ftproto.HoldoutGroup{newHoldoutGroup("holdout-group-id", ftproto.HoldoutGroup_SEGMENT, "")},
			features: []*ftproto.Feature{ft1, ft2},
			expected: nil,
		},
		{
			desc:     "nil entries are skipped",
			groups:   []*ftproto.HoldoutGroup{nil, {AssignmentType: ftproto.HoldoutGroup_SEGMENT, SegmentId: "id", Features: []*ftproto.HoldoutGroupFeature{nil}}},
			features: []*ftproto.Feature{ft1, ft2},
			expected: nil,
		},
		{
			// The evaluation module decides whether a removed target counts.
			// Loading a segment it doesn't read is harmless. Not loading one it reads is not.
			desc:     "the segment is never dropped for a reason the evaluation module owns",
			groups:   []*ftproto.HoldoutGroup{removedTarget},
			features: []*ftproto.Feature{ft2},
			expected: []string{"segment-id-removed"},
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, p.expected, holdoutSegmentIDs(p.groups, p.features))
		})
	}
}

// The segment must be loaded for every feature the evaluation module serves the holdout variation for.
// It evaluates the features with all the users of the segment, with and without the segment loaded by
// holdoutSegmentIDs, and expects the same result as with the segment always loaded.
func TestHoldoutSegmentIDsCoversTheEvaluationModule(t *testing.T) {
	t.Parallel()
	group := newHoldoutGroup("holdout-group-id", ftproto.HoldoutGroup_SEGMENT, "variation-false-id", ft2.Id)
	allSegmentUsers := map[string][]*ftproto.SegmentUser{holdoutSegmentID: holdoutSegment.Users}
	u := &userproto.User{Id: "user-id-1"}

	for _, features := range [][]*ftproto.Feature{{ft2}, {ft1, ft2}, {ft4}} {
		ids := make([]string, 0, len(features))
		for _, f := range features {
			ids = append(ids, f.Id)
		}
		t.Run(strings.Join(ids, ","), func(t *testing.T) {
			loaded := map[string][]*ftproto.SegmentUser{}
			for _, id := range holdoutSegmentIDs([]*ftproto.HoldoutGroup{group}, features) {
				loaded[id] = allSegmentUsers[id]
			}
			expected, err := evaluation.NewEvaluator().EvaluateFeatures(
				append([]*ftproto.Feature{}, features...), u, allSegmentUsers, nil, "server", group,
			)
			require.NoError(t, err)
			actual, err := evaluation.NewEvaluator().EvaluateFeatures(
				append([]*ftproto.Feature{}, features...), u, loaded, nil, "server", group,
			)
			require.NoError(t, err)
			require.Len(t, actual.Evaluations, len(expected.Evaluations))
			for i := range expected.Evaluations {
				assert.Equal(t, expected.Evaluations[i].FeatureId, actual.Evaluations[i].FeatureId)
				assert.Equal(t, expected.Evaluations[i].VariationId, actual.Evaluations[i].VariationId)
				assert.Equal(t, expected.Evaluations[i].Reason.Type, actual.Evaluations[i].Reason.Type)
			}
		})
	}
}

// NewEvaluator is exported, so its signature must not change:
// the callers written before the holdout groups would stop compiling.
var _ func(string, cache.FeaturesCache, cache.SegmentUsersCache) EvaluateLocally = NewEvaluator

func TestNewEvaluatorConstructors(t *testing.T) {
	t.Parallel()
	hashGroup := newHoldoutGroup("holdout-group-id", ftproto.HoldoutGroup_HASH, "variation-false-id", "feature-id")
	require.True(t, evaluation.IsHeldOut(hashGroup, &userproto.User{Id: userHeldOut.ID}, nil))
	// The segment of the group is not cached, so the evaluator logs it
	segmentGroup := newHoldoutGroup("holdout-group-id", ftproto.HoldoutGroup_SEGMENT, "variation-false-id", "feature-id")
	feature := &ftproto.Feature{
		Id:            "feature-id",
		Enabled:       true,
		Version:       1,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_BOOLEAN,
		Variations: []*ftproto.Variation{
			{Id: "variation-true-id", Value: "true"},
			{Id: "variation-false-id", Value: "false"},
		},
		DefaultStrategy: &ftproto.Strategy{
			Type:          ftproto.Strategy_FIXED,
			FixedStrategy: &ftproto.FixedStrategy{Variation: "variation-true-id"},
		},
		OffVariation: "variation-false-id",
	}
	notHeldOut := &model.Evaluation{
		ID:             "feature-id:1:" + userHeldOut.ID,
		FeatureID:      "feature-id",
		FeatureVersion: 1,
		UserID:         userHeldOut.ID,
		VariationID:    "variation-true-id",
		VariationValue: "true",
		Reason:         &model.Reason{Type: model.ReasonDefault},
	}
	heldOut := &model.Evaluation{
		ID:             "feature-id:1:" + userHeldOut.ID,
		FeatureID:      "feature-id",
		FeatureVersion: 1,
		UserID:         userHeldOut.ID,
		VariationID:    "variation-false-id",
		VariationValue: "false",
		Reason:         &model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: "holdout-group-id"},
	}
	loggers := log.NewLoggers(&log.LoggersConfig{ErrorLogger: log.DiscardErrorLogger})

	patterns := []struct {
		desc     string
		group    *ftproto.HoldoutGroup
		new      func(cache.FeaturesCache, cache.SegmentUsersCache, cache.HoldoutGroupsCache) EvaluateLocally
		expected *model.Evaluation
	}{
		{
			desc:  "NewEvaluator: the holdout groups are not applied",
			group: hashGroup,
			new: func(fc cache.FeaturesCache, sc cache.SegmentUsersCache, _ cache.HoldoutGroupsCache) EvaluateLocally {
				return NewEvaluator("server", fc, sc)
			},
			expected: notHeldOut,
		},
		{
			desc:  "NewEvaluatorWithHoldoutGroups: the holdout groups are applied",
			group: hashGroup,
			new: func(fc cache.FeaturesCache, sc cache.SegmentUsersCache, hc cache.HoldoutGroupsCache) EvaluateLocally {
				return NewEvaluatorWithHoldoutGroups("server", fc, sc, hc, loggers)
			},
			expected: heldOut,
		},
		{
			desc:  "NewEvaluatorWithHoldoutGroups: the holdout groups are applied without the loggers",
			group: hashGroup,
			new: func(fc cache.FeaturesCache, sc cache.SegmentUsersCache, hc cache.HoldoutGroupsCache) EvaluateLocally {
				return NewEvaluatorWithHoldoutGroups("server", fc, sc, hc, nil)
			},
			expected: heldOut,
		},
		{
			desc:  "NewEvaluatorWithHoldoutGroups: nothing is logged without the loggers",
			group: segmentGroup,
			new: func(fc cache.FeaturesCache, sc cache.SegmentUsersCache, hc cache.HoldoutGroupsCache) EvaluateLocally {
				return NewEvaluatorWithHoldoutGroups("server", fc, sc, hc, nil)
			},
			expected: notHeldOut,
		},
		{
			desc:  "NewEvaluatorWithHoldoutGroups: no holdout group is applied without the cache",
			group: hashGroup,
			new: func(fc cache.FeaturesCache, sc cache.SegmentUsersCache, _ cache.HoldoutGroupsCache) EvaluateLocally {
				return NewEvaluatorWithHoldoutGroups("server", fc, sc, nil, loggers)
			},
			expected: notHeldOut,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			inMemoryCache := cache.NewInMemoryCache()
			featuresCache := cache.NewFeaturesCache(inMemoryCache)
			require.NoError(t, featuresCache.Put(feature))
			holdoutGroupsCache := cache.NewHoldoutGroupsCache(inMemoryCache)
			require.NoError(t, holdoutGroupsCache.Put([]*ftproto.HoldoutGroup{p.group}))

			evaluator := p.new(featuresCache, cache.NewSegmentUsersCache(inMemoryCache), holdoutGroupsCache)
			actual, err := evaluator.Evaluate(&user.User{ID: userHeldOut.ID}, "feature-id")
			require.NoError(t, err)
			assert.Equal(t, p.expected, actual)
		})
	}
}
