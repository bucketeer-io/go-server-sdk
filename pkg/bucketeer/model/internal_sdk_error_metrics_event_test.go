package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInternalSDKErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewInternalSDKErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &InternalSDKErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, InternalSDKErrorMetricsEventType, e.Type)
}
