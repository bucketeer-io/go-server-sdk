package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
)

const (
	evaluationAPI    = "/get_evaluation"
	eventsAPI        = "/register_events"
	authorizationKey = "authorization"
)

type EventType string

type metricsDetailEventType string

//nolint:lll
const (
	GetEvaluationLatencyMetricsEventType metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.GetEvaluationLatencyMetricsEvent"
	GetEvaluationSizeMetricsEventType    metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.GetEvaluationSizeMetricsEvent"
	TimeoutErrorCountMetricsEventType    metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.TimeoutErrorCountMetricsEvent"
	InternalErrorCountMetricsEventType   metricsDetailEventType = "type.googleapis.com/bucketeer.event.client.InternalErrorCountMetricsEvent"
)

const (
	GoalEventType       EventType = "type.googleapis.com/bucketeer.event.client.GoalEvent"
	EvaluationEventType EventType = "type.googleapis.com/bucketeer.event.client.EvaluationEvent"
	MetricsEventType    EventType = "type.googleapis.com/bucketeer.event.client.MetricsEvent"
)

type wellKnownTypes string

const DurationType wellKnownTypes = "type.googleapis.com/google.protobuf.Duration"

type SourceIDType int32

const (
	SourceIDGoServer SourceIDType = 5
)

type RegisterEventsRequest struct {
	Events []*Event `json:"events,omitempty"`
}

type registerEventsRequest struct {
	*RegisterEventsRequest
}

type RegisterEventsResponse struct {
	Errors map[string]*RegisterEventsResponseError `json:"errors,omitempty"`
}

type GetEvaluationRequest struct {
	Tag       string     `json:"tag,omitempty"`
	User      *user.User `json:"user,omitempty"`
	FeatureID string     `json:"featureId,omitempty"`
}

type getEvaluationRequest struct {
	*GetEvaluationRequest
	SourceID SourceIDType `json:"sourceId,omitempty"`
}

type GetEvaluationResponse struct {
	Evaluation *Evaluation `json:"evaluation,omitempty"`
}

type Event struct {
	ID                   string          `json:"id,omitempty"`
	Event                json.RawMessage `json:"event,omitempty"`
	EnvironmentNamespace string          `json:"environmentNamespace,omitempty"`
}

type MetricsEvent struct {
	Timestamp int64           `json:"timestamp,omitempty"`
	Event     json.RawMessage `json:"event,omitempty"`
	Type      EventType       `json:"@type,omitempty"`
}

type GoalEvent struct {
	Timestamp int64        `json:"timestamp,omitempty"`
	GoalID    string       `json:"goalId,omitempty"`
	UserID    string       `json:"userId,omitempty"`
	Value     float64      `json:"value,omitempty"`
	User      *user.User   `json:"user,omitempty"`
	Tag       string       `json:"tag,omitempty"`
	SourceID  SourceIDType `json:"sourceId,omitempty"`
	Type      EventType    `json:"@type,omitempty"`
}

type Duration struct {
	Type  wellKnownTypes `json:"@type,omitempty"`
	Value string         `json:"value,omitempty"`
}

type InternalErrorCountMetricsEvent struct {
	Tag  string                 `json:"tag,omitempty"`
	Type metricsDetailEventType `json:"@type,omitempty"`
}

type GetEvaluationSizeMetricsEvent struct {
	Labels   map[string]string      `json:"labels,omitempty"`
	SizeByte int32                  `json:"sizeByte,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

type GetEvaluationLatencyMetricsEvent struct {
	Labels   map[string]string      `json:"labels,omitempty"`
	Duration *Duration              `json:"duration,omitempty"`
	Type     metricsDetailEventType `json:"@type,omitempty"`
}

type TimeoutErrorCountMetricsEvent struct {
	Tag  string                 `json:"tag,omitempty"`
	Type metricsDetailEventType `json:"@type,omitempty"`
}

type Variation struct {
	ID          string `json:"id,omitempty"`
	Value       string `json:"value,omitempty"` // number or even json object
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Evaluation struct {
	ID             string  `json:"id,omitempty"`
	FeatureID      string  `json:"featureId,omitempty"`
	FeatureVersion int32   `json:"featureVersion,omitempty"`
	UserID         string  `json:"userId,omitempty"`
	VariationID    string  `json:"variationId,omitempty"`
	Reason         *Reason `json:"reason,omitempty"`
	VariationValue string  `json:"variationValue,omitempty"`
}

type EvaluationEvent struct {
	Timestamp      int64        `json:"timestamp,omitempty"`
	FeatureID      string       `json:"featureId,omitempty"`
	FeatureVersion int32        `json:"featureVersion,omitempty"`
	VariationID    string       `json:"variationId,omitempty"`
	User           *user.User   `json:"user,omitempty"`
	Reason         *Reason      `json:"reason,omitempty"`
	Tag            string       `json:"tag,omitempty"`
	SourceID       SourceIDType `json:"sourceId,omitempty"`
	Type           EventType    `json:"@type,omitempty"`
}

type RegisterEventsResponseError struct {
	Retriable bool   `json:"retriable,omitempty"`
	Message   string `json:"message,omitempty"`
}

type ReasonType string

const (
	ReasonClient ReasonType = "CLIENT"
)

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"ruleId,omitempty"`
}

type UserEvaluationsState int32

const (
	UserEvaluationsFULL UserEvaluationsState = 2
)

func (c *client) GetEvaluation(req *GetEvaluationRequest) (*GetEvaluationResponse, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		evaluationAPI,
	)
	resp, err := c.sendHTTPRequest(
		url,
		&getEvaluationRequest{
			GetEvaluationRequest: req,
			SourceID:             SourceIDGoServer,
		},
	)
	if err != nil {
		return nil, err
	}
	var ger GetEvaluationResponse
	if err := json.Unmarshal(resp, &ger); err != nil {
		return nil, err
	}
	return &ger, nil
}

func (c *client) RegisterEvents(req *RegisterEventsRequest) (*RegisterEventsResponse, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		eventsAPI,
	)
	resp, err := c.sendHTTPRequest(
		url,
		&registerEventsRequest{
			RegisterEventsRequest: req,
		},
	)
	if err != nil {
		return nil, err
	}
	var rer RegisterEventsResponse
	if err := json.Unmarshal(resp, &rer); err != nil {
		return nil, err
	}
	return &rer, nil
}

func (c *client) sendHTTPRequest(url string, body interface{}) ([]byte, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(encoded))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encoded))
	if err != nil {
		return nil, err
	}
	req.Header.Add(authorizationKey, c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, NewErrStatus(resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
