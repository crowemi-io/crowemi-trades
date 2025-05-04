package main

import (
	"context"
	"log"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/config"
)

func main() {
	c, err := config.Bootstrap()
	if err != nil {
		println(err)
	}
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    c.Alpaca.APIKey,
		APISecret: c.Alpaca.APISecretKey,
		BaseURL:   c.Alpaca.APIBaseURL,
	})

	// Get latest activities -> BigQuery
	// Check free capital
	// Rebalance portfolio
	//

	client.StreamTradeUpdatesInBackground(context.Background(), func(tu alpaca.TradeUpdate) {
		log.Printf("TRADE UPDATE: %+v\n", tu)
		// option
		// stock

		// log to bigquery ?
	})

	// http handlers
	select {}
}
