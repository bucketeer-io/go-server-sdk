package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInternalErrorCountMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewInternalErrorCountMetricsEvent(tag)
	assert.IsType(t, &InternalErrorCountMetricsEvent{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, InternalErrorCountMetricsEventType, e.Type)
}
