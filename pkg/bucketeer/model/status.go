package model

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
)

// HTTP Mapping: 400 Bad Request
type BadRequestErrorMetricsEvent struct {
	APIID  APIID                `json:"api_id,omitempty"`
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

// HTTP Mapping: 401 Unauthorized
type UnauthorizedErrorMetricsEvent struct {
	APIID  APIID                `json:"api_id,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewUnauthorizedErrorMetricsEvent(tag string, api APIID) *UnauthorizedErrorMetricsEvent {
	return &UnauthorizedErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   UnauthorizedErrorMetricsEventType,
	}
}

// HTTP Mapping: 403 Forbidden
type ForbiddenErrorMetricsEvent struct {
	APIID  APIID                `json:"api_id,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Type   errorStatusEventType `json:"@type,omitempty"`
}

func NewForbiddenErrorMetricsEvent(tag string, api APIID) *ForbiddenErrorMetricsEvent {
	return &ForbiddenErrorMetricsEvent{
		APIID:  api,
		Labels: map[string]string{"tag": tag},
		Type:   ForbiddenErrorMetricsEventType,
	}
}

// HTTP Mapping: 404 Not Found
type NotFoundErrorMetricsEvent struct {
	APIID  APIID                `json:"api_id,omitempty"`
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
	APIID  APIID                `json:"api_id,omitempty"`
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
	APIID  APIID                `json:"api_id,omitempty"`
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
	APIID  APIID                `json:"api_id,omitempty"`
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
