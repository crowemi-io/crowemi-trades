package api

import (
	"net/http"
)

const accountLogPath = "account.process"

func (h *Handler) ProcessAccount(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
