package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetEvaluationSizeMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewGetEvaluationSizeMetricsEvent(tag, sizeByte)
	assert.IsType(t, &GetEvaluationSizeMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, sizeByte, e.SizeByte)
	assert.Equal(t, GetEvaluationSizeMetricsEventType, e.Type)
}
