//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package processor

import (
	"errors"
	"fmt"
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
	// Run the processor to poll the latest cache from the server
	Run()

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
	Close()
}

type processor struct {
	cache                   cache.Cache
	featureFlagsCache       cache.FeaturesCache
	pollingInterval         time.Duration
	apiClient               api.Client
	pushLatencyMetricsEvent func(duration time.Duration, api model.APIID)
	pushSizeMetricsEvent    func(sizeByte int, api model.APIID)
	pushErrorEvent          func(err error, api model.APIID)
	tag                     string
	closeCh                 chan struct{}
	loggers                 *log.Loggers
}

// ProcessorConfig is the config for Processor.
type FeatureFlagProcessorConfig struct {
	// Cache
	Cache cache.Cache

	// PollingInterval is a interval of polling feature flags from the server
	PollingInterval time.Duration

	// APIClient is the client for Bucketeer service.
	APIClient api.Client

	// PushLatencyMetricsEvent pushes the get evaluation latency metrics event to the queue.
	PushLatencyMetricsEvent func(duration time.Duration, api model.APIID)

	// PushSizeMetricsEvent pushes the get evaluation size metrics event to the queue.
	PushSizeMetricsEvent func(sizeByte int, api model.APIID)

	// Push error metric events to Bucketeer service
	PushErrorEvent func(err error, api model.APIID)

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

func NewFeatureFlagProcessor(conf *FeatureFlagProcessorConfig) FeatureFlagProcessor {
	p := &processor{
		cache:                   conf.Cache,
		featureFlagsCache:       cache.NewFeaturesCache(conf.Cache),
		pollingInterval:         conf.PollingInterval,
		apiClient:               conf.APIClient,
		pushLatencyMetricsEvent: conf.PushLatencyMetricsEvent,
		pushSizeMetricsEvent:    conf.PushSizeMetricsEvent,
		pushErrorEvent:          conf.PushErrorEvent,
		tag:                     conf.Tag,
		closeCh:                 make(chan struct{}),
		loggers:                 conf.Loggers,
	}
	return p
}

func (p *processor) Run() {
	go p.runProcessLoop()
}

func (p *processor) Close() {
	p.closeCh <- struct{}{}
}

func (p *processor) runProcessLoop() {
	defer func() {
		p.loggers.Debug("bucketeer/cache: runProcessLoop done")
	}()
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()
	// Update the cache once when starts polling
	if err := p.updateCache(); err != nil {
		p.loggers.Errorf("bucketeer/cache: failed to update feature flags cache. Error: %v", err)
	}
	for {
		select {
		case <-ticker.C:
			if err := p.updateCache(); err != nil {
				p.loggers.Errorf("bucketeer/cache: failed to update feature flags cache. Error: %v", err)
				continue
			}
		case <-p.closeCh:
			return
		}
	}
}

func (p *processor) updateCache() error {
	ftsID, err := p.getFeatureFlagsID()
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
			return err
		}
		p.loggers.Debug("bucketeer/cache: updateCache featureFlagsID not found")
	}
	requestedAt, err := p.getFeatureFlagsRequestedAt()
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
			return err
		}
		p.loggers.Debug("bucketeer/cache: updateCache requestedAt not found")
	}
	req := model.NewGetFeatureFlagsRequest(
		p.tag,
		ftsID,
		requestedAt,
	)
	// Get the latest cache from the server
	reqStart := time.Now()
	resp, size, err := p.apiClient.GetFeatureFlags(req)
	if err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
		return err
	}
	p.pushLatencyMetricsEvent(time.Since(reqStart), model.GetFeatureFlags)
	p.pushSizeMetricsEvent(size, model.GetFeatureFlags)

	p.loggers.Debugf("bucketeer/cache: GetFeatureFlags response: %v, size: %d", resp, size)
	// Delete all the local cache and save the new one
	if resp.ForceUpdate {
		if err := p.deleteAllAndSaveLocalCache(resp.FeatureFlagsId, resp.RequestedAt, resp.Features); err != nil {
			p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
			return err
		}
		return nil
	}
	// Update only the updated flags
	err = p.updateLocalCache(
		resp.FeatureFlagsId,
		resp.RequestedAt,
		resp.Features,
		resp.ArchivedFeatureFlagIds,
	)
	if err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
		return err
	}
	return nil
}

// This will delete all the flags in the cache,
// and save the new one return from the server.
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

// This will update the target flags, and delete the archived flags from the cache
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

func (p *processor) newInternalError(err error) error {
	return fmt.Errorf("internal error while updating feature flags: %w", err)
}
