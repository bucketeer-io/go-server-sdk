package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeoutErrorCountMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewTimeoutErrorCountMetricsEvent(tag)
	assert.IsType(t, &TimeoutErrorCountMetricsEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, TimeoutErrorCountMetricsEventType, e.Type)
}
