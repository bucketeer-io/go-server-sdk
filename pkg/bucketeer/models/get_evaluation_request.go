package models

import "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"

type GetEvaluationRequest struct {
	Tag       string     `json:"tag,omitempty"`
	User      *user.User `json:"user,omitempty"`
	FeatureID string     `json:"featureId,omitempty"`
}
