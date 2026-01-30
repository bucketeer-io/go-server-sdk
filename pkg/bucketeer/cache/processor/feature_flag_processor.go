//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../../test/mock/$GOPACKAGE/$GOFILE
package processor

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"

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

	// IsReady returns true if the processor has completed at least one successful cache update
	IsReady() bool
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
	sdkVersion              string
	sourceID                model.SourceIDType
	closeCh                 chan struct{}
	loggers                 *log.Loggers
	ready                   atomic.Bool
	healingInProgress       atomic.Bool // Track if we're waiting for a refresh after healing
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

	// SDKVersion is the SDK version.
	SDKVersion string

	// SourceID is the source ID of the SDK.
	SourceID model.SourceIDType
}

const (
	cacheTTL                   = time.Duration(0)
	featureFlagsIDKey          = "bucketeer_feature_flags_id"
	featureFlagsRequestedAtKey = "bucketeer_feature_flags_requested_at"

	// minAbsoluteDeadline is the minimum deadline for retry operations.
	// This ensures at least some retry attempts even with tight polling intervals.
	minAbsoluteDeadline = 5 * time.Second
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
		sdkVersion:              conf.SDKVersion,
		sourceID:                conf.SourceID,
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

func (p *processor) IsReady() bool {
	return p.ready.Load()
}

func (p *processor) runProcessLoop() {
	defer func() {
		p.loggers.Debug("bucketeer/cache: runProcessLoop done")
	}()
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()
	// Update the cache once when starts polling
	ctx := context.Background()
	if err := p.updateCache(ctx); err != nil {
		p.loggers.Errorf("bucketeer/cache: failed to update feature flags cache. Error: %v", err)
	}
	for {
		select {
		case <-ticker.C:
			if err := p.updateCache(ctx); err != nil {
				p.loggers.Errorf("bucketeer/cache: failed to update feature flags cache. Error: %v", err)
				continue
			}
		case <-p.closeCh:
			return
		}
	}
}

func (p *processor) updateCache(ctx context.Context) error {
	// Calculate deadline: leave 10% buffer before next tick
	// This ensures retries complete before the next polling cycle
	deadlineDuration := p.pollingInterval * 9 / 10

	// Enforce minimum absolute deadline to ensure at least some retry attempts
	deadlineDuration = max(deadlineDuration, minAbsoluteDeadline)

	deadline := time.Now().Add(deadlineDuration)

	// Self-healing check: if cache is stale, force full refresh
	// This handles cases where polling was delayed due to CPU pressure or other issues
	if err := p.checkAndHealStaleCache(); err != nil {
		p.loggers.Errorf("bucketeer/cache: failed to check/heal stale cache: %v", err)
		// Continue anyway - we'll try to update with whatever we have
	}

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

	// Log the request parameters to help diagnose stale cache issues
	// If ftsID matches what server has, server returns empty Features[] (no changes)
	// This is a common cause of "stale flag" reports - server thinks nothing changed
	p.loggers.Debugf("bucketeer/cache: requesting flags - tag=%s, ftsID=%s, requestedAt=%d",
		p.tag, ftsID, requestedAt)

	req := model.NewGetFeatureFlagsRequest(
		p.tag,
		ftsID,
		p.sdkVersion,
		p.sourceID,
		requestedAt,
	)
	// Get the latest cache from the server with retry support
	reqStart := time.Now()
	resp, size, err := p.apiClient.GetFeatureFlags(ctx, req, deadline)
	if err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
		return err
	}
	p.pushLatencyMetricsEvent(time.Since(reqStart), model.GetFeatureFlags)
	p.pushSizeMetricsEvent(size, model.GetFeatureFlags)

	// We convert the response to the proto message because it uses less memory in the cache,
	// and the evaluation module uses proto messages.
	pbResp := model.ConvertFeatureFlagsResponse(resp)

	// Enhanced logging to help diagnose "some instances updated, others not" issues
	// Key things to check:
	// - If forceUpdate=false and featuresCount=0, server detected no changes (check server-side caching)
	// - If featureFlagsId matches previous ftsID, incremental update path is taken
	// - If archivedCount > 0, flags are being deleted from cache
	p.loggers.Debugf(
		"bucketeer/cache: GetFeatureFlags response - forceUpdate=%v, newFtsID=%s, "+
			"featuresCount=%d, archivedCount=%d, requestedAt=%d, size=%d",
		pbResp.ForceUpdate, pbResp.FeatureFlagsId, len(pbResp.Features),
		len(pbResp.ArchivedFeatureFlagIds), pbResp.RequestedAt, size)

	// Delete all the local cache and save the new one
	if pbResp.ForceUpdate {
		if err := p.deleteAllAndSaveLocalCache(pbResp.FeatureFlagsId, pbResp.RequestedAt, pbResp.Features); err != nil {
			p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
			return err
		}
		p.ready.Store(true)
		// Log healing completion for observability
		if p.healingInProgress.Load() {
			p.loggers.Infof(
				"bucketeer/cache: HEALING COMPLETED via force update, featuresCount=%d",
				len(pbResp.Features))
		}
		p.healingInProgress.Store(false)
		return nil
	}

	// Incremental update path - only changed flags are applied
	// If a flag change isn't reflected, check:
	// 1. Is the flag included in pbResp.Features? If not, server didn't detect the change
	// 2. Is the flag in ArchivedFeatureFlagIds? It might have been archived instead of updated
	if len(pbResp.Features) > 0 {
		featureIDs := make([]string, 0, len(pbResp.Features))
		for _, f := range pbResp.Features {
			featureIDs = append(featureIDs, f.Id)
		}
		p.loggers.Debugf("bucketeer/cache: incremental update - updating flags: %v", featureIDs)
	}

	// Update only the updated flags
	err = p.updateLocalCache(
		pbResp.FeatureFlagsId,
		pbResp.RequestedAt,
		pbResp.Features,
		pbResp.ArchivedFeatureFlagIds,
	)
	if err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetFeatureFlags)
		return err
	}
	p.ready.Store(true)
	// Log healing completion for observability
	if p.healingInProgress.Load() {
		p.loggers.Infof("bucketeer/cache: HEALING COMPLETED successfully via incremental update")
	}
	p.healingInProgress.Store(false)
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

// checkAndHealStaleCache detects if the cache is stale and forces a full refresh if needed.
// Returns error only if the healing process fails, not if cache is fresh.
//
// This healing mechanism addresses scenarios where:
// - Polling was delayed due to CPU pressure, GC pauses, or container throttling
// - Metadata got corrupted or externally deleted
func (p *processor) checkAndHealStaleCache() error {
	// Only check if we're already ready (have data in cache)
	if !p.ready.Load() {
		return nil
	}

	requestedAt, err := p.getFeatureFlagsRequestedAt()
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) || errors.Is(err, cache.ErrInvalidType) {
			// Missing or corrupted requestedAt with ready=true
			// Check if this is expected (healing in progress) or unexpected
			if p.healingInProgress.Load() {
				// Expected: we cleared metadata and are waiting for next successful poll
				p.loggers.Debug("bucketeer/cache: healing in progress, waiting for refresh")
				return nil
			}

			// Unexpected: metadata missing/corrupted but we never triggered healing
			// This could be external deletion, type corruption, or memory corruption
			p.loggers.Errorf("bucketeer/cache: UNEXPECTED STATE - ready=true but requestedAt is missing or corrupted (%v). "+
				"Triggering healing.", err)

			// Clear ALL metadata to force full refresh from server
			// Without this, featureFlagsID might still exist and server would return
			// incremental update instead of force update, preventing proper healing
			p.cache.Delete(featureFlagsIDKey)
			p.cache.Delete(featureFlagsRequestedAtKey)

			// Mark healing in progress so next check knows metadata deletion is expected
			// Keep ready=true because cache data is still present and usable
			p.healingInProgress.Store(true)

			p.loggers.Infof("bucketeer/cache: HEALING STARTED - cleared cache metadata due to corruption. " +
				"Next poll will do full refresh. Cache remains usable.")
			return nil
		}
		return fmt.Errorf("failed to get requestedAt: %w", err)
	}

	cacheAge := time.Now().Unix() - requestedAt
	// Cache is stale if it's older than 2x polling interval
	// This indicates polling was delayed or failed multiple times
	maxAcceptableAge := int64(p.pollingInterval.Seconds() * 2)

	// Log when cache is fresh to confirm the check is running
	// This helps distinguish "check didn't run" from "check ran but found fresh cache"
	if cacheAge <= maxAcceptableAge {
		p.loggers.Debugf("bucketeer/cache: cache freshness check passed - age=%ds, maxAcceptable=%ds",
			cacheAge, maxAcceptableAge)
		return nil
	}

	// cacheAge > maxAcceptableAge means polling was delayed/failed
	// Common causes: CPU throttling, long GC pause, network issues, container eviction
	p.loggers.Errorf("bucketeer/cache: STALE CACHE DETECTED! Age: %ds (max: %ds). Forcing full refresh.",
		cacheAge, maxAcceptableAge)

	// Clear metadata to force full refresh from server
	// This makes the SDK send an empty featureFlagsID, so gateway returns all flags
	p.cache.Delete(featureFlagsIDKey)
	p.cache.Delete(featureFlagsRequestedAtKey)

	// Mark that we're in healing mode
	// This tells future checks that missing metadata is expected
	p.healingInProgress.Store(true)

	// NOTE: We keep ready=true because the stale cache is still usable
	// Using old cache data is better than forcing the app to use default values
	// The app can continue serving requests while we refresh in the background

	p.loggers.Infof("bucketeer/cache: HEALING STARTED - cleared cache metadata. " +
		"Next poll will do full refresh. Cache remains usable.")

	return nil
}
