package main

import (
	"net/http"

	trader "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/api"
)

func main() {

	config, err := trader.Bootstrap()
	if err != nil {
		// TODO: clean logging
		println("error")
	}

	// TODO: create go routine for trade updates
	// TODO: create go routine for market dater

	p := api.PortfolioHandler{TraderConfig: config}

	// http handlers
	http.HandleFunc("/v1/portfolio/", func(w http.ResponseWriter, r *http.Request) { p.Handler(w, r) })
	http.ListenAndServe(":8004", nil)
}
