package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetEvaluationLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewGetEvaluationLatencyMetricsEvent(tag, value)
	assert.IsType(t, &GetEvaluationLatencyMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, value, e.Duration.Value)
	assert.Equal(t, DurationType, e.Duration.Type)
	assert.Equal(t, GetEvaluationLatencyMetricsEventType, e.Type)
}
