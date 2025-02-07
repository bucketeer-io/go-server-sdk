package model

import (
	"strconv"

	"github.com/bucketeer-io/bucketeer/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
)

type GetSegmentUsersResponse struct {
	SegmentUsers      []SegmentUsers `json:"segmentUsers"`
	DeletedSegmentIDs []string       `json:"deletedSegmentIds"`
	RequestedAt       string         `json:"requestedAt"`
	ForceUpdate       bool           `json:"forceUpdate"`
}

type SegmentUsers struct {
	SegmentID string        `json:"segmentId"`
	Users     []SegmentUser `json:"users"`
	UpdatedAt string        `json:"updatedAt"`
}

type SegmentUser struct {
	ID        string `json:"id"`
	SegmentID string `json:"segmentId"`
	UserID    string `json:"userId"`
	State     string `json:"state"`
	Deleted   bool   `json:"deleted"`
}

func ConvertSegmentUsersResponse(response *GetSegmentUsersResponse) *gwproto.GetSegmentUsersResponse {
	requestedAt, _ := strconv.ParseInt(response.RequestedAt, 10, 64)
	pbResp := &gwproto.GetSegmentUsersResponse{
		SegmentUsers:      mapFields(response.SegmentUsers, convertSegmentUsersModel),
		DeletedSegmentIds: response.DeletedSegmentIDs,
		RequestedAt:       requestedAt,
		ForceUpdate:       response.ForceUpdate,
	}
	return pbResp
}

func convertSegmentUsersModel(s SegmentUsers) *feature.SegmentUsers {
	updatedAt, _ := strconv.ParseInt(s.UpdatedAt, 10, 64)
	pbSegmentUsers := &feature.SegmentUsers{
		SegmentId: s.SegmentID,
		Users: mapFields(s.Users, func(u SegmentUser) *feature.SegmentUser {
			return &feature.SegmentUser{
				Id:        u.ID,
				SegmentId: u.SegmentID,
				UserId:    u.UserID,
				State:     feature.SegmentUser_State(feature.SegmentUser_State_value[u.State]),
				Deleted:   u.Deleted,
			}
		}),
		UpdatedAt: updatedAt,
	}
	return pbSegmentUsers
}
