package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

const (
	authorizationKey = "authorization"
	evaluationAPI    = "/get_evaluation"
	featureFlagsAPI  = "/get_feature_flags"
	registerEventAPI = "/register_events"
	segmentUsersAPI  = "/get_segment_users"
)

func (c *client) GetEvaluation(req *model.GetEvaluationRequest) (*model.GetEvaluationResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
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
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		registerEventAPI,
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

func (c *client) GetFeatureFlags(req *model.GetFeatureFlagsRequest) (*model.GetFeatureFlagsResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		featureFlagsAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var gfr *model.GetFeatureFlagsResponse
	if err := json.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return gfr, size, nil
}

func (c *client) GetSegmentUsers(req *model.GetSegmentUsersRequest) (*model.GetSegmentUsersResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		segmentUsersAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var gfr model.GetSegmentUsersResponse
	if err := json.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return &gfr, size, nil
}

func (c *client) sendHTTPRequest(url string, body any) ([]byte, int, error) {
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
		Timeout: 60 * time.Second,
	}
	if c.scheme == "http" {
		client.Transport = &http.Transport{
			// This setting is for developing on local machines.
			// In production, it's not recommended to specify c.scheme as "http".
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
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
	return data, int(resp.ContentLength), nil
}
