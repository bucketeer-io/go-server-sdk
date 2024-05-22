package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
	"google.golang.org/protobuf/encoding/protojson"

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

// We convert the response to the proto message because it uses less memory in the cache,
// and the evaluation module uses proto messages.
// We are considering to use gRPC again to avoid converting.
func (c *client) GetFeatureFlags(req *model.GetFeatureFlagsRequest) (*gwproto.GetFeatureFlagsResponse, int, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		featureFlagsAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var gfr gwproto.GetFeatureFlagsResponse
	if err := protojson.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return &gfr, size, nil
}

// We convert the response to the proto message because it uses less memory in the cache,
// and the evaluation module uses proto messages.
// We are considering to use gRPC again to avoid converting.
func (c *client) GetSegmentUsers(req *model.GetSegmentUsersRequest) (*gwproto.GetSegmentUsersResponse, int, error) {
	url := fmt.Sprintf("https://%s%s",
		c.host,
		segmentUsersAPI,
	)
	resp, size, err := c.sendHTTPRequest(
		url,
		req,
	)
	if err != nil {
		return nil, 0, err
	}
	var gfr gwproto.GetSegmentUsersResponse
	if err := protojson.Unmarshal(resp, &gfr); err != nil {
		return nil, 0, err
	}
	return &gfr, size, nil
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
		Timeout: 60 * time.Second,
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
