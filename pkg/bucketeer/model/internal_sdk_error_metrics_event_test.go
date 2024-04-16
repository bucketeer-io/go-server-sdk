package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInternalSDKErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewInternalSDKErrorMetricsEvent(tag, GetEvaluation, errorMessage)
	assert.IsType(t, &InternalSDKErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, errorMessage, e.Labels["error_message"])
	assert.Equal(t, InternalSDKErrorMetricsEventType, e.Type)
}
