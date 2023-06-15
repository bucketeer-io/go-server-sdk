package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewLatencyMetricsEvent(tag, 0.1, GetEvaluation)
	assert.IsType(t, &LatencyMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, 0.1, e.LatencySecond)
	assert.Equal(t, LatencyMetricsEventType, e.Type)
}
