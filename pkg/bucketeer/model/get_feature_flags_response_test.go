package model

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
)

func TestConvertFeatureFlagsResponse(t *testing.T) {
	response := &GetFeatureFlagsResponse{
		FeatureFlagsID: "test-id",
		Features: []Feature{
			{
				ID:          "feature-1",
				Name:        "Feature 1",
				Description: "Description 1",
				Enabled:     true,
				Deleted:     false,
				TTL:         100,
				Version:     1,
				CreatedAt:   "1620000000",
				UpdatedAt:   "1620000001",
				Variations: []Variation{
					{
						ID:          "variation-1",
						Value:       "value-1",
						Name:        "Variation 1",
						Description: "Description 1",
					},
				},
				Targets: []Target{
					{
						Variation: "variation-1",
						Users:     []string{"user-1"},
					},
				},
				Rules: []Rule{
					{
						ID: "rule-1",
						Strategy: &Strategy{
							Type: "fixed",
							FixedStrategy: &FixedStrategy{
								Variation: "variation-1",
							},
						},
						Clauses: []Clause{
							{
								ID:        "clause-1",
								Attribute: "attr-1",
								Operator:  "EQUALS",
								Values:    []string{"value-1"},
							},
						},
					},
				},
				DefaultStrategy: &Strategy{
					Type: "fixed",
					FixedStrategy: &FixedStrategy{
						Variation: "variation-1",
					},
				},
				OffVariation:  "variation-1",
				Tags:          []string{"tag-1"},
				Maintainer:    "maintainer-1",
				VariationType: "BOOLEAN",
				Archived:      false,
				Prerequisites: []*Prerequisite{
					{
						FeatureID:   "feature-2",
						VariationID: "variation-2",
					},
				},
				SamplingSeed: "seed-1",
				LastUsedInfo: &FeatureLastUsedInfo{
					FeatureID:           "feature-1",
					Version:             1,
					LastUsedAt:          "1620000002",
					CreatedAt:           "1620000000",
					ClientOldestVersion: "1.0.0",
					ClientLatestVersion: "1.0.1",
				},
			},
		},
		ArchivedFeatureFlagIDs: []string{"archived-1"},
		RequestedAt:            "1620000000",
		ForceUpdate:            true,
	}

	expected := &gwproto.GetFeatureFlagsResponse{
		FeatureFlagsId: "test-id",
		Features: []*ftproto.Feature{
			{
				Id:          "feature-1",
				Name:        "Feature 1",
				Description: "Description 1",
				Enabled:     true,
				Deleted:     false,
				Ttl:         100,
				Version:     1,
				CreatedAt:   1620000000,
				UpdatedAt:   1620000001,
				Variations: []*ftproto.Variation{
					{
						Id:          "variation-1",
						Value:       "value-1",
						Name:        "Variation 1",
						Description: "Description 1",
					},
				},
				Targets: []*ftproto.Target{
					{
						Variation: "variation-1",
						Users:     []string{"user-1"},
					},
				},
				Rules: []*ftproto.Rule{
					{
						Id: "rule-1",
						Strategy: &ftproto.Strategy{
							Type: ftproto.Strategy_FIXED,
							FixedStrategy: &ftproto.FixedStrategy{
								Variation: "variation-1",
							},
							RolloutStrategy: &ftproto.RolloutStrategy{},
						},
						Clauses: []*ftproto.Clause{
							{
								Id:        "clause-1",
								Attribute: "attr-1",
								Operator:  ftproto.Clause_EQUALS,
								Values:    []string{"value-1"},
							},
						},
					},
				},
				DefaultStrategy: &ftproto.Strategy{
					Type: ftproto.Strategy_FIXED,
					FixedStrategy: &ftproto.FixedStrategy{
						Variation: "variation-1",
					},
					RolloutStrategy: &ftproto.RolloutStrategy{},
				},
				OffVariation:  "variation-1",
				Tags:          []string{"tag-1"},
				Maintainer:    "maintainer-1",
				VariationType: ftproto.Feature_BOOLEAN,
				Archived:      false,
				Prerequisites: []*ftproto.Prerequisite{
					{
						FeatureId:   "feature-2",
						VariationId: "variation-2",
					},
				},
				SamplingSeed: "seed-1",
				LastUsedInfo: &ftproto.FeatureLastUsedInfo{
					FeatureId:           "feature-1",
					Version:             1,
					LastUsedAt:          1620000002,
					CreatedAt:           1620000000,
					ClientOldestVersion: "1.0.0",
					ClientLatestVersion: "1.0.1",
				},
			},
		},
		ArchivedFeatureFlagIds: []string{"archived-1"},
		RequestedAt:            1620000000,
		ForceUpdate:            true,
	}

	actual := ConvertFeatureFlagsResponse(response)

	assert.True(t, reflect.DeepEqual(expected, actual), "Expected: %+v, Actual: %+v", expected, actual)
}
