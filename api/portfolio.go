package api

import (
	"net/http"

	"github.com/crowemi-io/crowemi-go-utils/db"
	trader "github.com/crowemi-io/crowemi-trades"
)

type PortfolioHandler struct {
	TraderConfig *trader.Config
}

// implements Handler
func (h *PortfolioHandler) Handler(w http.ResponseWriter, r *http.Request) {}
func (h *PortfolioHandler) GetConfig() *trader.Config {
	return h.TraderConfig
}
func (h *PortfolioHandler) SetConfig(config *trader.Config) {
	h.TraderConfig = config
}

func (p *PortfolioHandler) Rebalance() {
	// Check free capital
	// free capital = cash - (total allowed invested capital - total invested capital)
	// get cash
	cash, err := p.TraderConfig.AlpacaClient.GetCash()
	if err != nil {
		println(err)
	}

	freeCapital := cash

	// get total allowed invested capital
	watchlists, err := trader.GetWatchlists(p.TraderConfig.MongoClient)
	if err != nil {
		println(err)
	}
	f := []db.MongoFilter{
		{Field: "sell_at_utc", Operator: "$eq", Value: nil},
	}
	openOrders, err := trader.GetOrders(p.TraderConfig.MongoClient, f)
	if err != nil {
		println(err)
	}
	outstandingCapital := trader.GetOutstandingCapital(watchlists, openOrders)
	freeCapital -= outstandingCapital

	print(freeCapital)

	// total cost basis + free capital * percentage - current allocation
	portfolio, err := trader.GetPortfolio(p.TraderConfig.MongoClient, p.TraderConfig.AlpacaClient, nil, true)
	if err != nil {
		println(err)
	}
	for _, port := range portfolio {
		trader.Rebalance(p.TraderConfig.AlpacaClient.Client, &freeCapital, &port)
	}

}
