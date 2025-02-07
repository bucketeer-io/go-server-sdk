package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/bucketeer/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
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
			},
		},
		DeletedSegmentIds: []string{"deleted-segment-1", "deleted-segment-2"},
		RequestedAt:       1620000000,
		ForceUpdate:       true,
	}

	actual := ConvertSegmentUsersResponse(response)

	assert.Equal(t, expected, actual)
}
