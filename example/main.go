package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/uuid"
)

const (
	timeout = 10 * time.Second

	userIDKey = "user_id"
)

var (
	bucketeerTag            = flag.String("bucketeer-tag", "", "Bucketeer tag")
	bucketeerAPIKey         = flag.String("bucketeer-api-key", "", "Bucketeer api key")
	bucketeerHost           = flag.String("bucketeer-host", "", "Bucketeer host name, e.g. api-dev.bucketeer.jp")
	bucketeerPort           = flag.Int("bucketeer-port", 443, "Bucketeer port number, e.g. 443")
	bucketeerEnableDebugLog = flag.Bool("bucketeer-enable-debug-log", false, "Outpus sdk debug logs or not")
	bucketerFeatureID       = flag.String("bucketeer-feature-id", "", "Target feature id")
	bucketerGoalID          = flag.String("bucketeer-goal-id", "", "Target goal id")
)

func main() {
	flag.Parse()

	// Setup Bucketeer SDK
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sdk, err := bucketeer.NewSDK(
		ctx,
		bucketeer.WithTag(*bucketeerTag),
		bucketeer.WithAPIKey(*bucketeerAPIKey),
		bucketeer.WithHost(*bucketeerHost),
		bucketeer.WithPort(*bucketeerPort),
		bucketeer.WithEnableDebugLog(*bucketeerEnableDebugLog),
	)
	if err != nil {
		log.Fatalf("failed to new sdk: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sdk.Close(ctx); err != nil {
			log.Fatalf("failed to close sdk: %v", err)
		}
	}()

	// Setup http server
	app := exampleApp{
		sdk:       sdk,
		featureID: *bucketerFeatureID,
		goalID:    *bucketerGoalID,
	}
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      app.rootHandler(),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	// Serve
	srvShutdownCh := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Fatalf("failed to shutdown http server: %v", err)
		}
		close(srvShutdownCh)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("failed to close ListenAndServe: %v", err)
	}
	<-srvShutdownCh
}

type exampleApp struct {
	sdk       bucketeer.SDK
	featureID string
	goalID    string
}

func (a *exampleApp) rootHandler() http.Handler {
	mux := http.NewServeMux()
	// GET /variation
	mux.HandleFunc("/variation", a.getVariationHandler)
	// POST /track
	mux.HandleFunc("/track", a.postTrackHandler)
	return mux
}

func (a *exampleApp) getVariationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, err := a.userFrom(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	variation := a.sdk.StringVariation(ctx, user, a.featureID, "default")
	fmt.Fprint(w, variation)
}

func (a *exampleApp) postTrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, err := a.userFrom(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	a.sdk.Track(ctx, user, a.goalID)
	fmt.Fprint(w, "ok")
}

func (a *exampleApp) userFrom(r *http.Request) (*bucketeer.User, error) {
	var userID string
	cookie, err := r.Cookie(userIDKey)
	if err == nil {
		userID = cookie.Value
	} else {
		uuid, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		userID = uuid.String()
	}
	return bucketeer.NewUser(userID, nil), nil
}
