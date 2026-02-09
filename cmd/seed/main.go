package main

import (
	"encoding/json"
	"os"

	"github.com/crowemi-io/crowemi-trades/internal/config"
)

func main() {
	config, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	config.Logger.Log("msg", "start crowemi-trades-seed")
	contents, err := os.ReadFile("seed.json")
	if err != nil {
		panic(err)
	}

	type Seed struct {
	}

	json.Marshal()

}
