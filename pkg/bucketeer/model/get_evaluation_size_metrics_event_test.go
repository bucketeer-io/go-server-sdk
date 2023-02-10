package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSizeMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewSizeMetricsEvent(tag, sizeByte, GetEvaluation)
	assert.IsType(t, &SizeMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, sizeByte, e.SizeByte)
	assert.Equal(t, SizeMetricsEventType, e.Type)
}
