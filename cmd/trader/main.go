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
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	"github.com/crowemi-io/crowemi-trades/internal/notifier/telegram"
	"github.com/crowemi-io/crowemi-trades/internal/runtime"
	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	task "github.com/crowemi-io/crowemi-trades/internal/scheduler/tasks"
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

	if c.Notifier.Telegram != nil && c.Notifier.Telegram.BotToken != "" && c.Notifier.Telegram.ChatID != 0 {
		tg, err := telegram.New(telegram.Config{
			BotToken: c.Notifier.Telegram.BotToken,
			ChatID:   c.Notifier.Telegram.ChatID,
		})
		if err != nil {
			log.Fatal(err)
		}
		handler.Notifier = notifier.NewMulti(tg)
		if err := handler.Notifier.Notify(ctx, "crowemi-trades started"); err != nil {
			c.Logger.Log("msg", "start notification failed", "err", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /account/process", handler.ProcessAccount)
	server := &http.Server{
		Addr:    c.Runtime.HTTPListenAddr,
		Handler: mux,
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

	accountSyncTask := &task.AccountSyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
	}
	accountSyncTask.CronSchedule = c.Runtime.Scheduler.ScheduleForTask(accountSyncTask.Name())
	activitySyncTask := &task.ActivitySyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
	}
	activitySyncTask.CronSchedule = c.Runtime.Scheduler.ScheduleForTask(activitySyncTask.Name())
	orderSyncTask := &task.OrderSyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
	}
	orderSyncTask.CronSchedule = c.Runtime.Scheduler.ScheduleForTask(orderSyncTask.Name())
	positionSyncTask := &task.PositionSyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
	}
	positionSyncTask.CronSchedule = c.Runtime.Scheduler.ScheduleForTask(positionSyncTask.Name())
	corporateActionSyncTask := &task.CorporateActionSyncTask{
		AlpacaClient: alpacaClient,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
	}
	corporateActionSyncTask.CronSchedule = c.Runtime.Scheduler.ScheduleForTask(corporateActionSyncTask.Name())

	schedulerRunner := &scheduler.Runner{
		Logger: c.Logger,
		Tasks: []scheduler.Task{
			accountSyncTask,
			activitySyncTask,
			orderSyncTask,
			positionSyncTask,
			corporateActionSyncTask,
		},
		TaskTimeout: taskTimeout,
	}
	// Initialize SymbolManager for per-symbol websocket routing
	symbolMgr := stream.NewSymbolManager(c.Logger)

	// Load portfolio symbols from Firestore (use "default" or read from config)
	portfolioID := "default" // TODO: make configurable via config
	if err := symbolMgr.LoadSymbols(ctx, firestoreDB.Client, portfolioID); err != nil {
		c.Logger.Log("msg", "failed to load portfolio symbols", "err", err, "portfolio_id", portfolioID)
		// Continue without per-symbol routing if loading fails
		symbolMgr = nil
	}

	streamConsumer := &stream.TradeUpdatesConsumer{
		Logger:       c.Logger,
		Streamer:     alpacaClient,
		ReconnectMin: streamReconnectMin,
		ReconnectMax: streamReconnectMax,
		SymbolMgr:    symbolMgr, // inject the manager for per-symbol routing
	}

	app := &runtime.App{
		Logger:    c.Logger,
		Server:    server,
		Scheduler: schedulerRunner,
		Stream:    streamConsumer,
	}

	// Ensure SymbolManager is shut down on exit
	defer func() {
		if symbolMgr != nil {
			symbolMgr.Shutdown()
		}
	}()

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
