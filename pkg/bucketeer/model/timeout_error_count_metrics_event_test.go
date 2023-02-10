package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeoutErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewTimeoutErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &TimeoutErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, TimeoutErrorMetricsEventType, e.Type)
}
