package model

import "strconv"

type errorStatusEventType string

//nolint:lll
const (
	BadRequestErrorMetricsEventType          errorStatusEventType = "type.googleapis.com/bucketeer.event.client.BadRequestErrorMetricsEvent"
	UnauthorizedErrorMetricsEventType        errorStatusEventType = "type.googleapis.com/bucketeer.event.client.UnauthorizedErrorMetricsEvent"
	ForbiddenErrorMetricsEventType           errorStatusEventType = "type.googleapis.com/bucketeer.event.client.ForbiddenErrorMetricsEvent"
	NotFoundErrorMetricsEventType            errorStatusEventType = "type.googleapis.com/bucketeer.event.client.NotFoundErrorMetricsEvent"
	ClientClosedRequestErrorMetricsEventType errorStatusEventType = "type.googleapis.com/bucketeer.event.client.ClientClosedRequestErrorMetricsEvent"
	InternalServerErrorMetricsEventType      errorStatusEventType = "type.googleapis.com/bucketeer.event.client.InternalServerErrorMetricsEvent"
	ServiceUnavailableErrorMetricsEventType  errorStatusEventType = "type.googleapis.com/bucketeer.event.client.ServiceUnavailableErrorMetricsEvent"
	PayloadTooLargeErrorMetricsEventType     errorStatusEventType = "type.googleapis.com/bucketeer.event.client.PayloadTooLargeExceptionEvent"
	RedirectionRequestErrorMetricsEventType  errorStatusEventType = "type.googleapis.com/bucketeer.event.client.RedirectionRequestExceptionEvent"
)

// HTTP Mapping: 400 Bad Request
type BadRequestErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewBadRequestErrorMetricsEvent(tag string, api APIID) *BadRequestErrorMetricsEvent {
	return &BadRequestErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   BadRequestErrorMetricsEventType,
	}
}

// HTTP Mapping: 404 Not Found
type NotFoundErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewNotFoundErrorMetricsEvent(tag string, api APIID) *NotFoundErrorMetricsEvent {
	return &NotFoundErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   NotFoundErrorMetricsEventType,
	}
}

// HTTP Mapping: 499 Client Closed Request
type ClientClosedRequestErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewClientClosedRequestErrorMetricsEvent(tag string, api APIID) *ClientClosedRequestErrorMetricsEvent {
	return &ClientClosedRequestErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   ClientClosedRequestErrorMetricsEventType,
	}
}

// HTTP Mapping: 500 Internal Server Error
type InternalServerErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewInternalServerErrorMetricsEvent(tag string, api APIID) *InternalServerErrorMetricsEvent {
	return &InternalServerErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   InternalServerErrorMetricsEventType,
	}
}

// HTTP Mapping: 503 Service Unavailable
type ServiceUnavailableErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewServiceUnavailableErrorMetricsEvent(tag string, api APIID) *ServiceUnavailableErrorMetricsEvent {
	return &ServiceUnavailableErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   ServiceUnavailableErrorMetricsEventType,
	}
}

type PayloadTooLargeErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewPayloadTooLargeErrorMetricsEvent(tag string, api APIID) *PayloadTooLargeErrorMetricsEvent {
	return &PayloadTooLargeErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   PayloadTooLargeErrorMetricsEventType,
	}
}

type RedirectionRequestErrorMetricsEvent struct {
	APIID  APIID                `json:"apiId,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewRedirectionRequestErrorMetricsEvent(tag string, api APIID, code int) *RedirectionRequestErrorMetricsEvent {
	return &RedirectionRequestErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag, "response_code": strconv.Itoa(code)},
		Type:   RedirectionRequestErrorMetricsEventType,
	}
}
