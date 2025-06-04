package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewEvaluationEvent(t *testing.T) {
	t.Parallel()
	e := NewEvaluationEvent(tag, featureID, variationID, version.SDKVersion, featureVersion, SourceIDGoServer, newUser(t, id), &Reason{Type: ReasonErrorException})
	assert.IsType(t, &EvaluationEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, EvaluationEventType, e.Type)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, variationID, e.VariationID)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, id, e.User.ID)
	assert.Equal(t, ReasonErrorException, e.Reason.Type)
	assert.Equal(t, e.SourceID, SourceIDGoServer)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
