package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	"github.com/crowemi-io/crowemi-trades/internal/runtime"
	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	task "github.com/crowemi-io/crowemi-trades/internal/scheduler/tasks"
	api "github.com/crowemi-io/crowemi-trades/internal/server"
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

	var n *notifier.Notifier = nil
	if c.Notifier.Telegram != nil && c.Notifier.Telegram.BotToken != "" && c.Notifier.Telegram.ChatID != 0 {
		n, err = notifier.New(notifier.Config{
			BotToken: c.Notifier.Telegram.BotToken,
			ChatID:   c.Notifier.Telegram.ChatID,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	httpClient := &http.Client{}
	alpaca := &ct.Alpaca{
		Client: alpaca.NewClient(alpaca.ClientOpts{
			APIKey:     c.Alpaca.APIKey,
			APISecret:  c.Alpaca.APISecretKey,
			BaseURL:    c.Alpaca.APIBaseURL,
			HTTPClient: httpClient,
		}),
		Notifier: n,
	}
	// server init
	var ser *http.Server = nil
	if c.Runtime.Server.Enabled {
		handler := &api.Handler{
			Logger:      c.Logger,
			FirestoreDB: firestoreDB,
			Alpaca:      alpaca,
		}
		mux := http.NewServeMux()
		mux.HandleFunc("POST /account/process", handler.ProcessAccount)
		ser = &http.Server{
			Handler: mux,
		}
	}
	// scheduler init
	var sch *scheduler.Runner = nil
	if c.Runtime.Scheduler.Enabled {
		sch = &scheduler.Runner{}
		for _, t := range c.Runtime.Scheduler.Tasks {
			if !t.Enabled {
				continue
			}
			switch t.Name {
			case "account":
				sch.Tasks = append(sch.Tasks, &task.AccountTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "activity":
				sch.Tasks = append(sch.Tasks, &task.ActivityTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "order":
				sch.Tasks = append(sch.Tasks, &task.OrderTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "position":
				sch.Tasks = append(sch.Tasks, &task.PositionTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "corporate_action":
				sch.Tasks = append(sch.Tasks, &task.CorporateActionTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "rebalance":
				sch.Tasks = append(sch.Tasks, &task.RebalanceTask{
					Alpaca:       alpaca,
					FirestoreDB:  firestoreDB,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			default:
				continue
			}
		}
		sch.Logger = c.Logger
	}
	// stream updater init
	var upd *stream.Updater = nil
	if c.Runtime.Streamer.Updater.Enabled {
		upd = &stream.Updater{
			Logger: c.Logger,
			Alpaca: alpaca,
		}
	}
	// stream watcher init
	var wat *stream.Watcher = nil
	if c.Runtime.Streamer.Watcher.Enabled {
		portfolio, err := db.Get[*models.Portfolio](ctx, firestoreDB, db.CollectionPortfolios, c.Runtime.PortfolioID)
		if err != nil {
			c.Logger.Log("msg", "failed to load portfolio for watcher", "err", err, "portfolio_id", c.Runtime.PortfolioID)
		}
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
				wat = &stream.Watcher{
					Logger:         c.Logger,
					Symbols:        symbols,
					APIKey:         c.Alpaca.APIKey,
					APISecret:      c.Alpaca.APISecretKey,
					DataURL:        c.Alpaca.APIDataURL,
					MarketDataFeed: c.Alpaca.MarketDataFeed,
				}
			}
		} else if err != nil {
			c.Logger.Log("msg", "failed to load portfolio for minute bars", "err", err, "portfolio_id", c.Runtime.PortfolioID)
		}
	}

	app := &runtime.App{
		Logger:    c.Logger,
		Server:    ser,
		Scheduler: sch,
		Watcher:   wat,
		Updater:   upd,
	}

	if err := app.Run(ctx); err != nil {
		_ = n.Notify(ctx, "Error running app: "+err.Error())
		log.Fatal(err)
	}
}
