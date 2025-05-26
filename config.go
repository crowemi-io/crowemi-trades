package trader

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-go-utils/cloud"
	"github.com/crowemi-io/crowemi-go-utils/config"
	"github.com/crowemi-io/crowemi-go-utils/db"
	"github.com/crowemi-io/crowemi-go-utils/log"
)

type Config struct {
	Alpaca       config.Alpaca `json:"alpaca"`
	AlpacaClient *Alpaca
	Crowemi      config.Crowemi `json:"crowemi"`
	MongoClient  *db.MongoClient
	GoogleCloud  *config.GoogleCloud `json:"google_cloud"`
	GcpClient    *cloud.GcpClient
	Logger       *log.Logger
}

func Bootstrap() (*Config, error) {
	c, err := config.Bootstrap[Config]("")
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

	gcpClient, err := cloud.NewGcpClient(c.Crowemi.ClientName, c.GoogleCloud)
	if err != nil {
		return nil, err
	}
	c.GcpClient = gcpClient
	c.Logger = &log.Logger{GcpClient: c.GcpClient}

	return c, nil
}
