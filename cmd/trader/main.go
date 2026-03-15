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
	// Load portfolio symbols for minute bars subscription
	symbolMgr := stream.NewSymbolManager(c.Logger)
	if err := symbolMgr.LoadSymbols(ctx, &stream.FirestorePortfolioGetter{FS: firestoreDB}, c.Runtime.PortfolioID, "app"); err != nil {
		c.Logger.Log("msg", "failed to load portfolio symbols", "err", err, "portfolio_id", c.Runtime.PortfolioID)
		symbolMgr = nil
	}

	var streamRunner interface{ Run(context.Context) error }
	if symbolMgr != nil {
		symbols := symbolMgr.GetSymbols()
		if len(symbols) > 0 {
			streamRunner = &stream.MinuteBarsConsumer{
				Logger:    c.Logger,
				Symbols:   symbols,
				APIKey:    c.Alpaca.APIKey,
				APISecret: c.Alpaca.APISecretKey,
				DataURL:   c.Alpaca.APIDataURL,
			}
		}
		defer func() {
			if symbolMgr != nil {
				symbolMgr.Shutdown()
			}
		}()
	}

	app := &runtime.App{
		Logger:    c.Logger,
		Server:    server,
		Scheduler: schedulerRunner,
		Stream:    streamRunner,
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
