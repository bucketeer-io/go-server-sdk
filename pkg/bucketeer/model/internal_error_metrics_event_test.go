package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInternalErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewInternalErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &InternalErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, InternalErrorMetricsEventType, e.Type)
}
