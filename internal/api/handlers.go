package api

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/crowemi-io/crowemi-trades/internal/db"
	kitlog "github.com/go-kit/log"
)

type Handler struct {
	Logger      kitlog.Logger
	FirestoreDB *db.Firestore
	Alpaca      *alpaca.Client
}
