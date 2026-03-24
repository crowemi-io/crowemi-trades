package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/crowemi-io/crowemi-trades/internal/db"
	"github.com/crowemi-io/crowemi-trades/internal/models"
)

const accountLogPath = "account.process"

func (h *Handler) ProcessAccount(w http.ResponseWriter, r *http.Request) {
	sessionID := newSessionID()
	_ = h.Logger.Log("msg", accountLogPath+" start", "path", accountLogPath, "session_id", sessionID)

	account, err := h.Alpaca.Client.GetAccount()
	if err != nil {
		_ = h.Logger.Log("msg", accountLogPath+" error", "path", accountLogPath, "session_id", sessionID, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accountDoc := models.AccountFromAlpaca(account)
	if accountDoc == nil {
		_ = h.Logger.Log("msg", accountLogPath+" error: nil account", "path", accountLogPath, "session_id", sessionID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = db.Create[*models.Account](r.Context(), h.FirestoreDB, db.CollectionAccounts, accountDoc)
	if err != nil {
		_ = h.Logger.Log("msg", accountLogPath+" error", "path", accountLogPath, "session_id", sessionID, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "session_id_generation_error"
	}
	return hex.EncodeToString(b[:])
}
