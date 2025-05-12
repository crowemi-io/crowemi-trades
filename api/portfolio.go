package api

import (
	"fmt"
	"net/http"

	"github.com/crowemi-io/crowemi-go-utils/cloud"
	"github.com/crowemi-io/crowemi-go-utils/db"
	trader "github.com/crowemi-io/crowemi-trades"
)

type PortfolioHandler struct {
	TraderConfig *trader.Config
}

// implements Handler
func (h *PortfolioHandler) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.RequestURI {
	case "/v1/portfolio/rebalance":
		h.TraderConfig.GoogleCloudClient.Log("request rebalance", cloud.INFO, nil, "api/portfolio.Handler")
		h.Rebalance()
	}
	w.Write([]byte("Hello World"))
}

func (p *PortfolioHandler) Rebalance() {
	// Check free capital
	// free capital = cash - (total allowed invested capital - total invested capital)
	// get cash
	freeCapital, err := p.TraderConfig.AlpacaClient.GetCash()
	if err != nil {
		p.TraderConfig.GoogleCloudClient.Log(fmt.Sprintf("error calling GetCash: %e", err), cloud.ERROR, nil, "api/portfolio.Rebalance")
		println(err)
	}
	// get total allowed invested capital
	watchlists, err := trader.GetWatchlists(p.TraderConfig.MongoClient)
	if err != nil {
		p.TraderConfig.GoogleCloudClient.Log(fmt.Sprintf("error calling GetWatchlists: %e", err), cloud.ERROR, nil, "api/portfolio.Rebalance")
		println(err)
	}
	f := []db.MongoFilter{
		{Field: "sell_at_utc", Operator: "$eq", Value: nil},
	}
	// get the total outstanding orders to remove from free capital total
	openOrders, err := trader.GetOrders(p.TraderConfig.MongoClient, f)
	if err != nil {
		p.TraderConfig.GoogleCloudClient.Log(fmt.Sprintf("error calling GetOrders: %e", err), cloud.ERROR, nil, "api/portfolio.Rebalance")
		println(err)
	}
	outstandingCapital := trader.GetOutstandingCapital(watchlists, openOrders)
	freeCapital -= outstandingCapital
	p.TraderConfig.GoogleCloudClient.Log(fmt.Sprintf("Free capital: %f", freeCapital), cloud.INFO, nil, "api/portfolio.Rebalance")

	// total cost basis + free capital * percentage - current allocation
	portfolio, err := trader.GetPortfolio(p.TraderConfig.MongoClient, p.TraderConfig.AlpacaClient, nil, true)
	if err != nil {
		p.TraderConfig.GoogleCloudClient.Log(fmt.Sprintf("error calling GetPortfolio: %e", err), cloud.ERROR, nil, "api/portfolio.Rebalance")
		println(err)
	}
	for _, port := range portfolio {
		trader.Rebalance(p.TraderConfig.AlpacaClient.Client, &freeCapital, &port)
	}

}
