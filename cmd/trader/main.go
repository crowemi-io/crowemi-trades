package main

import (
	"net/http"

	"github.com/crowemi-io/crowemi-go-utils/cloud"
	trader "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/api"
)

func main() {

	config, err := trader.Bootstrap()
	if err != nil {
		// TODO: clean logging
		println("error")
	}
	config.GoogleCloudClient.Log("crowemi-trades start", cloud.INFO, nil, "main.main")

	// TODO: create go routine for trade updates
	// TODO: create go routine for market dater

	// http handlers
	http.HandleFunc("/v1/portfolio/", func(w http.ResponseWriter, r *http.Request) {
		h := api.PortfolioHandler{TraderConfig: config}
		h.Handler(w, r)
	})
	http.ListenAndServe(":8004", nil)
}
