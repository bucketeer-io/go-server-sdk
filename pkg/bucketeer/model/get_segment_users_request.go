package model

import (
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

type GetSegmentUsersRequest struct {
	SegmentIDs  []string     `json:"segmentIds"`
	RequestedAt int64        `json:"requestedAt"`
	SourceID    SourceIDType `json:"sourceId"`
	SDKVersion  string       `json:"sdkVersion"`
}

func NewGetSegmentUsersRequest(segmentIDs []string, requestedAt int64) *GetSegmentUsersRequest {
	return &GetSegmentUsersRequest{
		SegmentIDs:  segmentIDs,
		RequestedAt: requestedAt,
		SourceID:    SourceIDGoServer,
		SDKVersion:  version.SDKVersion,
	}
}
