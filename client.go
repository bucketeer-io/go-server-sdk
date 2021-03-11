package bucketeer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	event "github.com/ca-dp/bucketeer-go-server-sdk/proto/event/client"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
	"github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Client is bucketeer API client.
type Client interface {
	// BoolVariation fetches variation whose type is boolean using specified feature ID.
	// Default value is used when an error occurs.  A tracker can be used to track metrics which associates with the variation.
	BoolVariation(ctx context.Context, featureID string, def bool) (bool, Tracker)

	// Int64Variation fetches variation whose type is int64 using specified feature ID.
	// Default value is used when an error occurs.  A tracker can be used to track metrics which associates with the variation.
	Int64Variation(ctx context.Context, featureID string, def int64) (int64, Tracker)

	// Float64Variation fetches variation whose type is float64 using specified feature ID.
	// Default value is used when an error occurs.  A tracker can be used to track metrics which associates with the variation.
	Float64Variation(ctx context.Context, featureID string, def float64) (float64, Tracker)

	// StringVariation fetches variation whose type is string using specified feature ID.
	// Default value is used when an error occurs.  A tracker can be used to track metrics which associates with the variation.
	StringVariation(ctx context.Context, featureID string, def string) (string, Tracker)

	// JSONVariation fetches variation using specified feature ID and marshals to destination.
	// A tracker can be used to track metrics which associates with the variation.
	JSONVariation(ctx context.Context, featureID string, dst interface{}) Tracker

	Close()
}

type (
	Host   string
	APIKey string // APIKey is bucketeer API key.
	Tag    string // Tag is feature's tag.
)

// Config is a configuration of bucketeer client.
type Config struct {
	Log                func(error)   // Log is a function which is called when the function is not nil and an error occurs.
	Host               Host          // Host is bucketeer api host name.
	APIKey             APIKey        // APIKey is bucketeer api key.
	Tag                Tag           // Tag is specified in fetching variations.
	EventFlushInterval time.Duration // EventFlushInterval is an interval to flush events. Default: 1s
	EventBufferSize    int           // EventBufferSize is a buffer size of events channel as queue. Default: 65536
}

const (
	defaultBufferSize    = 1 << 16
	defaultFlushInterval = time.Second
	defaultFlushSize     = 100
)

type client struct {
	grpc               gateway.GatewayClient
	apiKey             string
	tag                string
	logFunc            func(error)
	eventsCh           chan *event.Event
	closeCh            chan struct{}
	eventBufferSize    int
	eventFlushSize     int
	eventFlushInterval time.Duration
	events             []*event.Event
	eventMap           map[string]*event.Event
	now                func() time.Time
}

func NewClient(ctx context.Context, cfg *Config) (Client, error) {
	creds, err := newPerRPCCredentials(string(cfg.APIKey))
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(string(cfg.Host),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(creds),
	)
	cli := &client{
		grpc:               gateway.NewGatewayClient(conn),
		apiKey:             string(cfg.APIKey),
		tag:                string(cfg.Tag),
		logFunc:            cfg.Log,
		eventBufferSize:    cfg.EventBufferSize,
		eventFlushInterval: cfg.EventFlushInterval,
		closeCh:            make(chan struct{}),
		now:                time.Now,
	}
	if cli.eventBufferSize <= 0 {
		cli.eventBufferSize = defaultBufferSize
	}
	if cli.eventFlushInterval <= 0 {
		cli.eventFlushInterval = defaultFlushInterval
	}
	if cli.eventFlushSize <= 0 {
		cli.eventFlushSize = defaultFlushSize
	}
	cli.eventsCh = make(chan *event.Event, cli.eventBufferSize)
	cli.resetEvents()
	go cli.start(ctx)
	return cli, nil
}

func (c *client) log(err error) {
	if c.logFunc == nil {
		return
	}
	c.logFunc(err)
}

func (c *client) start(ctx context.Context) {
	defer close(c.closeCh)
	ticker := time.NewTicker(c.eventFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.flushEvents(context.Background())
			return
		case <-ticker.C:
			c.flushEvents(ctx)
		case e := <-c.eventsCh:
			c.events = append(c.events, e)
			c.eventMap[e.Id] = e
			if len(c.events) < c.eventFlushSize {
				continue
			}
			c.flushEvents(ctx)
		}
	}
}

func (c *client) Close() {
	<-c.closeCh
}

func (c *client) flushEvents(ctx context.Context) {
	req := &gateway.RegisterEventsRequest{Events: c.events}
	resp, err := c.grpc.RegisterEvents(ctx, req)
	if err != nil {
		if c.retriable(err) {
			return
		}
		c.log(&Error{
			Message:     "failed to register events",
			originalErr: err,
		})
		c.resetEvents()
		return
	}
	for id, err := range resp.Errors {
		e := c.eventMap[id]
		if e == nil {
			continue
		}
		if !err.GetRetriable() {
			c.log(&Error{Message: err.Message})
			continue
		}
		c.eventsCh <- e
	}
	c.resetEvents()
}

func (c *client) retriable(err error) bool {
	code := status.Code(err)
	switch code {
	case codes.Internal,
		codes.Unavailable:
		return true
	}
	return false
}

func (c *client) resetEvents() {
	c.events = make([]*event.Event, 0, c.eventFlushSize)
	c.eventMap = make(map[string]*event.Event, c.eventFlushSize)
}

func (c *client) fetchVariation(ctx context.Context, featureID string) (string, Tracker, error) {
	userID := userID(ctx)
	attrs := attributes(ctx)
	req := &gateway.GetEvaluationRequest{
		Tag:       c.tag,
		User:      &user.User{Id: userID, Data: attrs},
		FeatureId: featureID,
	}
	resp, err := c.grpc.GetEvaluation(ctx, req)
	if err != nil {
		e := &Error{
			Message:        "failed to get evaluation",
			UserID:         userID,
			UserAttributes: attrs,
			FeatureID:      featureID,
			originalErr:    err,
		}
		c.log(e)
		return "", nil, e
	}
	if resp.Evaluation == nil || resp.Evaluation.FeatureId != featureID {
		c.log(&Error{
			Message:        "evaluation not found",
			UserID:         userID,
			UserAttributes: attrs,
			FeatureID:      featureID,
		})
	}
	c.sendEvaluationEvent(ctx, resp.Evaluation)

	t := &tracker{
		ch:          c.eventsCh,
		now:         c.now,
		userID:      userID,
		attributes:  attrs,
		evaluations: []*feature.Evaluation{resp.Evaluation},
		logFunc:     c.logFunc,
	}
	return resp.Evaluation.Variation.Value, t, nil
}

func (c *client) sendEvaluationEvent(ctx context.Context, evaluation *feature.Evaluation) {
	userID, attrs := userID(ctx), attributes(ctx)
	e := &event.EvaluationEvent{
		Timestamp:      c.now().Unix(),
		FeatureId:      evaluation.FeatureId,
		FeatureVersion: evaluation.FeatureVersion,
		UserId:         userID,
		VariationId:    evaluation.VariationId,
		User:           &user.User{Id: userID, Data: attrs},
		Reason:         evaluation.Reason,
	}
	any, err := ptypes.MarshalAny(e)
	if err != nil {
		c.log(&Error{
			Message:        "failed to marshal evaluation event",
			UserID:         userID,
			UserAttributes: attrs,
			FeatureID:      evaluation.FeatureId,
			Evaluation:     e,
			originalErr:    err,
		})
	}
	c.eventsCh <- &event.Event{
		Id:    uuid.New().String(),
		Event: any,
	}
}

func (c *client) BoolVariation(ctx context.Context, featureID string, def bool) (bool, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	if variation == "" {
		return def, t
	}
	v, err := strconv.ParseBool(variation)
	if err == nil {
		return v, t
	}
	c.log(&Error{
		Message:        "failed to parse bool",
		UserID:         userID(ctx),
		UserAttributes: attributes(ctx),
		FeatureID:      featureID,
		originalErr:    err,
	})
	return v, t
}

func (c *client) Int64Variation(ctx context.Context, featureID string, def int64) (int64, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	if variation == "" {
		return def, t
	}
	const base, bitSize = 10, 64
	v, err := strconv.ParseInt(variation, base, bitSize)
	if err == nil {
		return v, t
	}
	c.log(&Error{
		Message:        "failed to parse int",
		UserID:         userID(ctx),
		UserAttributes: attributes(ctx),
		FeatureID:      featureID,
		originalErr:    err,
	})
	return v, t
}

func (c *client) Float64Variation(ctx context.Context, featureID string, def float64) (float64, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	if variation == "" {
		return def, t
	}
	const bitSize = 64
	v, err := strconv.ParseFloat(variation, bitSize)
	if err == nil {
		return v, t
	}
	c.log(&Error{
		Message:        "failed to parse float",
		UserID:         userID(ctx),
		UserAttributes: attributes(ctx),
		FeatureID:      featureID,
		originalErr:    err,
	})
	return v, t
}

func (c *client) StringVariation(ctx context.Context, featureID string, def string) (string, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	if variation == "" {
		return def, t
	}
	return variation, t
}

func (c *client) JSONVariation(ctx context.Context, featureID string, dst interface{}) Tracker {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return t
	}
	if variation == "" {
		return t
	}
	err = json.NewDecoder(strings.NewReader(variation)).Decode(dst)
	if err != nil {
		c.log(&Error{
			Message:        "failed to decode json",
			UserID:         userID(ctx),
			UserAttributes: attributes(ctx),
			FeatureID:      featureID,
			originalErr:    err,
		})
	}
	return t
}

type perRPCCredentials struct {
	APIKey string
}

func newPerRPCCredentials(apiKey string) (credentials.PerRPCCredentials, error) {
	return perRPCCredentials{
		APIKey: apiKey,
	}, nil
}

func (c perRPCCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.APIKey,
	}, nil
}

func (c perRPCCredentials) RequireTransportSecurity() bool {
	return true
}
