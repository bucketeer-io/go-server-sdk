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

// GetEvaluation retrieves evaluation for a single feature flag with retry support.
func (c *client) GetEvaluation(
	ctx context.Context,
	req *model.GetEvaluationRequest,
	deadline time.Time,
) (*model.GetEvaluationResponse, int, error) {
	url := fmt.Sprintf(
		"%s://%s%s",
		c.scheme,
		c.apiEndpoint,
		evaluationAPI,
	)

	resp, size, err := c.sendHTTPRequestWithRetry(ctx, url, req, deadline)
	if err != nil {
		return nil, 0, err
	}
	var ger model.GetEvaluationResponse
	if err := json.Unmarshal(resp, &ger); err != nil {
		return nil, 0, err
	}
	return &ger, size, nil
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
	cfg := retry.Config{
		MaxRetries:      c.retryConfig.MaxRetries,
		InitialInterval: c.retryConfig.InitialInterval,
		MaxInterval:     c.retryConfig.MaxInterval,
		Multiplier:      c.retryConfig.Multiplier,
		Deadline:        deadline,
	}

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

// GetFeatureFlags retrieves feature flags with retry support.
func (c *client) GetFeatureFlags(
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

// GetSegmentUsers retrieves segment users with retry support.
func (c *client) GetSegmentUsers(
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

	var gsur model.GetSegmentUsersResponse
	if err := json.Unmarshal(resp, &gsur); err != nil {
		return nil, 0, err
	}
	return &gsur, size, nil
}

// RegisterEvents registers events with retry support.
func (c *client) RegisterEvents(
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
