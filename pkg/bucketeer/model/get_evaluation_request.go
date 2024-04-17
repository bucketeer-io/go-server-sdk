package model

import (
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

type GetEvaluationRequest struct {
	Tag        string       `json:"tag,omitempty"`
	User       *user.User   `json:"user,omitempty"`
	FeatureID  string       `json:"featureId,omitempty"`
	SourceID   SourceIDType `json:"sourceId,omitempty"`
	SDKVersion string       `json:"sdkVersion,omitempty"`
}

func NewGetEvaluationRequest(tag, featureID string, user *user.User) *GetEvaluationRequest {
	return &GetEvaluationRequest{
		Tag:        tag,
		User:       user,
		FeatureID:  featureID,
		SourceID:   SourceIDGoServer,
		SDKVersion: version.SDKVersion,
	}
}
