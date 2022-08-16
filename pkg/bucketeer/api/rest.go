package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
)

const (
	version          = "/v1"
	service          = "/gateway"
	pingAPI          = "/ping"
	evaluationAPI    = "/evaluation"
	eventsAPI        = "/events"
	authorizationKey = "authorization"
)

type eventType int

type metricsDetailEventType int

const (
	GetEvaluationLatencyMetricsEventType metricsDetailEventType = iota + 1
	GetEvaluationSizeMetricsEventType
	TimeoutErrorCountMetricsEventType
	InternalErrorCountMetricsEventType
)

const (
	GoalEventType eventType = iota + 1 // eventType starts from 1 for validation.
	GoalBatchEventType
	EvaluationEventType
	MetricsEventType
)

type SourceIDType int32

const (
	SourceIDGoServer SourceIDType = 5
)

type successResponse struct {
	Data json.RawMessage `json:"data"`
}

type pingResponse struct {
	Time int64 `json:"time,omitempty"`
}

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
	FeatureID string     `json:"feature_id,omitempty"`
}

type getEvaluationRequest struct {
	*GetEvaluationRequest
	SourceID SourceIDType `json:"source_id,omitempty"`
}

type GetEvaluationResponse struct {
	Evaluation *Evaluation `json:"evaluations,omitempty"`
}

type Event struct {
	ID                   string          `json:"id,omitempty"`
	Event                json.RawMessage `json:"event,omitempty"`
	EnvironmentNamespace string          `json:"environment_namespace,omitempty"`
	Type                 eventType       `json:"type,omitempty"`
}

type MetricsEvent struct {
	Timestamp int64                  `json:"timestamp,omitempty"`
	Event     json.RawMessage        `json:"event,omitempty"`
	Type      metricsDetailEventType `json:"type,omitempty"`
}

type GoalEvent struct {
	Timestamp int64        `json:"timestamp,omitempty"`
	GoalID    string       `json:"goal_id,omitempty"`
	UserID    string       `json:"user_id,omitempty"`
	Value     float64      `json:"value,omitempty"`
	User      *user.User   `json:"user,omitempty"`
	Tag       string       `json:"tag,omitempty"`
	SourceID  SourceIDType `json:"source_id,omitempty"`
}

type InternalErrorCountMetricsEvent struct {
	Tag string `json:"tag,omitempty"`
}

type GetEvaluationSizeMetricsEvent struct {
	Labels   map[string]string `json:"labels,omitempty"`
	SizeByte int32             `json:"size_byte,omitempty"`
}

type GetEvaluationLatencyMetricsEvent struct {
	Labels   map[string]string `json:"labels,omitempty"`
	Duration time.Duration     `json:"duration,omitempty"`
}

type TimeoutErrorCountMetricsEvent struct {
	Tag string `json:"tag,omitempty"`
}

type Variation struct {
	ID          string `json:"id,omitempty"`
	Value       string `json:"value,omitempty"` // number or even json object
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Evaluation struct {
	ID             string  `json:"id,omitempty"`
	FeatureID      string  `json:"feature_id,omitempty"`
	FeatureVersion int32   `json:"feature_version,omitempty"`
	UserID         string  `json:"user_id,omitempty"`
	VariationID    string  `json:"variation_id,omitempty"`
	Reason         *Reason `json:"reason,omitempty"`
	VariationValue string  `json:"variation_value,omitempty"`
}

type EvaluationEvent struct {
	Timestamp      int64      `json:"timestamp,omitempty"`
	FeatureID      string     `json:"feature_id,omitempty"`
	FeatureVersion int32      `json:"feature_version,omitempty"`
	VariationID    string     `json:"variation_id,omitempty"`
	User           *user.User `json:"user,omitempty"`
	Reason         *Reason    `json:"reason,omitempty"`
	Tag            string     `json:"tag,omitempty"`
	SourceID       int32      `json:"source_id,omitempty"`
}

type RegisterEventsResponseError struct {
	Retriable bool   `json:"retriable,omitempty"`
	Message   string `json:"message,omitempty"`
}

type ReasonType int32

const (
	ReasonClient ReasonType = 4
)

type Reason struct {
	Type   ReasonType `json:"type,omitempty"`
	RuleID string     `json:"rule_id,omitempty"`
}

type UserEvaluationsState int32

const (
	UserEvaluationsFULL UserEvaluationsState = 2
)

func (c *client) ping() (*pingResponse, error) {
	url := fmt.Sprintf("https://%s%s%s%s",
		c.host,
		version,
		service,
		pingAPI,
	)
	resp, err := c.sendHTTPRequest(url, struct{}{})
	if err != nil {
		return nil, err
	}
	var pr pingResponse
	if err := json.Unmarshal(resp.Data, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *client) GetEvaluation(req *GetEvaluationRequest) (*GetEvaluationResponse, error) {
	url := fmt.Sprintf("https://%s%s%s%s",
		c.host,
		version,
		service,
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
	if err := json.Unmarshal(resp.Data, &ger); err != nil {
		return nil, err
	}
	return &ger, nil
}

func (c *client) RegisterEvents(req *RegisterEventsRequest) (*RegisterEventsResponse, error) {
	url := fmt.Sprintf("https://%s%s%s%s",
		c.host,
		version,
		service,
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
	if err := json.Unmarshal(resp.Data, &rer); err != nil {
		return nil, err
	}
	return &rer, nil
}

func (c *client) sendHTTPRequest(url string, body interface{}) (*successResponse, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("bucketeer/api: send HTTP request failed: %d", resp.StatusCode)
	}
	var sr successResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}
