package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/bucketeer/v2/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/v2/proto/gateway"
)

func TestConvertSegmentUsersResponse(t *testing.T) {
	response := &GetSegmentUsersResponse{
		SegmentUsers: []SegmentUsers{
			{
				SegmentID: "segment-1",
				Users: []SegmentUser{
					{
						ID:        "user-1",
						SegmentID: "segment-1",
						UserID:    "user-1-id",
						State:     "EXCLUDED",
						Deleted:   false,
					},
				},
				UpdatedAt: "1620000000",
			},
			{
				SegmentID: "segment-2",
				Users: []SegmentUser{
					{
						ID:        "user-2",
						SegmentID: "segment-2",
						UserID:    "user-2-id",
						State:     "INCLUDED",
						Deleted:   false,
					},
				},
				UpdatedAt: "1620000001",
				Rules: []Rule{
					{
						ID: "rule-1",
						Clauses: []Clause{
							{
								ID:        "clause-1",
								Attribute: "country",
								Operator:  "EQUALS",
								Values:    []string{"japan"},
							},
							{
								ID:        "clause-2",
								Attribute: "age",
								Operator:  "GREATER",
								Values:    []string{"20"},
							},
						},
					},
					{
						ID: "rule-2",
						Clauses: []Clause{
							{
								ID:        "clause-3",
								Attribute: "email",
								Operator:  "STARTS_WITH",
								Values:    []string{"test@"},
							},
						},
					},
				},
			},
		},
		DeletedSegmentIDs: []string{"deleted-segment-1", "deleted-segment-2"},
		RequestedAt:       "1620000000",
		ForceUpdate:       true,
	}

	expected := &gwproto.GetSegmentUsersResponse{
		SegmentUsers: []*feature.SegmentUsers{
			{
				SegmentId: "segment-1",
				Users: []*feature.SegmentUser{
					{
						Id:        "user-1",
						SegmentId: "segment-1",
						UserId:    "user-1-id",
						State:     feature.SegmentUser_EXCLUDED,
						Deleted:   false,
					},
				},
				UpdatedAt: 1620000000,
				Rules:     []*feature.Rule{},
			},
			{
				SegmentId: "segment-2",
				Users: []*feature.SegmentUser{
					{
						Id:        "user-2",
						SegmentId: "segment-2",
						UserId:    "user-2-id",
						State:     feature.SegmentUser_INCLUDED,
						Deleted:   false,
					},
				},
				UpdatedAt: 1620000001,
				Rules: []*feature.Rule{
					{
						Id: "rule-1",
						Clauses: []*feature.Clause{
							{
								Id:        "clause-1",
								Attribute: "country",
								Operator:  feature.Clause_EQUALS,
								Values:    []string{"japan"},
							},
							{
								Id:        "clause-2",
								Attribute: "age",
								Operator:  feature.Clause_GREATER,
								Values:    []string{"20"},
							},
						},
					},
					{
						Id: "rule-2",
						Clauses: []*feature.Clause{
							{
								Id:        "clause-3",
								Attribute: "email",
								Operator:  feature.Clause_STARTS_WITH,
								Values:    []string{"test@"},
							},
						},
					},
				},
			},
		},
		DeletedSegmentIds: []string{"deleted-segment-1", "deleted-segment-2"},
		RequestedAt:       1620000000,
		ForceUpdate:       true,
	}

	actual := ConvertSegmentUsersResponse(response)

	assert.Equal(t, expected, actual)
}
