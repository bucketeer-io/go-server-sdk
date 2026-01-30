package event

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
)

const (
	processorUserID      = "user-id"
	processorFeatureID   = "feature-id"
	processorVariationID = "variation-id"
	processorGoalID      = "goal-id"
)

func TestPushEvaluationEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	evaluation := newEvaluation(t, processorFeatureID, processorVariationID)
	p.PushEvaluationEvent(user, evaluation)
	evt := <-p.evtQueue.eventCh()
	e := model.NewEvaluationEvent(p.tag, processorFeatureID, "", version.SDKVersion, 0, model.SourceIDGoServer, user, &model.Reason{Type: model.ReasonErrorFlagNotFound})
	err := json.Unmarshal(evt.Event, e)
	assert.NoError(t, err)
	assert.Equal(t, p.tag, e.Tag)
	assert.Equal(t, model.EvaluationEventType, e.Type)
	assert.Equal(t, processorFeatureID, e.FeatureID)
	assert.Equal(t, processorVariationID, e.VariationID)
	assert.Equal(t, evaluation.FeatureVersion, e.FeatureVersion)
	assert.Equal(t, processorUserID, e.User.ID)
	assert.Equal(t, model.ReasonErrorFlagNotFound, e.Reason.Type)
	assert.Equal(t, e.SourceID, model.SourceIDGoServer)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}

func TestPushGoalEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	user := newUser(t, processorUserID)
	p.PushGoalEvent(user, processorGoalID, 1.1)
	evt := <-p.evtQueue.eventCh()
	e := model.NewGoalEvent(p.tag, processorGoalID, version.SDKVersion, 1.1, model.SourceIDGoServer, user)
	err := json.Unmarshal(evt.Event, e)
	assert.NoError(t, err)
	assert.Equal(t, p.tag, e.Tag)
	assert.Equal(t, model.GoalEventType, e.Type)
	assert.Equal(t, processorUserID, e.User.ID)
	assert.Equal(t, processorGoalID, e.GoalID)
	assert.Equal(t, 1.1, e.Value)
	assert.Equal(t, model.SourceIDGoServer, e.SourceID)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}

func TestPushLatencyMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	t1 := time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC)
	p.PushLatencyMetricsEvent(t2.Sub(t1), model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gelMetricsEvt := &model.LatencyMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, gelMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, gelMetricsEvt.APIID)
	assert.Equal(t, p.tag, gelMetricsEvt.Labels["tag"])
	assert.Equal(t, t2.Sub(t1).Seconds(), gelMetricsEvt.LatencySecond)
	assert.Equal(t, model.LatencyMetricsEventType, gelMetricsEvt.Type)
}

func TestPushSizeMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.PushSizeMetricsEvent(1, model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	gesMetricsEvt := &model.SizeMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, gesMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, gesMetricsEvt.APIID)
	assert.Equal(t, p.tag, gesMetricsEvt.Labels["tag"])
	assert.Equal(t, int32(1), gesMetricsEvt.SizeByte)
	assert.Equal(t, model.SizeMetricsEventType, gesMetricsEvt.Type)
}

func TestPushTimeoutErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushTimeoutErrorMetricsEvent(model.GetEvaluation)
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	tecMetricsEvt := &model.TimeoutErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, tecMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, tecMetricsEvt.APIID)
}

func TestPushInternalSDKErrorMetricsEvent(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushInternalSDKErrorMetricsEvent(model.GetEvaluation, errors.New("error"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iecMetricsEvt := &model.InternalSDKErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iecMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iecMetricsEvt.APIID)
	assert.Equal(t, model.InternalSDKErrorMetricsEventType, iecMetricsEvt.Type)
}

func TestUnauthorizedError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusUnauthorized, errors.New("StatusUnauthorized"))
	select {
	case evt := <-p.evtQueue.eventCh():
		// If we receive an event, the test should fail
		t.Errorf("Expected no event for unauthorized error, but got: %v", evt)
	case <-time.After(time.Millisecond * 300):
		// No event received
	}
}

func TestStatusForbiddenError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusForbidden, errors.New("StatusForbidden"))
	select {
	case evt := <-p.evtQueue.eventCh():
		// If we receive an event, the test should fail
		t.Errorf("Expected no event for forbidden error, but got: %v", evt)
	case <-time.After(time.Millisecond * 300):
		// No event received
	}
}

func TestPushErrorStatusCodeMetricsEventInternalServerError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusInternalServerError, errors.New("InternalServerError"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.InternalServerErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.InternalServerErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventMethodNotAllowed(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusMethodNotAllowed, errors.New("MethodNotAllowed"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.InternalSDKErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.InternalSDKErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventRequestTimeout(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusRequestTimeout, errors.New("StatusRequestTimeout"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.TimeoutErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.TimeoutErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventRequestEntityTooLarge(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusRequestEntityTooLarge, errors.New("StatusRequestEntityTooLarge"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.PayloadTooLargeErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.PayloadTooLargeErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventBadGateway(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusBadGateway, errors.New("StatusBadGateway"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.ServiceUnavailableErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.ServiceUnavailableErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventRedirectionRequestError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, 333, errors.New("333 error"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.RedirectionRequestErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.RedirectionRequestErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorStatusCodeMetricsEventUnknownError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, 999, errors.New("999 error"))
	evt := <-p.evtQueue.eventCh()
	metricsEvt := &model.MetricsEvent{}
	err := json.Unmarshal(evt.Event, metricsEvt)
	assert.NoError(t, err)
	iseMetricsEvt := &model.UnknownErrorMetricsEvent{}
	err = json.Unmarshal(metricsEvt.Event, iseMetricsEvt)
	assert.NoError(t, err)
	assert.Equal(t, model.GetEvaluation, iseMetricsEvt.APIID)
	assert.Equal(t, model.UnknownErrorMetricsEventType, iseMetricsEvt.Type)
}

func TestPushErrorEventWhenNetworkError(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc string
		err  error
	}{
		{
			desc: "connection refused (net.OpError)",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: syscall.ECONNREFUSED,
			},
		},
		{
			desc: "host unreachable (net.OpError)",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: syscall.EHOSTUNREACH,
			},
		},
		{
			desc: "url.Error wrapping net.OpError",
			err: &url.Error{
				Op:  "Get",
				URL: "http://localhost:9999",
				Err: &net.OpError{
					Op:  "dial",
					Net: "tcp",
					Err: syscall.ECONNREFUSED,
				},
			},
		},
		{
			desc: "DNS error",
			err: &net.DNSError{
				Err:        "no such host",
				Name:       "example.invalid",
				IsNotFound: true,
			},
		},
		{
			desc: "context.Canceled",
			err:  context.Canceled,
		},
		{
			desc: "wrapped syscall ECONNRESET",
			err:  fmt.Errorf("connection error: %w", syscall.ECONNRESET),
		},
		{
			desc: "wrapped syscall ENETUNREACH",
			err:  fmt.Errorf("network error: %w", syscall.ENETUNREACH),
		},
	}
	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, 10)
			p.PushErrorEvent(pt.err, model.RegisterEvents)
			evt := <-p.evtQueue.eventCh()
			metricsEvt := &model.MetricsEvent{}
			err := json.Unmarshal(evt.Event, metricsEvt)
			assert.NoError(t, err)
			actual := &model.NetworkErrorMetricsEvent{}
			err = json.Unmarshal(metricsEvt.Event, actual)
			assert.NoError(t, err)
			assert.Equal(t, model.RegisterEvents, actual.APIID)
			assert.Equal(t, p.tag, actual.Labels["tag"])
			assert.Equal(t, model.NetworkErrorMetricsEventType, actual.Type)
		})
	}
}

func TestPushErrorEventWhenInternalSDKError(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc string
		err  error
	}{
		{
			desc: "Internal error",
			err:  errors.New("Internal error"),
		},
	}
	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, 10)
			p.PushErrorEvent(pt.err, model.RegisterEvents)
			evt := <-p.evtQueue.eventCh()
			metricsEvt := &model.MetricsEvent{}
			err := json.Unmarshal(evt.Event, metricsEvt)
			assert.NoError(t, err)
			actual := &model.InternalSDKErrorMetricsEvent{}
			err = json.Unmarshal(metricsEvt.Event, actual)
			assert.NoError(t, err)
			assert.Equal(t, model.RegisterEvents, actual.APIID)
			assert.Equal(t, p.tag, actual.Labels["tag"])
			assert.Equal(t, model.InternalSDKErrorMetricsEventType, actual.Type)
		})
	}
}

// mockTimeoutError implements net.Error with Timeout() returning true
type mockTimeoutError struct {
	msg string
}

func (e *mockTimeoutError) Error() string   { return e.msg }
func (e *mockTimeoutError) Timeout() bool   { return true }
func (e *mockTimeoutError) Temporary() bool { return true }

func TestPushErrorEventWhenTimeoutErr(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc string
		err  error
	}{
		{
			desc: "StatusGatewayTimeout",
			err:  api.NewErrStatus(http.StatusGatewayTimeout),
		},
		{
			desc: "context.DeadlineExceeded",
			err:  context.DeadlineExceeded,
		},
		{
			desc: "wrapped context.DeadlineExceeded",
			err:  fmt.Errorf("operation failed: %w", context.DeadlineExceeded),
		},
		{
			desc: "net.Error with Timeout() true",
			err:  &mockTimeoutError{msg: "connection timed out"},
		},
		{
			desc: "url.Error with Timeout",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: &mockTimeoutError{msg: "i/o timeout"},
			},
		},
		{
			desc: "DNS timeout error",
			err: &net.DNSError{
				Err:       "lookup timed out",
				Name:      "example.com",
				IsTimeout: true,
			},
		},
		{
			desc: "net.OpError with timeout",
			err: &net.OpError{
				Op:  "read",
				Net: "tcp",
				Err: &mockTimeoutError{msg: "i/o timeout"},
			},
		},
	}
	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, 10)
			p.PushErrorEvent(pt.err, model.RegisterEvents)
			evt := <-p.evtQueue.eventCh()
			metricsEvt := &model.MetricsEvent{}
			err := json.Unmarshal(evt.Event, metricsEvt)
			assert.NoError(t, err)
			actual := &model.TimeoutErrorMetricsEvent{}
			err = json.Unmarshal(metricsEvt.Event, actual)
			assert.NoError(t, err)
			assert.Equal(t, model.RegisterEvents, actual.APIID)
			assert.Equal(t, p.tag, actual.Labels["tag"])
			assert.Equal(t, model.TimeoutErrorMetricsEventType, actual.Type)
		})
	}
}

func TestPushErrorEventWhenOtherStatus(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc string
		err  error
	}{
		{
			desc: "InternalServerError",
			err:  api.NewErrStatus(http.StatusInternalServerError),
		},
	}
	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, 10)
			p.PushErrorEvent(pt.err, model.RegisterEvents)
			evt := <-p.evtQueue.eventCh()
			metricsEvt := &model.MetricsEvent{}
			err := json.Unmarshal(evt.Event, metricsEvt)
			assert.NoError(t, err)
			actual := &model.InternalServerErrorMetricsEvent{}
			err = json.Unmarshal(metricsEvt.Event, actual)
			assert.NoError(t, err)
			assert.Equal(t, model.RegisterEvents, actual.APIID)
			assert.Equal(t, p.tag, actual.Labels["tag"])
			assert.Equal(t, model.InternalServerErrorMetricsEventType, actual.Type)
		})
	}
}

func TestClassifyError(t *testing.T) {
	t.Parallel()
	patterns := []struct {
		desc           string
		err            error
		expectedType   errorType
		expectedStatus int
	}{
		// Nil error
		{
			desc:           "nil error returns unknown",
			err:            nil,
			expectedType:   errorTypeUnknown,
			expectedStatus: 0,
		},
		// HTTP status code errors
		{
			desc:           "HTTP 500 returns statusCode",
			err:            api.NewErrStatus(http.StatusInternalServerError),
			expectedType:   errorTypeStatusCode,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			desc:           "HTTP 502 returns statusCode",
			err:            api.NewErrStatus(http.StatusBadGateway),
			expectedType:   errorTypeStatusCode,
			expectedStatus: http.StatusBadGateway,
		},
		{
			desc:           "HTTP 503 returns statusCode",
			err:            api.NewErrStatus(http.StatusServiceUnavailable),
			expectedType:   errorTypeStatusCode,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			desc:           "HTTP 504 returns statusCode",
			err:            api.NewErrStatus(http.StatusGatewayTimeout),
			expectedType:   errorTypeStatusCode,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			desc:           "HTTP 429 returns statusCode",
			err:            api.NewErrStatus(http.StatusTooManyRequests),
			expectedType:   errorTypeStatusCode,
			expectedStatus: http.StatusTooManyRequests,
		},
		// Context errors
		{
			desc:           "context.DeadlineExceeded returns timeout",
			err:            context.DeadlineExceeded,
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped context.DeadlineExceeded returns timeout",
			err:            fmt.Errorf("operation failed: %w", context.DeadlineExceeded),
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		{
			desc:           "context.Canceled returns network",
			err:            context.Canceled,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped context.Canceled returns network",
			err:            fmt.Errorf("request canceled: %w", context.Canceled),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		// net.Error with Timeout
		{
			desc:           "net.Error with Timeout() true returns timeout",
			err:            &mockTimeoutError{msg: "connection timed out"},
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		// net.OpError
		{
			desc: "net.OpError returns network",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: syscall.ECONNREFUSED,
			},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc: "net.OpError with timeout returns timeout",
			err: &net.OpError{
				Op:  "read",
				Net: "tcp",
				Err: &mockTimeoutError{msg: "i/o timeout"},
			},
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		// url.Error
		{
			desc: "url.Error returns network",
			err: &url.Error{
				Op:  "Get",
				URL: "http://localhost:9999",
				Err: &net.OpError{
					Op:  "dial",
					Net: "tcp",
					Err: syscall.ECONNREFUSED,
				},
			},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc: "url.Error with timeout returns timeout",
			err: &url.Error{
				Op:  "Get",
				URL: "http://example.com",
				Err: &mockTimeoutError{msg: "i/o timeout"},
			},
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		// DNS errors
		{
			desc: "net.DNSError returns network",
			err: &net.DNSError{
				Err:        "no such host",
				Name:       "example.invalid",
				IsNotFound: true,
			},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc: "net.DNSError with IsTimeout returns timeout",
			err: &net.DNSError{
				Err:       "lookup timed out",
				Name:      "example.com",
				IsTimeout: true,
			},
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		// Syscall errors
		{
			desc:           "syscall.ECONNREFUSED returns network",
			err:            syscall.ECONNREFUSED,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ECONNRESET returns network",
			err:            syscall.ECONNRESET,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ECONNABORTED returns network",
			err:            syscall.ECONNABORTED,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.EHOSTUNREACH returns network",
			err:            syscall.EHOSTUNREACH,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ENETUNREACH returns network",
			err:            syscall.ENETUNREACH,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ETIMEDOUT returns timeout",
			err:            syscall.ETIMEDOUT,
			expectedType:   errorTypeTimeout,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped syscall.ECONNREFUSED returns network",
			err:            fmt.Errorf("connection failed: %w", syscall.ECONNREFUSED),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		// Additional syscall errors (new)
		{
			desc:           "syscall.EPIPE returns network",
			err:            syscall.EPIPE,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ENETDOWN returns network",
			err:            syscall.ENETDOWN,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "syscall.ENETRESET returns network",
			err:            syscall.ENETRESET,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped syscall.EPIPE returns network",
			err:            fmt.Errorf("write failed: %w", syscall.EPIPE),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		// EOF errors (new)
		{
			desc:           "io.EOF returns network",
			err:            io.EOF,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "io.ErrUnexpectedEOF returns network",
			err:            io.ErrUnexpectedEOF,
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped io.EOF returns network",
			err:            fmt.Errorf("read failed: %w", io.EOF),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		// TLS/certificate errors (new)
		{
			desc:           "x509.CertificateInvalidError returns network",
			err:            x509.CertificateInvalidError{Reason: x509.Expired},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "x509.UnknownAuthorityError returns network",
			err:            x509.UnknownAuthorityError{},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "x509.HostnameError returns network",
			err:            x509.HostnameError{Host: "example.com"},
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped x509 error returns network",
			err:            fmt.Errorf("handshake failed: %w", x509.CertificateInvalidError{Reason: x509.Expired}),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "TLS error string returns network",
			err:            errors.New("tls: bad certificate"),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		{
			desc:           "x509 error string returns network",
			err:            errors.New("x509: certificate has expired"),
			expectedType:   errorTypeNetwork,
			expectedStatus: 0,
		},
		// Unknown errors
		{
			desc:           "generic error returns internal",
			err:            errors.New("some internal error"),
			expectedType:   errorTypeInternal,
			expectedStatus: 0,
		},
		{
			desc:           "wrapped generic error returns internal",
			err:            fmt.Errorf("wrapped: %w", errors.New("some internal error")),
			expectedType:   errorTypeInternal,
			expectedStatus: 0,
		},
	}

	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			errType, statusCode := classifyError(pt.err)
			assert.Equal(t, pt.expectedType, errType, "error type mismatch")
			assert.Equal(t, pt.expectedStatus, statusCode, "status code mismatch")
		})
	}
}

func TestPushEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc               string
		eventQueueCapacity int
		encodedEvt         []byte
		isErr              bool
	}{
		{
			desc:               "return error when failed to push event",
			eventQueueCapacity: 0,
			encodedEvt:         newEncodedEvaluationEvent(t, processorFeatureID),
			isErr:              true,
		},
		{
			desc:               "success",
			eventQueueCapacity: 10,
			encodedEvt:         newEncodedEvaluationEvent(t, processorFeatureID),
			isErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestPushEvent(t, tt.eventQueueCapacity)
			err := p.PushEvent(tt.encodedEvt)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				evt := <-p.evtQueue.eventCh()
				evalationEvt := &model.EvaluationEvent{}
				err := json.Unmarshal(evt.Event, evalationEvt)
				assert.NoError(t, err)
			}
		})
	}
}

func newProcessorForTestPushEvent(t *testing.T, eventQueueCapacity int) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: eventQueueCapacity,
		}),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
	}
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return &user.User{ID: id}
}

func newEvaluation(t *testing.T, featureID, variationID string) *model.Evaluation {
	t.Helper()
	return &model.Evaluation{
		FeatureID:      featureID,
		FeatureVersion: 2,
		VariationID:    variationID,
		Reason:         &model.Reason{Type: model.ReasonErrorFlagNotFound},
	}
}

func newEncodedEvaluationEvent(t *testing.T, featureID string) []byte {
	t.Helper()
	evaluationEvt := &model.EvaluationEvent{
		FeatureID: featureID,
		SourceID:  model.SourceIDGoServer,
	}
	encodedEvt, err := json.Marshal(evaluationEvt)
	assert.NoError(t, err)
	return encodedEvt
}

func TestFlushEvents(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := context.TODO()
	tests := []struct {
		desc             string
		setup            func(*processor, []*model.Event)
		events           []*model.Event
		expectedQueueLen int
	}{
		{
			desc: "do nothing when events length is 0",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			events:           make([]*model.Event, 0, 10),
			expectedQueueLen: 0,
		},
		{
			desc: "re-push all events when failed to register events",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					0,
					api.NewErrStatus(http.StatusInternalServerError),
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 4,
		},
		{
			desc: "faled to re-push all events when failed to register events if queue is closed",
			setup: func(p *processor, events []*model.Event) {
				p.evtQueue.close()
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					0,
					api.NewErrStatus(http.StatusInternalServerError),
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "re-push events when register events res contains retriable errors",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&model.RegisterEventsResponse{
						Errors: map[string]*model.RegisterEventsResponseError{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					0,
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 1,
		},
		{
			desc: "faled to re-push events when register events res contains retriable errors if queue is closed",
			setup: func(p *processor, events []*model.Event) {
				p.evtQueue.close()
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&model.RegisterEventsResponse{
						Errors: map[string]*model.RegisterEventsResponseError{
							"id-0": {Retriable: true, Message: "retriable"},
							"id-1": {Retriable: false, Message: "non retriable"},
						},
					},
					0,
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
		{
			desc: "success",
			setup: func(p *processor, events []*model.Event) {
				p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&model.RegisterEventsResponse{
						Errors: make(map[string]*model.RegisterEventsResponseError),
					},
					0,
					nil,
				)
			},
			events:           []*model.Event{{ID: "id-0"}, {ID: "id-1"}, {ID: "id-2"}},
			expectedQueueLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestWorker(t, mockCtrl)
			if tt.setup != nil {
				tt.setup(p, tt.events)
			}
			p.flushEvents(ctx, tt.events)
			assert.Len(t, p.evtQueue.eventCh(), tt.expectedQueueLen)
		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := context.TODO()
	tests := []struct {
		desc    string
		setup   func(*processor)
		timeout time.Duration
		isErr   bool
	}{
		{
			desc:    "return error when ctx is canceled",
			setup:   nil,
			timeout: 1 * time.Millisecond,
			isErr:   true,
		},
		{
			desc: "success",
			setup: func(p *processor) {
				go p.startWorkers(ctx)
			},
			timeout: 1 * time.Minute,
			isErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := newProcessorForTestWorker(t, mockCtrl)
			if tt.setup != nil {
				tt.setup(p)
			}
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()
			err := p.Close(ctx)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newProcessorForTestWorker(t *testing.T, mockCtrl *gomock.Controller) *processor {
	t.Helper()
	return &processor{
		evtQueue: newQueue(&queueConfig{
			capacity: 10,
		}),
		numFlushWorkers: 3,
		flushInterval:   1 * time.Minute,
		flushSize:       10,
		flushTimeout:    10 * time.Second,
		apiClient:       mockapi.NewMockClient(mockCtrl),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
		closeCh:  make(chan struct{}),
		sourceID: model.SourceIDGoServer,
	}
}
