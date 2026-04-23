package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/crowemi-io/crowemi-trades/internal/config"
	"github.com/crowemi-io/crowemi-trades/internal/db"
)

type SeedData map[string]Category

type Category struct {
	Rebalance  bool              `json:"rebalance"`
	Percentage float64           `json:"percentage"`
	Symbols    map[string]Symbol `json:"symbols"`
}

type Symbol struct {
	Weight float64 `json:"weight"`
}

func main() {
	cfg, err := config.Bootstrap(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}
	postgres, err := db.NewPostgres(context.TODO(), cfg.Crowemi.DatabaseURI)
	if err != nil {
		panic(err)
	}
	defer postgres.Close()

	cfg.Logger.Log("msg", "start crowemi-trades-seed")

	addAccount := `
    INSERT INTO app.account (name, account_number, created_at, created_by, updated_at, updated_by)
    VALUES ($1, $2, NOW(), $3, NOW(), $4)
    ON CONFLICT (account_number) DO UPDATE SET name = $1, account_number = $2, updated_at = NOW(), updated_by = $3
    RETURNING id;
	`

	var accountID int64
	err = postgres.Pool.QueryRow(context.TODO(), addAccount,
		cfg.Crowemi.ClientName, cfg.Alpaca.AccountID, "system", "system",
	).Scan(&accountID)

	if err != nil {
		panic(err)
	}

	contents, err := os.ReadFile("seed.json")
	if err != nil {
		panic(err)
	}

	var seedData SeedData
	if err := json.Unmarshal(contents, &seedData); err != nil {
		panic(err)
	}

	for categoryName, category := range seedData {

		addPortfolio := `
		INSERT INTO app.portfolio (account_id, portfolio_key, name,	weight,	created_at,	created_by,	updated_at,	updated_by)
		VALUES ($1,	$2,	$3,	$4,	$5,	$6,	$5,	$6)
		ON CONFLICT (account_id, name) DO UPDATE 
		SET account_id = $1, portfolio_key = $2, name = $3, weight = $4, updated_at = $5, updated_by = $6
    	RETURNING id;`

		var portfolioID int64
		err := postgres.Pool.QueryRow(context.TODO(), addPortfolio,
			accountID,
			nil,
			categoryName,
			category.Percentage,
			time.Now().UTC(),
			"system",
		).Scan(&portfolioID)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Category: %s, %f%%\n", categoryName, category.Percentage)
		for symbolName, symbol := range category.Symbols {
			addSymbol := `
			INSERT INTO app.portfolio_symbol (portfolio_id, symbol, weight, created_at, created_by, updated_at, updated_by)
			VALUES ($1,	$2,	$3,	$4,	$5,	$4,	$5)
			ON CONFLICT (portfolio_id, symbol) DO UPDATE 
			SET portfolio_id = $1, symbol = $2, weight = $3, updated_at = $4, updated_by = $5
			RETURNING id;`
			postgres.Pool.Exec(context.TODO(), addSymbol,
				portfolioID,
				symbolName,
				symbol.Weight,
				time.Now().UTC(),
				"system",
			)
			fmt.Printf("  %s: %f%%\n", symbolName, symbol.Weight)
		}
	}

	cfg.Logger.Log("msg", "completed crowemi-trades-seed")
}
