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
	"github.com/crowemi-io/crowemi-trades/internal/models"
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	"github.com/crowemi-io/crowemi-trades/internal/notifier/telegram"
	"github.com/crowemi-io/crowemi-trades/internal/runtime"
	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	task "github.com/crowemi-io/crowemi-trades/internal/scheduler/tasks"
	"github.com/crowemi-io/crowemi-trades/internal/stream"

	ct "github.com/crowemi-io/crowemi-trades"
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

	// mux := http.NewServeMux()
	// mux.HandleFunc("POST /account/process", handler.ProcessAccount)
	// server := &http.Server{
	// 	Addr:    c.Runtime.HTTPListenAddr,
	// 	Handler: mux,
	// }

	taskTimeout, err := time.ParseDuration(c.Runtime.TaskTimeout)
	if err != nil {
		log.Fatal(err)
	}
	streamReconnectMin := time.Second
	if c.Runtime.StreamReconnectMin != "" {
		if d, err := time.ParseDuration(c.Runtime.StreamReconnectMin); err == nil {
			streamReconnectMin = d
		}
	}
	streamReconnectMax := 30 * time.Second
	if c.Runtime.StreamReconnectMax != "" {
		if d, err := time.ParseDuration(c.Runtime.StreamReconnectMax); err == nil {
			streamReconnectMax = d
		}
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

	alpacaWrapper := &ct.Alpaca{Client: alpacaClient, Notifier: handler.Notifier}
	rebalanceTask := &task.RebalanceTask{
		Alpaca:       alpacaWrapper,
		FirestoreDB:  firestoreDB,
		Logger:       c.Logger,
		CronSchedule: c.Runtime.Scheduler.ScheduleForTask("rebalance"),
	}

	schedulerRunner := &scheduler.Runner{
		Logger: c.Logger,
		Tasks: []scheduler.Task{
			accountSyncTask,
			activitySyncTask,
			orderSyncTask,
			positionSyncTask,
			corporateActionSyncTask,
			rebalanceTask,
		},
		TaskTimeout: taskTimeout,
	}
	tradeUpdatesConsumer := &stream.TradeUpdatesConsumer{
		Logger:       c.Logger,
		Streamer:     alpacaClient,
		ReconnectMin: streamReconnectMin,
		ReconnectMax: streamReconnectMax,
	}

	portfolio, err := db.Get[*models.Portfolio](ctx, firestoreDB, db.CollectionPortfolios, c.Runtime.PortfolioID)
	if err != nil {
		c.Logger.Log("msg", "failed to load portfolio for minute bars", "err", err, "portfolio_id", c.Runtime.PortfolioID)
	}

	var streamRunner interface{ Run(context.Context) error }
	if err == nil && portfolio != nil {
		symbolSet := make(map[string]bool)
		if alloc, ok := portfolio.Allocations["app"]; ok {
			for s := range alloc.Symbols {
				symbolSet[s] = true
			}
		}
		symbols := make([]string, 0, len(symbolSet))
		for s := range symbolSet {
			symbols = append(symbols, s)
		}
		if len(symbols) > 0 {
			streamRunner = &stream.MinuteBarsConsumer{
				Logger:    c.Logger,
				Symbols:   symbols,
				APIKey:    c.Alpaca.APIKey,
				APISecret: c.Alpaca.APISecretKey,
				DataURL:   c.Alpaca.APIDataURL,
			}
		}
	} else if err != nil {
		c.Logger.Log("msg", "failed to load portfolio for minute bars", "err", err, "portfolio_id", c.Runtime.PortfolioID)
	}

	app := &runtime.App{
		Logger:       c.Logger,
		Server:       nil, // TODO: add server
		Scheduler:    schedulerRunner,
		Stream:       streamRunner,
		TradeUpdates: tradeUpdatesConsumer,
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
