package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/models"
)

const (
	evaluationAPI    = "/get_evaluation"
	eventsAPI        = "/register_events"
	authorizationKey = "authorization"
)

type getEvaluationRequest struct {
	*models.GetEvaluationRequest
	SourceID models.SourceIDType `json:"sourceId,omitempty"`
}

type registerEventsRequest struct {
	*models.RegisterEventsRequest
}

func (c *client) GetEvaluation(req *models.GetEvaluationRequest) (*models.GetEvaluationResponse, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		evaluationAPI,
	)
	resp, err := c.sendHTTPRequest(
		url,
		&getEvaluationRequest{
			GetEvaluationRequest: req,
			SourceID:             models.SourceIDGoServer,
		},
	)
	if err != nil {
		return nil, err
	}
	var ger models.GetEvaluationResponse
	if err := json.Unmarshal(resp, &ger); err != nil {
		return nil, err
	}
	return &ger, nil
}

func (c *client) RegisterEvents(req *models.RegisterEventsRequest) (*models.RegisterEventsResponse, error) {
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
	var rer models.RegisterEventsResponse
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
