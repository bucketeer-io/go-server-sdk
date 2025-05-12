package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewGetSegmentUsersRequest(t *testing.T) {
	t.Parallel()
	segmentIDs := []string{"seg-1", "seg-2"}
	requestedAt := int64(1)
	sdkVersion := version.SDKVersion
	e := NewGetSegmentUsersRequest(segmentIDs, requestedAt, sdkVersion, SourceIDGoServer)
	assert.IsType(t, &GetSegmentUsersRequest{}, e)
	assert.Equal(t, segmentIDs, e.SegmentIDs)
	assert.Equal(t, requestedAt, e.RequestedAt)
	assert.Equal(t, SourceIDGoServer, e.SourceID)
	assert.Equal(t, sdkVersion, e.SDKVersion)
}
