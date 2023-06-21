package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/model"
)

const (
	evaluationAPI    = "/get_evaluation"
	eventsAPI        = "/register_events"
	authorizationKey = "authorization"
)

func (c *client) GetEvaluation(req *model.GetEvaluationRequest) (*model.GetEvaluationResponse, int, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		evaluationAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var ger model.GetEvaluationResponse
	if err := json.Unmarshal(resp, &ger); err != nil {
		return nil, 0, err
	}
	return &ger, size, nil
}

func (c *client) RegisterEvents(req *model.RegisterEventsRequest) (*model.RegisterEventsResponse, int, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		eventsAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var rer model.RegisterEventsResponse
	if err := json.Unmarshal(resp, &rer); err != nil {
		return nil, 0, err
	}
	return &rer, size, nil
}

func (c *client) sendHTTPRequest(url string, body interface{}) ([]byte, int, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(encoded))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Add(authorizationKey, c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, NewErrStatus(resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	size, err := c.getBodySize(resp)
	if err != nil {
		return nil, 0, err
	}
	return data, size, nil
}

func (*client) getBodySize(resp *http.Response) (int, error) {
	// Given keys are case insensitive in `func (Header) Get`.
	// FYI: https://pkg.go.dev/net/http#Header.Get
	header := resp.Header.Get("Content-Length")
	if header == "" {
		header = "0"
	}
	length, err := strconv.Atoi(header)
	if err != nil {
		return 0, err
	}
	return length, nil
}
