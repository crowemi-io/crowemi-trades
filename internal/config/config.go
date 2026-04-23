package config

import (
	"encoding/base64"
	"encoding/json"
	"os"

	kitlog "github.com/go-kit/log"
)

type GoogleCloudFirestore struct {
	Database string `json:"database" omitempty:"true"`
}

type GoogleCloud struct {
	OrganizationID string               `json:"organization_id" omitempty:"true"`
	ProjectID      string               `json:"project_id" omitempty:"true"`
	Region         string               `json:"region" omitempty:"true"`
	Firestore      GoogleCloudFirestore `json:"firestore" omitempty:"true"`
}

type Alpaca struct {
	AccountID      string `json:"account_id" omitempty:"true"`
	APIKey         string `json:"api_key" omitempty:"true"`
	APISecretKey   string `json:"api_secret_key" omitempty:"true"`
	APIBaseURL     string `json:"api_base_url" omitempty:"true"`
	APIDataURL     string `json:"api_data_url" omitempty:"true"`
	MarketDataFeed string `json:"market_data_feed" omitempty:"true"`
}

type Crowemi struct {
	ClientName      string            `json:"client_name" omitempty:"true"`
	ClientID        string            `json:"client_id" omitempty:"true"`
	ClientSecretKey string            `json:"client_secret_key" omitempty:"true"`
	Uri             map[string]string `json:"uri" omitempty:"true"`
	DatabaseURI     string            `json:"database_uri" omitempty:"true"`
}

type Runtime struct {
	Server    Server    `json:"server" omitempty:"true"`
	Scheduler Scheduler `json:"scheduler" omitempty:"true"`
	Streamer  Streamer  `json:"streamer" omitempty:"true"`
}

type Notifier struct {
	Telegram *Telegram `json:"telegram" omitempty:"true"`
}

type Telegram struct {
	BotToken string `json:"bot_token" omitempty:"true"`
	ChatID   int64  `json:"chat_id" omitempty:"true"`
}

type Config struct {
	Alpaca      Alpaca      `json:"alpaca"`
	Crowemi     Crowemi     `json:"crowemi"`
	GoogleCloud GoogleCloud `json:"google_cloud"`
	Notifier    Notifier    `json:"notifier" omitempty:"true"`
	Runtime     Runtime     `json:"runtime"`
	Logger      kitlog.Logger
}

func (c *Config) RootCollection() string {
	return "accounts/" + c.Alpaca.AccountID + "/"
}

func Bootstrap(configPath string) (*Config, error) {
	var config Config
	value := os.Getenv("CONFIG")
	if value != "" {
		decode, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(decode, &config)
	} else {
		contents, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(contents, &config)
	}
	// logger
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", kitlog.DefaultCaller)
	config.Logger = logger
	return &config, nil
}
