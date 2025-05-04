package config

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/crowemi-io/crowemi-go-utils/config"
)

type Config struct {
	Alpaca  config.Alpaca  `json:"alpaca"`
	Crowemi config.Crowemi `json:"crowemi"`
}

func Bootstrap() (*Config, error) {

	config := &Config{
		Alpaca:  config.Alpaca{},
		Crowemi: config.Crowemi{},
	}

	value := os.Getenv("CONFIG")
	if value != "" {
		decode, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(decode, &config)
	} else {
		contents, err := os.ReadFile("../.secret/config-local.json")
		if err != nil {
			return nil, err
		}
		json.Unmarshal(contents, &config)
	}
	return config, nil
}
