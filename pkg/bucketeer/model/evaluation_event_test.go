package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/version"
)

func TestNewEvaluationEvent(t *testing.T) {
	t.Parallel()
	e := NewEvaluationEvent(tag, featureID, variationID, featureVersion, newUser(t, id), &Reason{Type: ReasonClient})
	assert.IsType(t, &EvaluationEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, EvaluationEventType, e.Type)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, variationID, e.VariationID)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, id, e.User.ID)
	assert.Equal(t, ReasonClient, e.Reason.Type)
	assert.Equal(t, e.SourceID, SourceIDGoServer)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
