package model

import "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"

type RegisterEventsRequest struct {
	Events     []*Event     `json:"events,omitempty"`
	SDKVersion string       `json:"sdkVersion,omitempty"`
	SourceID   SourceIDType `json:"sourceId,omitempty"`
}

func NewRegisterEventsRequest(event []*Event) *RegisterEventsRequest {
	return &RegisterEventsRequest{
		Events:     event,
		SDKVersion: version.SDKVersion,
		SourceID:   SourceIDGoServer,
	}
}
