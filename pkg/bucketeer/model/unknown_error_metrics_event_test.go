package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUnknownErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewUnknownErrorMetricsEvent(tag, errorStatus, errorMessage, GetEvaluation)
	assert.IsType(t, &UnknownErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, errorMessage, e.Labels["error_message"])
	assert.Equal(t, fmt.Sprint(errorStatus), e.Labels["response_code"])
	assert.Equal(t, UnknownErrorMetricsEventType, e.Type)
}
