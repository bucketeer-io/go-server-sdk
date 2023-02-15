package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewLatencyMetricsEvent(tag, value, GetEvaluation)
	assert.IsType(t, &LatencyMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, value, e.Duration.Value)
	assert.Equal(t, DurationType, e.Duration.Type)
	assert.Equal(t, LatencyMetricsEventType, e.Type)
}
