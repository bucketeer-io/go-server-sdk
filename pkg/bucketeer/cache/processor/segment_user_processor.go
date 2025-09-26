//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../../test/mock/$GOPACKAGE/$GOFILE
package processor

import (
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

// SegmentUserProcessor defines the interface for processing segment users cache.
//
// In the background, the Processor polls the latest segment users from the server,
// and updates them in the local cache.
type SegmentUserProcessor interface {
	// Run the processor to poll the latest cache from the server
	Run()

	// Close tears down all Processor activities, after ensuring that all events have been delivered.
	Close()

	// IsReady returns true if the processor has completed at least one successful cache update
	IsReady() bool
}

type segmentUserProcessor struct {
	cache                   cache.Cache
	segmentUsersCache       cache.SegmentUsersCache
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
}

// ProcessorConfig is the config for Processor.
type SegmentUserProcessorConfig struct {
	// Cache
	Cache cache.Cache

	// PollingInterval is a interval of polling segment users from the server
	PollingInterval time.Duration

	// APIClient is the client for Bucketeer service.
	APIClient api.Client

	// PushLatencyMetricsEvent pushes the get evaluation latency metrics event to the queue.
	PushLatencyMetricsEvent func(duration time.Duration, api model.APIID)

	// PushSizeMetricsEvent pushes the get evaluation size metrics event to the queue.
	PushSizeMetricsEvent func(sizeByte int, api model.APIID)

	// PushErrorEvent pushes error metric events to Bucketeer service
	PushErrorEvent func(err error, api model.APIID)

	// Loggers is the Bucketeer SDK Loggers.
	Loggers *log.Loggers

	// Tag is the Feature Flag tag
	// The tag is set when a Feature Flag is created, and can be retrieved from the admin console.
	//
	// Note: this tag is used to report metric events.
	Tag string

	// SDKVersion is the SDK version.
	SDKVersion string

	// SourceID is the source ID of the SDK.
	SourceID model.SourceIDType
}

const (
	segmentUserCacheTTL        = time.Duration(0)
	segmentUsersRequestedAtKey = "bucketeer_segment_users_requested_at"
)

func NewSegmentUserProcessor(conf *SegmentUserProcessorConfig) SegmentUserProcessor {
	p := &segmentUserProcessor{
		cache:                   conf.Cache,
		segmentUsersCache:       cache.NewSegmentUsersCache(conf.Cache),
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

func (p *segmentUserProcessor) Run() {
	go p.runProcessLoop()
}

func (p *segmentUserProcessor) Close() {
	p.closeCh <- struct{}{}
}

func (p *segmentUserProcessor) IsReady() bool {
	return p.ready.Load()
}

func (p *segmentUserProcessor) runProcessLoop() {
	defer func() {
		p.loggers.Debug("bucketeer/cache: segmentUsers runProcessLoop done")
	}()
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()
	// Update the cache once when starts polling
	if err := p.updateCache(); err != nil {
		p.loggers.Errorf("bucketeer/cache: segmentUsers failed to update segment users cache. Error: %v", err)
	}
	for {
		select {
		case <-ticker.C:
			if err := p.updateCache(); err != nil {
				p.loggers.Errorf("bucketeer/cache: segmentUsers failed to update segment users cache. Error: %v", err)
				continue
			}
		case <-p.closeCh:
			return
		}
	}
}

func (p *segmentUserProcessor) updateCache() error {
	ftsID, err := p.segmentUsersCache.GetSegmentIDs()
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			p.pushErrorEvent(p.newInternalError(err), model.GetSegmentUsers)
			return err
		}
		p.loggers.Debug("bucketeer/cache: segmentUsers updateCache segment IDs not found")
	}
	requestedAt, err := p.getSegmentUsersRequestedAt()
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			p.pushErrorEvent(p.newInternalError(err), model.GetSegmentUsers)
			return err
		}
		p.loggers.Debug("bucketeer/cache: segmentUsers updateCache requestedAt not found")
	}
	req := model.NewGetSegmentUsersRequest(
		ftsID,
		requestedAt,
		p.sdkVersion,
		p.sourceID,
	)
	// Get the latest cache from the server
	reqStart := time.Now()
	resp, size, err := p.apiClient.GetSegmentUsers(req)
	if err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetSegmentUsers)
		return err
	}
	// We convert the response to the proto message because it uses less memory in the cache,
	// and the evaluation module uses proto messages.
	pbResp := model.ConvertSegmentUsersResponse(resp)
	p.pushLatencyMetricsEvent(time.Since(reqStart), model.GetSegmentUsers)
	p.pushSizeMetricsEvent(size, model.GetSegmentUsers)

	p.loggers.Debugf("bucketeer/cache: GetSegmentUsers response: %v, size: %d", resp, size)
	// Delete all the local cache and save the new one
	if resp.ForceUpdate {
		if err := p.deleteAllAndSaveLocalCache(pbResp.RequestedAt, pbResp.SegmentUsers); err != nil {
			p.pushErrorEvent(p.newInternalError(err), model.GetSegmentUsers)
			return err
		}
		p.ready.Store(true)
		return nil
	}
	// Update only the updated segment users
	if err := p.updateLocalCache(pbResp.RequestedAt, pbResp.SegmentUsers, pbResp.DeletedSegmentIds); err != nil {
		p.pushErrorEvent(p.newInternalError(err), model.GetSegmentUsers)
		return err
	}
	p.ready.Store(true)
	return nil
}

// This will delete all the segment users in the cache,
// and save the new one return from the server.
func (p *segmentUserProcessor) deleteAllAndSaveLocalCache(
	requestedAt int64,
	segments []*ftproto.SegmentUsers,
) error {
	// Delete all the old segment users
	if err := p.segmentUsersCache.DeleteAll(); err != nil {
		return err
	}
	// Save the new segment users
	for _, s := range segments {
		if err := p.segmentUsersCache.Put(s); err != nil {
			return err
		}
	}
	// Update the requestedAt
	if err := p.putSegmentUsersRequestedAt(requestedAt); err != nil {
		return err
	}
	return nil
}

// This will update the updated and deleted segment users from the cache
func (p *segmentUserProcessor) updateLocalCache(
	requestedAt int64,
	segments []*ftproto.SegmentUsers,
	deletedSegmentIDs []string,
) error {
	// Update the updated segment users
	for _, s := range segments {
		if err := p.segmentUsersCache.Put(s); err != nil {
			return err
		}
	}
	// Delete the archived segment users
	for _, sID := range deletedSegmentIDs {
		p.segmentUsersCache.Delete(sID)
	}
	// Update the requestedAt
	if err := p.putSegmentUsersRequestedAt(requestedAt); err != nil {
		return err
	}
	return nil
}

func (p *segmentUserProcessor) getSegmentUsersRequestedAt() (int64, error) {
	value, err := p.cache.Get(segmentUsersRequestedAtKey)
	if err != nil {
		return 0, err
	}
	v, err := cache.Int64(value)
	if err != nil {
		return 0, cache.ErrInvalidType
	}
	return v, nil
}

func (p *segmentUserProcessor) putSegmentUsersRequestedAt(timestamp int64) error {
	return p.cache.Put(segmentUsersRequestedAtKey, timestamp, segmentUserCacheTTL)
}

func (p *segmentUserProcessor) newInternalError(err error) error {
	return fmt.Errorf("internal error while updating segment users: %w", err)
}
