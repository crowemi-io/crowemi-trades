package main

import (
	"log"
	"net/http"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/config"
)

func main() {
	c, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatal(err)
	}

	c.Logger.Log("msg", "start crowemi-trades")

	httpClient := &http.Client{}

	alpacaClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:     c.Alpaca.APIKey,
		APISecret:  c.Alpaca.APISecretKey,
		BaseURL:    c.Alpaca.APIBaseURL,
		HTTPClient: httpClient,
	})

	account, err := alpacaClient.GetAccount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Account: %+v\n", account)

	// rebalance portfolio

}
