package bucketeer_go_sdk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	event "github.com/ca-dp/bucketeer-go-sdk/proto/event/client"
	"github.com/ca-dp/bucketeer-go-sdk/proto/feature"
	"github.com/ca-dp/bucketeer-go-sdk/proto/gateway"
	"github.com/ca-dp/bucketeer-go-sdk/proto/user"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

type client struct {
	grpc               gateway.GatewayClient
	apiKey             string
	tag                string
	logFunc            func(error)
	eventsCh           chan *event.ExperimentEvent
	eventBufferSize    int
	eventFlushSize     int
	eventFlushInterval time.Duration
	events             []*event.Event
	experiments        map[string]*event.ExperimentEvent
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
		now:                time.Now,
	}
	if cli.eventBufferSize < 0 {
		cli.eventBufferSize = 1 << 16
	}
	if cli.eventFlushInterval < 0 {
		cli.eventFlushInterval = time.Second
	}
	cli.eventsCh = make(chan *event.ExperimentEvent, cli.eventBufferSize)
	cli.resetEvents()
	cli.resetExperiments()
	go cli.start(ctx)
	return cli, nil
}

func (c *client) log(err error) {
	if c.logFunc == nil {
		return
	}
	c.log(err)
}

func (c *client) start(ctx context.Context) {
	ticker := time.NewTicker(c.eventFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.flushEvents(context.Background())
		case <-ticker.C:
			c.flushEvents(ctx)
		case experiment := <-c.eventsCh:
			eventAny, err := ptypes.MarshalAny(experiment)
			if err != nil {
				const msg = "Failed marshal protobuf any"
				c.log(newError(ctx, err, experiment.FeatureId, experiment, msg))
				continue
			}
			e := &event.Event{
				Id:    uuid.New().String(),
				Event: eventAny,
			}
			c.events = append(c.events, e)
			c.experiments[e.Id] = experiment
			if len(c.events) < c.eventFlushSize {
				continue
			}
			c.flushEvents(ctx)
		}
	}
}

func (c *client) flushEvents(ctx context.Context) {
	req := &gateway.RegisterEventsRequest{Events: c.events}
	resp, err := c.grpc.RegisterEvents(ctx, req)
	if err != nil {
		if c.retriable(err) {
			return
		}
		const msg = "Failed to register events"
		c.log(newError(ctx, err, "", nil, msg))
		c.resetExperiments()
		c.resetEvents()
		return
	}
	for id, err := range resp.Errors {
		e := c.experiments[id]
		if e == nil {
			continue
		}
		if !err.GetRetriable() {
			c.log(newError(ctx, nil, e.FeatureId, e, err.Message))
			continue
		}
		c.eventsCh <- e
	}
	c.resetExperiments()
	c.resetEvents()
}

func (c *client) retriable(err error) bool {
	return false
}

func (c *client) resetEvents() {
	c.events = make([]*event.Event, 0, c.eventFlushSize)
}

func (c *client) resetExperiments() {
	c.experiments = make(map[string]*event.ExperimentEvent, c.eventFlushSize)
}

func (c *client) fetchVariation(ctx context.Context, featureID string) (string, Tracker, error) {
	userID := userID(ctx)
	attrs := attributes(ctx)
	req := &gateway.GetEvaluationsRequest{
		Tag:       c.tag,
		User:      &user.User{Id: userID, Data: attrs},
		FeatureId: featureID,
	}
	resp, err := c.grpc.GetEvaluations(ctx, req)
	if err != nil {
		const msg = "Failed to get evaluations"
		err = newError(ctx, err, featureID, nil, msg)
		c.log(err)
		return "", nil, err
	}
	if resp.State != feature.UserEvaluations_FULL {
		msg := fmt.Sprintf("Unexpected state, %s", resp.State.String())
		err := newError(ctx, nil, featureID, nil, msg)
		c.log(err)
		return "", nil, err
	}
	var evaluation *feature.Evaluation
	for _, e := range resp.Evaluations.Evaluations {
		if e.FeatureId != featureID {
			continue
		}
		evaluation = e
	}
	if evaluation == nil || evaluation.Variation == nil {
		const msg = "Evaluation or variation not found"
		err := newError(ctx, nil, featureID, nil, msg)
		c.log(err)
		return "", nil, err
	}
	t := &tracker{
		ch:             c.eventsCh,
		now:            c.now,
		experimentID:   evaluation.Id,
		featureID:      featureID,
		featureVersion: evaluation.FeatureVersion,
		variationID:    evaluation.VariationId,
		userID:         userID,
		attributes:     attrs,
	}
	return evaluation.Variation.Value, t, nil
}

func (c *client) BoolVariation(ctx context.Context, featureID string, def bool) (bool, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	v, err := strconv.ParseBool(variation)
	if err == nil {
		return v, t
	}
	const msg = "Failed to parse bool"
	c.log(newError(ctx, err, featureID, nil, msg))
	return v, t
}

func (c *client) Int64Variation(ctx context.Context, featureID string, def int64) (int64, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	const base, bitSize = 10, 64
	v, err := strconv.ParseInt(variation, base, bitSize)
	if err == nil {
		return v, t
	}
	const msg = "Failed to parse int"
	c.log(newError(ctx, err, featureID, nil, msg))
	return v, t
}

func (c *client) Float64Variation(ctx context.Context, featureID string, def float64) (float64, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	const bitSize = 64
	v, err := strconv.ParseFloat(variation, bitSize)
	if err == nil {
		return v, t
	}
	const msg = "Failed to parse float"
	c.log(newError(ctx, err, featureID, nil, msg))
	return v, t
}

func (c *client) StringVariation(ctx context.Context, featureID string, def string) (string, Tracker) {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return def, t
	}
	return variation, t
}

func (c *client) JSONVariation(ctx context.Context, featureID string, dst interface{}) Tracker {
	variation, t, err := c.fetchVariation(ctx, featureID)
	if err != nil {
		return t
	}
	err = json.NewDecoder(strings.NewReader(variation)).Decode(dst)
	if err == nil {
		const msg = "Failed to decode json"
		c.log(newError(ctx, err, featureID, nil, msg))
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

func (c perRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.APIKey,
	}, nil
}

func (c perRPCCredentials) RequireTransportSecurity() bool {
	return true
}
