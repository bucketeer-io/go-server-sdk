package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewRegisterEventsRequest(t *testing.T) {
	t.Parallel()
	event := []*Event{&Event{}}
	r := NewRegisterEventsRequest(event, SourceIDGoServer)
	assert.IsType(t, &RegisterEventsRequest{}, r)
	assert.Equal(t, event, r.Events)
	assert.Equal(t, SourceIDGoServer, r.SourceID)
	assert.Equal(t, version.SDKVersion, r.SDKVersion)
}
