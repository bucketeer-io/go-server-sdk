package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUnknownErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewUnknownErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &UnknownErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, UnknownErrorMetricsEventType, e.Type)
}
