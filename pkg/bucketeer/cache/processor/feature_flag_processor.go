//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

// FeatureFlagProcessor defines the interface for processing feature flags cache.
//
// In the background, the Processor polls the latest feature flags from the server,
// and updates them in the local cache.
type FeatureFlagProcessor interface {
	// Get all the feature flags by tag
	GetFeatureFlag(id string) (*ftproto.Feature, error)

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
	Close(ctx context.Context) error
}

type processor struct {
	cache             cache.Cache
	featureFlagsCache cache.FeaturesCache
	pollingInterval   time.Duration
	apiClient         api.Client
	loggers           *log.Loggers
	closeCh           chan struct{}
	workerWG          sync.WaitGroup
	tag               string
}

// ProcessorConfig is the config for Processor.
type FeatureFlagProcessorConfig struct {
	// Cache
	Cache cache.Cache

	// PollingInterval is a interval of polling feature flags from the server
	PollingInterval time.Duration

	// APIClient is the client for Bucketeer service.
	APIClient api.Client

	// Loggers is the Bucketeer SDK Loggers.
	Loggers *log.Loggers

	// Tag is the Feature Flag tag
	// The tag is set when a Feature Flag is created, and can be retrieved from the admin console.
	Tag string
}

const (
	cacheTTL                   = time.Duration(0)
	featureFlagsIDKey          = "bucketeer_feature_flags_id"
	featureFlagsRequestedAtKey = "bucketeer_feature_flags_requested_at"
)

func NewFeatureFlagProcessor(ctx context.Context, conf *FeatureFlagProcessorConfig) FeatureFlagProcessor {
	p := &processor{
		cache:             conf.Cache,
		featureFlagsCache: cache.NewFeaturesCache(conf.Cache),
		pollingInterval:   conf.PollingInterval,
		apiClient:         conf.APIClient,
		loggers:           conf.Loggers,
		closeCh:           make(chan struct{}),
		tag:               conf.Tag,
	}
	go p.startWorker(ctx)
	return p
}

func (p *processor) GetFeatureFlag(id string) (*ftproto.Feature, error) {
	return p.featureFlagsCache.Get(id)
}

func (p *processor) Close(ctx context.Context) error {
	select {
	case <-p.closeCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("bucketeer/cache: ctx is canceled: %v", ctx.Err())
	}
}

func (p *processor) startWorker(ctx context.Context) {
	p.workerWG.Add(1)
	go p.runWorkerProcessLoop(ctx)
	p.workerWG.Wait()
	close(p.closeCh)
}

func (p *processor) runWorkerProcessLoop(ctx context.Context) {
	defer func() {
		p.loggers.Debug("bucketeer/cache: runWorkerProcessLoop done")
		p.workerWG.Done()
	}()
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := p.updateCache(); err != nil {
				p.loggers.Errorf("bucketeer/cache: failed to update feature flags cache. Error: %v", err)
				continue
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *processor) updateCache() error {
	ftsID, err := p.getFeatureFlagsID()
	if err != nil {
		return err
	}
	requestedAt, err := p.getFeatureFlagsRequestedAt()
	if err != nil {
		return err
	}
	req := model.NewGetFeatureFlagsRequest(
		p.tag,
		ftsID,
		requestedAt,
	)
	// Get the latest cache from the server
	resp, _, err := p.apiClient.GetFeatureFlags(req)
	if err != nil {
		return err
	}
	// Delete all the local cache and save the new one
	if resp.ForceUpdate {
		return p.deleteAllAndSaveLocalCache(resp.FeatureFlagsId, resp.RequestedAt, resp.Features)
	}
	// Update only the updated flags
	return p.updateLocalCache(
		resp.FeatureFlagsId,
		resp.RequestedAt,
		resp.Features,
		resp.ArchivedFeatureFlagIds,
	)
}

func (p *processor) deleteAllAndSaveLocalCache(
	featureFlagsID string,
	requestedAt int64,
	features []*ftproto.Feature,
) error {
	// Delete all the old flags
	if err := p.featureFlagsCache.DeleteAll(); err != nil {
		return err
	}
	// Save the new flags
	for _, f := range features {
		if err := p.featureFlagsCache.Put(f); err != nil {
			return err
		}
	}
	// Update the featureFlagsID
	if err := p.putFeatureFlagsID(featureFlagsID); err != nil {
		return err
	}
	// Update the requestedAt
	if err := p.putFeatureFlagsRequestedAt(requestedAt); err != nil {
		return err
	}
	return nil
}

func (p *processor) updateLocalCache(
	featureFlagsID string,
	requestedAt int64,
	features []*ftproto.Feature,
	archivedFeatureFlagIds []string,
) error {
	// Update the updated flags
	for _, f := range features {
		if err := p.featureFlagsCache.Put(f); err != nil {
			return err
		}
	}
	// Delete the archived flags
	for _, fID := range archivedFeatureFlagIds {
		p.featureFlagsCache.Delete(fID)
	}
	// Update the featureFlagsID
	if err := p.putFeatureFlagsID(featureFlagsID); err != nil {
		return err
	}
	// Update the requestedAt
	if err := p.putFeatureFlagsRequestedAt(requestedAt); err != nil {
		return err
	}
	return nil
}

func (p *processor) getFeatureFlagsID() (string, error) {
	value, err := p.cache.Get(featureFlagsIDKey)
	if err != nil {
		return "", err
	}
	v, err := cache.String(value)
	if err != nil {
		return "", cache.ErrInvalidType
	}
	return v, nil
}

func (p *processor) putFeatureFlagsID(id string) error {
	return p.cache.Put(featureFlagsIDKey, id, cacheTTL)
}

func (p *processor) getFeatureFlagsRequestedAt() (int64, error) {
	value, err := p.cache.Get(featureFlagsRequestedAtKey)
	if err != nil {
		return 0, err
	}
	v, err := cache.Int64(value)
	if err != nil {
		return 0, cache.ErrInvalidType
	}
	return v, nil
}

func (p *processor) putFeatureFlagsRequestedAt(timestamp int64) error {
	return p.cache.Put(featureFlagsRequestedAtKey, timestamp, cacheTTL)
}
