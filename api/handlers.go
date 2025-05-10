package api

import (
	"net/http"

	trader "github.com/crowemi-io/crowemi-trades"
)

type Handler interface {
	Handler(w http.ResponseWriter, r *http.Request)
	GetConfig() *trader.Config
	SetConfig(config *trader.Config)
}

func GetHandler[T Handler](config *trader.Config) Handler {
	var handler T
	handler.SetConfig(config)
	return handler
}
