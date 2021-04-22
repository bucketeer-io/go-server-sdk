package bucketeer

import (
	"context"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	registerOnce sync.Once
	latencyMs    = stats.Int64(
		"bucketeer_go_server_sdk_get_evaluation_latency_distribution",
		"Get evaluation latency in milliseconds",
		stats.UnitMilliseconds,
	)
	counter = stats.Int64(
		"bucketeer_go_server_sdk_get_evaluation_call_count",
		"Get evaluation call count",
		stats.UnitDimensionless,
	)
)

var (
	keyFeatureID = tag.MustNewKey("featureID")
	keyStatus    = tag.MustNewKey("status")
)

func RegisterMetrics() error {
	views := []*view.View{
		{
			Name:        latencyMs.Name(),
			Measure:     latencyMs,
			Description: latencyMs.Description(),
			TagKeys:     []tag.Key{keyFeatureID, keyStatus},
			Aggregation: newLatencyDistribution(),
		},
		{
			Name:        counter.Name(),
			Measure:     counter,
			Description: counter.Description(),
			TagKeys:     []tag.Key{keyFeatureID, keyStatus},
			Aggregation: view.Count(),
		},
	}
	var err error
	registerOnce.Do(func() {
		err = view.Register(views...)
	})
	return err
}

func measure(ctx context.Context, v time.Duration) {
	stats.Record(ctx, latencyMs.M(v.Milliseconds()))
}

func count(ctx context.Context) {
	stats.Record(ctx, counter.M(1))
}

func newMetricsContext(ctx context.Context, mutators []tag.Mutator) (context.Context, error) {
	return tag.New(
		ctx,
		mutators...,
	)
}

func newLatencyDistribution() *view.Aggregation {
	const begin, count, exp = 25.0, 9, 2
	dist := make([]float64, count)
	dist[0] = begin
	for i := 1; i < count; i++ {
		dist[i] = dist[i-1] * exp
	}
	return view.Distribution(dist...)
}
