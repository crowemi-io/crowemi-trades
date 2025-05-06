package main

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-go-utils/config"
	"github.com/crowemi-io/crowemi-go-utils/db"
	ct "github.com/crowemi-io/crowemi-trades"
)

type Config struct {
	Alpaca  config.Alpaca  `json:"alpaca"`
	Crowemi config.Crowemi `json:"crowemi"`
}

func main() {
	c, err := config.Bootstrap[Config]() // config.Bootstrap()
	if err != nil {
		println(err)
	}
	alpacaClient := ct.Alpaca{Config: c.Alpaca, Client: alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    c.Alpaca.APIKey,
		APISecret: c.Alpaca.APISecretKey,
		BaseURL:   c.Alpaca.APIBaseURL,
	})}

	mongoClient := db.MongoClient{}
	mongoClient.Connect(context.TODO(), c.Crowemi.DatabaseURI, c.Crowemi.ClientName)
	err = mongoClient.Ping()
	if err != nil {
		println(err)
	}

	// Check free capital
	// free capital = cash - (total allowed invested capital - total invested capital)
	// get cash
	cash, err := alpacaClient.GetCash()
	if err != nil {
		println(err)
	}

	freeCapital := cash

	// get total allowed invested capital
	watchlists, err := ct.GetWatchlists(&mongoClient)
	if err != nil {
		println(err)
	}
	for _, w := range watchlists {
		if w.IsActive && !w.IsSuspended {
			if w.Type == "stock" {
				f := []db.MongoFilter{
					{Field: "symbol", Operator: "$eq", Value: w.Symbol},
					{Field: "sell_at_utc", Operator: "$eq", Value: nil},
				}
				openOrders, err := ct.GetOrders(&mongoClient, f)
				if err != nil {
					println(err)
				}
				// when the total number of open orders is less than the allowed batches
				// we need to account for the potential outstanding orders
				if len(openOrders) < w.AllowedBatches {
					outstanding := float64(w.BatchSize * (w.AllowedBatches - len(openOrders)))
					// then subtract the outstanding orders from the free capital
					freeCapital = freeCapital - outstanding
				}

			}
		}
	}

	print(freeCapital)

	// total invested capital
	alpacaClient.GetPositions()

	// Rebalance portfolio
	//

	// client.StreamTradeUpdatesInBackground(context.Background(), func(tu alpaca.TradeUpdate) {
	// 	log.Printf("TRADE UPDATE: %+v\n", tu)
	// 	// option
	// 	// stock

	// 	// log to bigquery ?
	// })

	// http handlers
	select {}
}
