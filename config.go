package trader

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-go-utils/config"
	"github.com/crowemi-io/crowemi-go-utils/db"
)

type Config struct {
	Alpaca       config.Alpaca `json:"alpaca"`
	AlpacaClient *Alpaca
	Crowemi      config.Crowemi `json:"crowemi"`
	MongoClient  *db.MongoClient
}

func Bootstrap() (*Config, error) {
	c, err := config.Bootstrap[Config]()
	if err != nil {
		println(err)
	}
	c.AlpacaClient = &Alpaca{Client: alpaca.NewClient(alpaca.ClientOpts{
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
	c.MongoClient = &mongoClient
	return c, nil
}
