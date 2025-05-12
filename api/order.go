package api

import (
	"net/http"

	trader "github.com/crowemi-io/crowemi-trades"
)

type OrderHandler struct {
	TradeConfig *trader.Config
}

func (o *OrderHandler) Handler(w http.ResponseWriter, r *http.Request) {}
func (o *OrderHandler) GetConfig() *trader.Config                      { return nil }
func (o *OrderHandler) SetConfig(config *trader.Config)                {}

// do stuffs
