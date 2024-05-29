//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
package processor

import (
	"errors"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"

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
}

type segmentUserProcessor struct {
	cache             cache.Cache
	segmentUsersCache cache.SegmentUsersCache
	pollingInterval   time.Duration
	apiClient         api.Client
	closeCh           chan struct{}
	loggers           *log.Loggers
}

// ProcessorConfig is the config for Processor.
type SegmentUserProcessorConfig struct {
	// Cache
	Cache cache.Cache

	// PollingInterval is a interval of polling segment users from the server
	PollingInterval time.Duration

	// APIClient is the client for Bucketeer service.
	APIClient api.Client

	// Loggers is the Bucketeer SDK Loggers.
	Loggers *log.Loggers
}

const (
	segmentUserCacheTTL        = time.Duration(0)
	segmentUsersRequestedAtKey = "bucketeer_segment_users_requested_at"
)

func NewSegmentUserProcessor(conf *SegmentUserProcessorConfig) SegmentUserProcessor {
	p := &segmentUserProcessor{
		cache:             conf.Cache,
		segmentUsersCache: cache.NewSegmentUsersCache(conf.Cache),
		pollingInterval:   conf.PollingInterval,
		apiClient:         conf.APIClient,
		closeCh:           make(chan struct{}),
		loggers:           conf.Loggers,
	}
	return p
}

func (p *segmentUserProcessor) Run() {
	go p.runProcessLoop()
}

func (p *segmentUserProcessor) Close() {
	p.closeCh <- struct{}{}
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
			return err
		}
		p.loggers.Debug("bucketeer/cache: segmentUsers updateCache segment IDs not found")
	}
	requestedAt, err := p.getSegmentUsersRequestedAt()
	if err != nil {
		if !errors.Is(err, cache.ErrNotFound) {
			return err
		}
		p.loggers.Debug("bucketeer/cache: segmentUsers updateCache requestedAt not found")
	}
	req := model.NewGetSegmentUsersRequest(
		ftsID,
		requestedAt,
	)
	// Get the latest cache from the server
	resp, size, err := p.apiClient.GetSegmentUsers(req)
	if err != nil {
		return err
	}
	p.loggers.Debugf("bucketeer/cache: GetSegmentUsers response: %v, size: %d", resp, size)
	// Delete all the local cache and save the new one
	if resp.ForceUpdate {
		return p.deleteAllAndSaveLocalCache(resp.RequestedAt, resp.SegmentUsers)
	}
	// Update only the updated flags
	return p.updateLocalCache(
		resp.RequestedAt,
		resp.SegmentUsers,
		resp.DeletedSegmentIds,
	)
}

// This will delete all the flags in the cache,
// and save the new one return from the server.
func (p *segmentUserProcessor) deleteAllAndSaveLocalCache(
	requestedAt int64,
	segments []*ftproto.SegmentUsers,
) error {
	// Delete all the old flags
	if err := p.segmentUsersCache.DeleteAll(); err != nil {
		return err
	}
	// Save the new flags
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

// This will update the target flags, and delete the archived flags from the cache
func (p *segmentUserProcessor) updateLocalCache(
	requestedAt int64,
	segments []*ftproto.SegmentUsers,
	deletedSegmentIDs []string,
) error {
	// Update the updated flags
	for _, s := range segments {
		if err := p.segmentUsersCache.Put(s); err != nil {
			return err
		}
	}
	// Delete the archived flags
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
