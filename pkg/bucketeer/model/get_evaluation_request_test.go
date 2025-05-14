package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewGetEvaluationRequest(t *testing.T) {
	t.Parallel()
	e := NewGetEvaluationRequest(tag, featureID, version.SDKVersion, SourceIDGoServer, newUser(t, id))
	assert.IsType(t, &GetEvaluationRequest{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, SourceIDGoServer, e.SourceID)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, id, e.User.ID)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
