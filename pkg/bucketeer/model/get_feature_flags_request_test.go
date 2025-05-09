package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/version"
)

func TestNewGetFeatureFlagsRequest(t *testing.T) {
	t.Parallel()
	featureFlagsID := "fid-1"
	requestedAt := int64(1)
	sdkVersion := version.SDKVersion
	e := NewGetFeatureFlagsRequest(tag, featureFlagsID, sdkVersion, requestedAt)
	assert.IsType(t, &GetFeatureFlagsRequest{}, e)
	assert.Equal(t, tag, e.Tag)
	assert.Equal(t, featureFlagsID, e.FeatureFlagsID)
	assert.Equal(t, SourceIDGoServer, e.SourceID)
	assert.Equal(t, version.SDKVersion, e.SDKVersion)
}
