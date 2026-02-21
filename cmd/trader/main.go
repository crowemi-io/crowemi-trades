package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/api"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/runtime"
	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	"github.com/crowemi-io/crowemi-trades/internal/stream"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	firestoreDB := db.NewFirestore(c.Firestore)

	c.Logger.Log("msg", "start crowemi-trades")

	httpClient := &http.Client{}

	alpacaClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:     c.Alpaca.APIKey,
		APISecret:  c.Alpaca.APISecretKey,
		BaseURL:    c.Alpaca.APIBaseURL,
		HTTPClient: httpClient,
	})

	handler := &api.Handler{
		Logger:      c.Logger,
		FirestoreDB: firestoreDB,
		Alpaca:      alpacaClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /account/process", handler.ProcessAccount)
	server := &http.Server{
		Addr:    c.Runtime.HTTPListenAddr,
		Handler: mux,
	}

	accountSyncInterval, err := time.ParseDuration(c.Runtime.AccountSyncInterval)
	if err != nil {
		log.Fatal(err)
	}
	taskTimeout, err := time.ParseDuration(c.Runtime.TaskTimeout)
	if err != nil {
		log.Fatal(err)
	}
	streamReconnectMin, err := time.ParseDuration(c.Runtime.StreamReconnectMin)
	if err != nil {
		log.Fatal(err)
	}
	streamReconnectMax, err := time.ParseDuration(c.Runtime.StreamReconnectMax)
	if err != nil {
		log.Fatal(err)
	}

	schedulerRunner := &scheduler.Runner{
		Logger: c.Logger,
		Tasks: []scheduler.Task{
			&scheduler.AccountSyncTask{
				AlpacaClient: alpacaClient,
				FirestoreDB:  firestoreDB,
				Every:        accountSyncInterval,
			},
		},
		TaskTimeout: taskTimeout,
	}
	streamConsumer := &stream.TradeUpdatesConsumer{
		Logger:       c.Logger,
		Streamer:     alpacaClient,
		ReconnectMin: streamReconnectMin,
		ReconnectMax: streamReconnectMax,
	}

	app := &runtime.App{
		Logger:    c.Logger,
		Server:    server,
		Scheduler: schedulerRunner,
		Stream:    streamConsumer,
	}
	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
