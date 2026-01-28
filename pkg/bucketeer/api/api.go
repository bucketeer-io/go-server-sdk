package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/retry"
)

const (
	authorizationKey = "authorization"
	evaluationAPI    = "/get_evaluation"
	featureFlagsAPI  = "/get_feature_flags"
	registerEventAPI = "/register_events"
	segmentUsersAPI  = "/get_segment_users"
)

// GetEvaluation is used for server-side evaluation mode (no WithDeadline variant needed).
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

// TODO: Replace usages with RegisterEventsWithDeadline for deadline-aware retry.
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

// TODO: Replace usages with GetFeatureFlagsWithDeadline for deadline-aware retry.
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

// TODO: Replace usages with GetSegmentUsersWithDeadline for deadline-aware retry.
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
	return c.sendHTTPRequestWithContext(context.Background(), url, body)
}

func (c *client) sendHTTPRequestWithContext(ctx context.Context, url string, body any) ([]byte, int, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(encoded))
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

	// Read response body regardless of status code (needed for retry-after header)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, 0, newErrStatusWithRetryAfter(resp.StatusCode, retryAfter)
	}

	return data, int(resp.ContentLength), nil
}

// sendHTTPRequestWithRetry wraps sendHTTPRequestWithContext with retry logic.
func (c *client) sendHTTPRequestWithRetry(
	ctx context.Context,
	url string,
	body any,
	deadline time.Time,
) ([]byte, int, error) {
	cfg := c.retryConfig
	cfg.Deadline = deadline

	var respData []byte
	var respSize int

	err := retry.Do(ctx, cfg, func() error {
		data, size, err := c.sendHTTPRequestWithContext(ctx, url, body)
		if err != nil {
			if isRetryableError(err) {
				retryAfter := getRetryAfter(err)
				return &retry.RetryableError{Err: err, Retryable: true, RetryAfter: retryAfter}
			}
			return err
		}

		respData = data
		respSize = size
		return nil
	})

	return respData, respSize, err
}

// GetFeatureFlagsWithDeadline retrieves feature flags with retry support and deadline.
func (c *client) GetFeatureFlagsWithDeadline(
	ctx context.Context,
	req *model.GetFeatureFlagsRequest,
	deadline time.Time,
) (*model.GetFeatureFlagsResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		featureFlagsAPI,
	)

	resp, size, err := c.sendHTTPRequestWithRetry(ctx, url, req, deadline)
	if err != nil {
		return nil, 0, err
	}

	var gfr *model.GetFeatureFlagsResponse
	if err := json.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return gfr, size, nil
}

// GetSegmentUsersWithDeadline retrieves segment users with retry support and deadline.
func (c *client) GetSegmentUsersWithDeadline(
	ctx context.Context,
	req *model.GetSegmentUsersRequest,
	deadline time.Time,
) (*model.GetSegmentUsersResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		segmentUsersAPI,
	)

	resp, size, err := c.sendHTTPRequestWithRetry(ctx, url, req, deadline)
	if err != nil {
		return nil, 0, err
	}

	var gfr model.GetSegmentUsersResponse
	if err := json.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return &gfr, size, nil
}

// RegisterEventsWithDeadline registers events with retry support and deadline.
func (c *client) RegisterEventsWithDeadline(
	ctx context.Context,
	req *model.RegisterEventsRequest,
	deadline time.Time,
) (*model.RegisterEventsResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		registerEventAPI,
	)

	resp, size, err := c.sendHTTPRequestWithRetry(ctx, url, req, deadline)
	if err != nil {
		return nil, 0, err
	}

	var rer model.RegisterEventsResponse
	if err := json.Unmarshal(resp, &rer); err != nil {
		return nil, 0, err
	}
	return &rer, size, nil
}
