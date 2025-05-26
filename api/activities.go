package api

import (
	"net/http"

	trader "github.com/crowemi-io/crowemi-trades"
)

type ActivityHandler struct {
	TraderConfig *trader.Config
}

func (h ActivityHandler) Get(w http.ResponseWriter, r *http.Request)    {}
func (h ActivityHandler) Post(w http.ResponseWriter, r *http.Request)   {}
func (h ActivityHandler) Put(w http.ResponseWriter, r *http.Request)    {}
func (h ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {}
