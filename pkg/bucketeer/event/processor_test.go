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
	"sync"
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	// Unauthorized errors should not push events to the queue
	time.Sleep(time.Millisecond * 50) // Give time for any async processing
	evt, ok := p.evtQueue.pop()
	assert.False(t, ok, "Expected no event for unauthorized error")
	assert.Nil(t, evt)
}

func TestStatusForbiddenError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusForbidden, errors.New("StatusForbidden"))
	// Forbidden errors should not push events to the queue
	time.Sleep(time.Millisecond * 50) // Give time for any async processing
	evt, ok := p.evtQueue.pop()
	assert.False(t, ok, "Expected no event for forbidden error")
	assert.Nil(t, evt)
}

func TestPushErrorStatusCodeMetricsEventInternalServerError(t *testing.T) {
	t.Parallel()
	p := newProcessorForTestPushEvent(t, 10)
	p.pushErrorStatusCodeMetricsEvent(model.GetEvaluation, http.StatusInternalServerError, errors.New("InternalServerError"))
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
	evt, ok := p.evtQueue.pop()
	assert.True(t, ok)
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
			evt, ok := p.evtQueue.pop()
			assert.True(t, ok)
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
			evt, ok := p.evtQueue.pop()
			assert.True(t, ok)
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
			evt, ok := p.evtQueue.pop()
			assert.True(t, ok)
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
			evt, ok := p.evtQueue.pop()
			assert.True(t, ok)
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
	t.Run("return error when queue is full", func(t *testing.T) {
		// Create processor with minimum capacity (1)
		p := newProcessorForTestPushEvent(t, 1)
		encodedEvt := newEncodedEvaluationEvent(t, processorFeatureID)
		// First push should succeed
		err := p.PushEvent(encodedEvt)
		assert.NoError(t, err)
		// Second push should fail because queue is full
		err = p.PushEvent(encodedEvt)
		assert.Error(t, err)
	})
	t.Run("success", func(t *testing.T) {
		p := newProcessorForTestPushEvent(t, 10)
		encodedEvt := newEncodedEvaluationEvent(t, processorFeatureID)
		err := p.PushEvent(encodedEvt)
		assert.NoError(t, err)
		evt, ok := p.evtQueue.pop()
		assert.True(t, ok)
		evalationEvt := &model.EvaluationEvent{}
		err = json.Unmarshal(evt.Event, evalationEvt)
		assert.NoError(t, err)
	})
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
			desc: "failed to re-push all events when failed to register events if draining",
			setup: func(p *processor, events []*model.Event) {
				p.isDraining = true
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
			desc: "failed to re-push events when register events res contains retriable errors if draining",
			setup: func(p *processor, events []*model.Event) {
				p.isDraining = true
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
			p.flushEvents(tt.events)
			assert.Equal(t, tt.expectedQueueLen, p.evtQueue.len())
		})
	}
}

func TestDrain(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("success when drain completes", func(t *testing.T) {
		p := newProcessorForTestWorker(t, mockCtrl)
		// Start the dispatcher
		go p.runDispatcher()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := p.Drain(ctx)
		assert.NoError(t, err)
	})
}

func newProcessorForTestWorker(t *testing.T, mockCtrl *gomock.Controller) *processor {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
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
		ctx:       ctx,
		cancel:    cancel,
		drainCh:   make(chan struct{}),
		done:      make(chan struct{}),
		workerSem: make(chan struct{}, 3),
		sourceID:  model.SourceIDGoServer,
	}
}

func TestSubmitBatch(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("empty batch does nothing", func(t *testing.T) {
		p := newProcessorForTestWorker(t, mockCtrl)
		// Should not panic or do anything
		p.submitBatch(nil)
		p.submitBatch([]*model.Event{})
	})

	t.Run("spawns worker goroutine when semaphore available", func(t *testing.T) {
		p := newProcessorForTestWorker(t, mockCtrl)
		p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&model.RegisterEventsResponse{Errors: make(map[string]*model.RegisterEventsResponseError)},
			0,
			nil,
		)

		batch := []*model.Event{{ID: "1"}, {ID: "2"}}
		ok := p.trySubmitBatch(batch)
		assert.True(t, ok)

		// Wait for worker to complete
		p.waitForWorkers()
	})

	t.Run("blocks until worker available", func(t *testing.T) {
		p := newProcessorForTestWorker(t, mockCtrl)

		// Fill up semaphore (simulate all workers busy)
		for i := 0; i < cap(p.workerSem); i++ {
			p.workerSem <- struct{}{}
		}

		// Expect worker goroutine to flush
		p.apiClient.(*mockapi.MockClient).EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&model.RegisterEventsResponse{Errors: make(map[string]*model.RegisterEventsResponseError)},
			0,
			nil,
		)

		batch := []*model.Event{{ID: "1"}}
		done := make(chan struct{})
		go func() {
			p.trySubmitBatch(batch) // Should block until worker available
			close(done)
		}()

		// Should be blocked
		select {
		case <-done:
			t.Fatal("trySubmitBatch should block when all workers busy")
		case <-time.After(50 * time.Millisecond):
			// Expected - still waiting for worker
		}

		// Release one permit (simulate worker finishing)
		<-p.workerSem

		// Now trySubmitBatch should unblock and spawn worker
		select {
		case <-done:
			// Success - trySubmitBatch unblocked
		case <-time.After(100 * time.Millisecond):
			t.Fatal("trySubmitBatch should unblock when worker becomes available")
		}

		// Wait a bit for spawned worker goroutine to complete and release its permit
		time.Sleep(50 * time.Millisecond)

		// Release remaining 2 permits we held at start
		for i := 0; i < cap(p.workerSem)-1; i++ {
			<-p.workerSem
		}
	})

	t.Run("returns false on hard shutdown during drain", func(t *testing.T) {
		// Create processor with drain channel
		ctx, cancel := context.WithCancel(context.Background())
		drainCh := make(chan struct{})
		p := &processor{
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
			ctx:       ctx,
			cancel:    cancel,
			drainCh:   drainCh,
			done:      make(chan struct{}),
			workerSem: make(chan struct{}, 3),
			sourceID:  model.SourceIDGoServer,
		}

		// Fill up semaphore (simulate all workers busy)
		for i := 0; i < cap(p.workerSem); i++ {
			p.workerSem <- struct{}{}
		}

		// Close drain channel to trigger drain path
		close(drainCh)

		batch := []*model.Event{{ID: "1"}}
		done := make(chan bool)
		go func() {
			result := p.trySubmitBatch(batch)
			done <- result
		}()

		// Should be blocked waiting for worker permit
		select {
		case <-done:
			t.Fatal("trySubmitBatch should block waiting for worker during drain")
		case <-time.After(50 * time.Millisecond):
			// Expected - blocked waiting for worker
		}

		// Trigger hard shutdown
		cancel()

		// Now trySubmitBatch should return false (batch lost)
		select {
		case result := <-done:
			assert.False(t, result, "trySubmitBatch should return false on hard shutdown")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("trySubmitBatch should unblock on hard shutdown")
		}

		// Release semaphore for cleanup
		for i := 0; i < cap(p.workerSem); i++ {
			<-p.workerSem
		}
	})
}

func TestWaitForWorkers(t *testing.T) {
	t.Parallel()

	t.Run("returns immediately when no workers", func(t *testing.T) {
		p := &processor{
			workerSem: make(chan struct{}, 3),
		}
		// Should not block
		done := make(chan struct{})
		go func() {
			p.waitForWorkers()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("waitForWorkers should return immediately when no workers")
		}
	})

	t.Run("waits for workers to complete", func(t *testing.T) {
		p := &processor{
			workerSem: make(chan struct{}, 3),
		}

		// Simulate worker holding permit
		p.workerSem <- struct{}{}

		done := make(chan struct{})
		go func() {
			p.waitForWorkers()
			close(done)
		}()

		// Should be blocked
		select {
		case <-done:
			t.Fatal("waitForWorkers should block while worker holds permit")
		case <-time.After(50 * time.Millisecond):
			// Expected - still waiting
		}

		// Release the permit (worker done)
		<-p.workerSem

		// Now should complete
		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("waitForWorkers should complete after worker releases permit")
		}
	})
}

// TestDispatcherBatchBehavior tests the actual runDispatcher function behavior:
// - pollTicker only sends full batches (>= flushSize)
// - flushTicker sends partial batches (< flushSize)
// - shutdown drains all remaining events
func TestDispatcherBatchBehavior(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		desc            string
		numEvents       int
		flushSize       int
		flushInterval   time.Duration
		waitBeforeClose time.Duration // time to wait before closing queue
		expectedBatches []int
	}{
		{
			desc:            "poll sends full batches before shutdown drains partial",
			numEvents:       25,
			flushSize:       10,
			flushInterval:   1 * time.Hour, // Long interval so flush won't trigger
			waitBeforeClose: 200 * time.Millisecond,
			expectedBatches: []int{10, 10, 5}, // Poll sends 10+10, shutdown drains 5
		},
		{
			desc:            "flush sends partial batch",
			numEvents:       5,
			flushSize:       10,
			flushInterval:   50 * time.Millisecond, // Short interval to trigger flush
			waitBeforeClose: 200 * time.Millisecond,
			expectedBatches: []int{5},
		},
		{
			desc:            "shutdown drains all events immediately",
			numEvents:       25,
			flushSize:       10,
			flushInterval:   1 * time.Hour,
			waitBeforeClose: 0, // Close immediately
			expectedBatches: []int{10, 10, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var sentBatches []int
			var mu sync.Mutex

			mockClient := mockapi.NewMockClient(mockCtrl)
			mockClient.EXPECT().RegisterEvents(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, req *model.RegisterEventsRequest, _ time.Time) (*model.RegisterEventsResponse, int, error) {
					mu.Lock()
					sentBatches = append(sentBatches, len(req.Events))
					mu.Unlock()
					return &model.RegisterEventsResponse{Errors: make(map[string]*model.RegisterEventsResponseError)}, 0, nil
				},
			).AnyTimes()

			// Create drain channel
			ctx, cancel := context.WithCancel(context.Background())
			drainCh := make(chan struct{})

		p := &processor{
			evtQueue:        newQueue(&queueConfig{capacity: 100}),
			numFlushWorkers: 3,
			flushInterval:   tt.flushInterval,
			flushSize:       tt.flushSize,
			flushTimeout:    10 * time.Second,
			apiClient:       mockClient,
			loggers: log.NewLoggers(&log.LoggersConfig{
				EnableDebugLog: false,
				ErrorLogger:    log.DiscardErrorLogger,
			}),
			ctx:       ctx,
			cancel:    cancel,
			drainCh:   drainCh,
			done:      make(chan struct{}),
			workerSem: make(chan struct{}, 3),
			sourceID:  model.SourceIDGoServer,
		}

		// Push events
			for i := 0; i < tt.numEvents; i++ {
				p.evtQueue.push(&model.Event{ID: fmt.Sprintf("event-%d", i)})
			}

			// Start the actual dispatcher
			go p.runDispatcher()

			// Wait before closing (allows poll/flush to process)
			if tt.waitBeforeClose > 0 {
				time.Sleep(tt.waitBeforeClose)
			}

			// Close shutdown channel to trigger shutdown
			close(drainCh)

			// Wait for dispatcher to finish
			<-p.done

			// Verify (use ElementsMatch since batch order is non-deterministic due to concurrent workers)
			mu.Lock()
			assert.ElementsMatch(t, tt.expectedBatches, sentBatches)
			mu.Unlock()
			assert.Equal(t, 0, p.evtQueue.len(), "queue should be empty after shutdown")
		})
	}
}
