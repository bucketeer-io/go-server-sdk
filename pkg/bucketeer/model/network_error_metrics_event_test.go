package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNetworkErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewNetworkErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &NetworkErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, NetworkErrorMetricsEventType, e.Type)
}
