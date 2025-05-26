package api

import (
	"net/http"

	trader "github.com/crowemi-io/crowemi-trades"
)

type Handler interface {
	// Handler(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
	Put(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func HandlerFactory(path string, config *trader.Config) Handler {
	switch path {
	case "activities":
		return ActivityHandler{TraderConfig: config}
	case "portfolio":
		return PortfolioHandler{TraderConfig: config}
	default:
		return nil
	}
}
