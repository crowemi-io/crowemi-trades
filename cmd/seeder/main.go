package main

import (
	"context"
	"encoding/json"
	"os"

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

	var allocations map[string]models.Allocation
	if err := json.Unmarshal(contents, &allocations); err != nil {
		panic(err)
	}

	account := models.Account{
		ID: cfg.Alpaca.AccountID,
	}

	for categoryName, allocation := range allocations {
		category := models.NewCategory(categoryName, allocation.Rebalance, allocation.Percentage)

		act := &models.Account{
			ID: account.ID,
		}
		account := firestoreDB.Client.Doc("accounts/" + account.ID)
		account.Set(context.TODO(), &act)
		account.Collection("allocations").Doc(categoryName).Set(context.TODO(), &category)

		cfg.Logger.Log("msg", "seeded category", "category", categoryName)

		for ID, d := range allocation.Symbols {
			symbol := models.NewSymbol(ID, d.Weight)
			firestoreDB.Client.Doc("accounts/"+account.ID).Collection("allocations").Doc(categoryName).Collection("symbols").Doc(symbol.ID).Set(context.TODO(), &symbol)
			cfg.Logger.Log("msg", "seeded symbol", "symbol", symbol.ID)
		}
	}

	cfg.Logger.Log("msg", "completed crowemi-trades-seed")
}
