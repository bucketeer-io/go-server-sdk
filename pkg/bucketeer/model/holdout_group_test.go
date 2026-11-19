package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	evaluation "github.com/bucketeer-io/bucketeer/v2/evaluation/go"
	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/v2/proto/gateway"
	userproto "github.com/bucketeer-io/bucketeer/v2/proto/user"
)

// The API gateway encodes the response with protojson and emits the unpopulated fields.
var gatewayMarshalOptions = protojson.MarshalOptions{EmitUnpopulated: true}

// The groups hold only the fields the server sends to the SDK.
// See sdkHoldoutGroups in pkg/api/api/holdout.go of the bucketeer repository.
func newHoldoutGroupProtos(t *testing.T) []*ftproto.HoldoutGroup {
	t.Helper()
	return []*ftproto.HoldoutGroup{
		{
			Id:             "holdout-group-id-hash",
			EnvironmentId:  "environment-id",
			AssignmentType: ftproto.HoldoutGroup_HASH,
			SamplingSeed:   "sampling-seed",
			BucketStart:    97.125,
			Percentage:     5.5,
			Status:         ftproto.HoldoutGroup_RUNNING,
			UpdatedAt:      1763500000,
			Features: []*ftproto.HoldoutGroupFeature{
				{
					Id:                 "holdout-group-feature-id-1",
					HoldoutGroupId:     "holdout-group-id-hash",
					FeatureId:          "feature-id-1",
					EnvironmentId:      "environment-id",
					HoldoutVariationId: "variation-id-1",
					AddedAt:            1763400000,
				},
				{
					Id:                 "holdout-group-feature-id-2",
					HoldoutGroupId:     "holdout-group-id-hash",
					FeatureId:          "feature-id-2",
					EnvironmentId:      "environment-id",
					HoldoutVariationId: "variation-id-2",
					AddedAt:            1763400001,
				},
			},
		},
		{
			Id:             "holdout-group-id-segment",
			EnvironmentId:  "environment-id",
			AssignmentType: ftproto.HoldoutGroup_SEGMENT,
			SegmentId:      "segment-id",
			Status:         ftproto.HoldoutGroup_ANALYSIS,
			UpdatedAt:      1763500001,
			// A group can serve before it has a target.
			Features: []*ftproto.HoldoutGroupFeature{},
		},
	}
}

// It decodes the JSON the API gateway produces from the proto messages,
// so the SDK model is checked against proto/feature/holdout.proto, not against a copy of it.
func TestConvertFeatureFlagsResponseHoldoutGroupsFromGatewayJSON(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc     string
		response *gwproto.GetFeatureFlagsResponse
	}{
		{
			desc: "the groups come with a flag update",
			response: &gwproto.GetFeatureFlagsResponse{
				// The ID has a holdout part while a group is serving. The SDK keeps it as is.
				FeatureFlagsId:         "1234567890:9876543210",
				Features:               []*ftproto.Feature{{Id: "feature-id-1"}},
				ArchivedFeatureFlagIds: []string{},
				RequestedAt:            1763500002,
				ForceUpdate:            true,
				HoldoutGroups:          newHoldoutGroupProtos(t),
			},
		},
		{
			desc: "the groups come with an empty flag diff",
			response: &gwproto.GetFeatureFlagsResponse{
				FeatureFlagsId:         "1234567890:9876543210",
				Features:               []*ftproto.Feature{},
				ArchivedFeatureFlagIds: []string{},
				RequestedAt:            1763500002,
				HoldoutGroups:          newHoldoutGroupProtos(t),
			},
		},
		{
			desc: "no group is serving",
			response: &gwproto.GetFeatureFlagsResponse{
				FeatureFlagsId:         "1234567890",
				Features:               []*ftproto.Feature{},
				ArchivedFeatureFlagIds: []string{},
				RequestedAt:            1763500002,
				HoldoutGroups:          []*ftproto.HoldoutGroup{},
			},
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			encoded, err := gatewayMarshalOptions.Marshal(p.response)
			require.NoError(t, err)

			var decoded GetFeatureFlagsResponse
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			actual := ConvertFeatureFlagsResponse(&decoded)

			assert.Equal(t, p.response.FeatureFlagsId, actual.FeatureFlagsId)
			assert.Equal(t, p.response.RequestedAt, actual.RequestedAt)
			assert.Equal(t, p.response.ForceUpdate, actual.ForceUpdate)
			require.Len(t, actual.HoldoutGroups, len(p.response.HoldoutGroups))
			for i, expected := range p.response.HoldoutGroups {
				assert.Truef(
					t,
					proto.Equal(expected, actual.HoldoutGroups[i]),
					"Expected: %v, Actual: %v", expected, actual.HoldoutGroups[i],
				)
			}
		})
	}
}

func TestConvertFeatureFlagsResponseHoldoutGroupsFromJSON(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc     string
		json     string
		expected []*ftproto.HoldoutGroup
	}{
		{
			desc: "a HASH group",
			json: `{
				"featureFlagsId": "1234567890:9876543210",
				"features": [],
				"archivedFeatureFlagIds": [],
				"requestedAt": "1763500002",
				"forceUpdate": false,
				"holdoutGroups": [
					{
						"id": "holdout-group-id",
						"environmentId": "environment-id",
						"assignmentType": "HASH",
						"samplingSeed": "sampling-seed",
						"bucketStart": 97.125,
						"percentage": 5.5,
						"segmentId": "",
						"status": "RUNNING",
						"updatedAt": "1763500000",
						"features": [
							{
								"id": "holdout-group-feature-id",
								"holdoutGroupId": "holdout-group-id",
								"featureId": "feature-id",
								"environmentId": "environment-id",
								"holdoutVariationId": "variation-id",
								"addedAt": "1763400000",
								"removedAt": "0"
							}
						]
					}
				]
			}`,
			expected: []*ftproto.HoldoutGroup{
				{
					Id:             "holdout-group-id",
					EnvironmentId:  "environment-id",
					AssignmentType: ftproto.HoldoutGroup_HASH,
					SamplingSeed:   "sampling-seed",
					BucketStart:    97.125,
					Percentage:     5.5,
					Status:         ftproto.HoldoutGroup_RUNNING,
					UpdatedAt:      1763500000,
					Features: []*ftproto.HoldoutGroupFeature{
						{
							Id:                 "holdout-group-feature-id",
							HoldoutGroupId:     "holdout-group-id",
							FeatureId:          "feature-id",
							EnvironmentId:      "environment-id",
							HoldoutVariationId: "variation-id",
							AddedAt:            1763400000,
						},
					},
				},
			},
		},
		{
			desc: "a SEGMENT group",
			json: `{
				"featureFlagsId": "1234567890:9876543210",
				"holdoutGroups": [
					{
						"id": "holdout-group-id",
						"environmentId": "environment-id",
						"assignmentType": "SEGMENT",
						"segmentId": "segment-id",
						"status": "ANALYSIS",
						"updatedAt": "1763500000",
						"features": []
					}
				]
			}`,
			expected: []*ftproto.HoldoutGroup{
				{
					Id:             "holdout-group-id",
					EnvironmentId:  "environment-id",
					AssignmentType: ftproto.HoldoutGroup_SEGMENT,
					SegmentId:      "segment-id",
					Status:         ftproto.HoldoutGroup_ANALYSIS,
					UpdatedAt:      1763500000,
				},
			},
		},
		{
			desc:     "the list is empty: no group is serving",
			json:     `{"featureFlagsId": "1234567890", "holdoutGroups": []}`,
			expected: []*ftproto.HoldoutGroup{},
		},
		{
			desc:     "the field is absent: the server doesn't support holdout groups",
			json:     `{"featureFlagsId": "1234567890", "features": [], "requestedAt": "1763500002"}`,
			expected: []*ftproto.HoldoutGroup{},
		},
		{
			desc:     "the field is null",
			json:     `{"featureFlagsId": "1234567890", "holdoutGroups": null}`,
			expected: []*ftproto.HoldoutGroup{},
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			t.Parallel()
			var decoded GetFeatureFlagsResponse
			require.NoError(t, json.Unmarshal([]byte(p.json), &decoded))
			actual := ConvertFeatureFlagsResponse(&decoded)
			require.Len(t, actual.HoldoutGroups, len(p.expected))
			for i, expected := range p.expected {
				assert.Truef(
					t,
					proto.Equal(expected, actual.HoldoutGroups[i]),
					"Expected: %v, Actual: %v", expected, actual.HoldoutGroups[i],
				)
			}
		})
	}
}

// An assignment type added by a newer server must not be read as HASH, which is the zero value.
func TestConvertHoldoutGroupModelUnknownAssignmentType(t *testing.T) {
	t.Parallel()
	group := convertHoldoutGroupModel(HoldoutGroup{
		ID:             "holdout-group-id",
		EnvironmentID:  "environment-id",
		AssignmentType: "ADDED_BY_A_NEWER_SERVER",
		SamplingSeed:   "sampling-seed",
		BucketStart:    0,
		Percentage:     50,
		SegmentID:      "segment-id",
		Status:         "RUNNING",
	})
	assert.NotEqual(t, ftproto.HoldoutGroup_HASH, group.AssignmentType)
	assert.NotEqual(t, ftproto.HoldoutGroup_SEGMENT, group.AssignmentType)
	_, known := ftproto.HoldoutGroup_AssignmentType_name[int32(group.AssignmentType)]
	assert.False(t, known)

	// The same group holds users out when it is a HASH group, and nobody when the type is unknown.
	hash := proto.Clone(group).(*ftproto.HoldoutGroup)
	hash.AssignmentType = ftproto.HoldoutGroup_HASH
	segmentUsers := []*ftproto.SegmentUser{}
	heldOutAsHash := 0
	for _, id := range []string{"user-id-1", "user-id-2", "user-id-3", "user-id-4", "user-id-5", "user-id-6"} {
		user := &userproto.User{Id: id}
		segmentUsers = append(segmentUsers, &ftproto.SegmentUser{
			SegmentId: "segment-id",
			UserId:    id,
			State:     ftproto.SegmentUser_INCLUDED,
		})
		if evaluation.IsHeldOut(hash, user, nil) {
			heldOutAsHash++
		}
		assert.False(t, evaluation.IsHeldOut(group, user, segmentUsers), id)
	}
	assert.Positive(t, heldOutAsHash)
}
