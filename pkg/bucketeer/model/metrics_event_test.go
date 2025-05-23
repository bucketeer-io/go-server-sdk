package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewMetricsEvent(t *testing.T) {
	t.Parallel()
	json := json.RawMessage{}
	e := NewMetricsEvent(json, SourceIDGoServer, version.SDKVersion)
	assert.IsType(t, &MetricsEvent{}, e)
	assert.Equal(t, MetricsEventType, e.Type)
	assert.Equal(t, e.SourceID, SourceIDGoServer)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
