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
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	"github.com/crowemi-io/crowemi-trades/internal/runtime"
	"github.com/crowemi-io/crowemi-trades/internal/scheduler"
	task "github.com/crowemi-io/crowemi-trades/internal/scheduler/tasks"
	api "github.com/crowemi-io/crowemi-trades/internal/server"
	"github.com/crowemi-io/crowemi-trades/internal/stream"

	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/go-kit/log/level"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	postgresDB, err := db.NewPostgres(ctx, c.Crowemi.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

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
			Logger:  c.Logger,
			Queries: postgresDB.Queries,
			Alpaca:  alpaca,
		}
		mux := http.NewServeMux()
		mux.HandleFunc("POST /", handler.ProcessAccount)
		ser = &http.Server{
			Handler: mux,
			Addr:    ":8080",
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
			_ = level.Debug(c.Logger).Log("msg", "loading task from config", "name", t.Name, "schedule", t.Schedule, "enabled", t.Enabled)
			switch t.Name {
			case "account":
				sch.Tasks = append(sch.Tasks, &task.AccountTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "activity":
				sch.Tasks = append(sch.Tasks, &task.ActivityTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "order":
				sch.Tasks = append(sch.Tasks, &task.OrderTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "position":
				sch.Tasks = append(sch.Tasks, &task.PositionTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "corporate_action":
				sch.Tasks = append(sch.Tasks, &task.CorporateActionTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
				})
			case "rebalance":
				sch.Tasks = append(sch.Tasks, &task.RebalanceTask{
					Config:       c,
					Alpaca:       alpaca,
					Queries:      postgresDB.Queries,
					Logger:       c.Logger,
					CronSchedule: t.Schedule,
					Options:      t.Options,
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
		// get the symbols associated with app category from portfolio
		// First get the account to find the account_id
		account, err := postgresDB.Queries.GetAccountByAlpacaID(ctx, &c.Alpaca.AccountID)
		if err != nil {
			c.Logger.Log("msg", "failed to load account for watcher", "err", err, "AccountID", c.Alpaca.AccountID)
		} else {
			// Get portfolio symbols for this account
			symbols, err := postgresDB.Queries.ListPortfolioSymbolsByAccountID(ctx, account.ID)
			if err != nil {
				c.Logger.Log("msg", "failed to load portfolio for watcher", "err", err, "AccountID", c.Alpaca.AccountID)
			}
			if err != nil {
				c.Logger.Log("msg", "failed to load portfolio symbols for watcher", "err", err, "AccountID", c.Alpaca.AccountID)
			} else if len(symbols) > 0 {
				symbolList := make([]string, len(symbols))
				for i, s := range symbols {
					symbolList[i] = s
				}
				wat = &stream.Watcher{
					Logger:         c.Logger,
					Symbols:        symbolList,
					APIKey:         c.Alpaca.APIKey,
					APISecret:      c.Alpaca.APISecretKey,
					DataURL:        c.Alpaca.APIDataURL,
					MarketDataFeed: c.Alpaca.MarketDataFeed,
				}
			}
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
