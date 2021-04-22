package bucketeer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func TestRegisterMetrics(t *testing.T) {
	assert.NoError(t, RegisterMetrics())
}

func TestNewMetricsContext(t *testing.T) {
	expectedCtx, err := tag.New(context.Background(), tag.Insert(keyFeatureID, sdkFeatureID))
	assert.NoError(t, err)

	ctx := context.Background()
	mutators := []tag.Mutator{
		tag.Insert(keyFeatureID, sdkFeatureID),
	}
	ctx, err = newMetricsContext(ctx, mutators)
	assert.NoError(t, err)

	assert.Equal(t, expectedCtx, ctx)
}

func TestNewLatencyDistribution(t *testing.T) {
	expected := view.Distribution(25, 50, 100, 200, 400, 800, 1600, 3200, 6400).Buckets
	assert.Equal(t, expected, newLatencyDistribution().Buckets)
}
