package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBadRequestErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewBadRequestErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &BadRequestErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, BadRequestErrorMetricsEventType, e.Type)
}

func TestNewNotFoundErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewNotFoundErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &NotFoundErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, NotFoundErrorMetricsEventType, e.Type)
}

func TestNewClientClosedRequestErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewClientClosedRequestErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &ClientClosedRequestErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, ClientClosedRequestErrorMetricsEventType, e.Type)
}

func TestNewInternalServerErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewInternalServerErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &InternalServerErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, InternalServerErrorMetricsEventType, e.Type)
}

func TestNewServiceUnavailableErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewServiceUnavailableErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &ServiceUnavailableErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, ServiceUnavailableErrorMetricsEventType, e.Type)
}

func TestNewPayloadTooLargeErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewPayloadTooLargeErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &PayloadTooLargeErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, PayloadTooLargeErrorMetricsEventType, e.Type)
}

func TestNewRedirectionRequestErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewRedirectionRequestErrorMetricsEvent(tag, GetEvaluation, errorStatus)
	assert.IsType(t, &RedirectionRequestErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, fmt.Sprint(errorStatus), e.Labels["response_code"])
	assert.Equal(t, RedirectionRequestErrorMetricsEventType, e.Type)
}
