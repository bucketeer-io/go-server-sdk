package model

import (
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

func TestNewUnauthorizedErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewUnauthorizedErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &UnauthorizedErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, UnauthorizedErrorMetricsEventType, e.Type)
}

func TestNewForbiddenErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	e := NewForbiddenErrorMetricsEvent(tag, GetEvaluation)
	assert.IsType(t, &ForbiddenErrorMetricsEvent{}, e)
	assert.Equal(t, tag, e.Labels["tag"])
	assert.Equal(t, ForbiddenErrorMetricsEventType, e.Type)
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
