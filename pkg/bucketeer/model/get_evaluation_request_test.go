package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetEvaluationRequest(t *testing.T) {
	t.Parallel()
	e := NewGetEvaluationRequest(tag, featureID, newUser(t, id))
	assert.IsType(t, &GetEvaluationRequest{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, SourceIDGoServer, e.SourceID)
	assert.Equal(t, featureID, e.FeatureID)
	assert.Equal(t, id, e.User.ID)
}
