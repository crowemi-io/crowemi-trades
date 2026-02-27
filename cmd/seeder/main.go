package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/uuid"

	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

func main() {
	cfg, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}
	firestoreDB := db.NewFirestore(cfg.Firestore)

	cfg.Logger.Log("msg", "start crowemi-trades-seed")

	contents, err := os.ReadFile("seed.json")
	if err != nil {
		panic(err)
	}

	var portfolio models.Portfolio
	if err := json.Unmarshal(contents, &portfolio); err != nil {
		panic(err)
	}

	portfolio.ID = uuid.New().String()

	ctx := context.Background()
	id, err := db.Create(ctx, firestoreDB, db.CollectionPortfolios, &portfolio)
	if err != nil {
		cfg.Logger.Log("msg", "failed to seed portfolio", "err", err)
		panic(err)
	}

	cfg.Logger.Log("msg", "seeded portfolio", "id", id)
	cfg.Logger.Log("msg", "completed crowemi-trades-seed")
}
