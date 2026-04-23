package api

import (
	ct "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	kitlog "github.com/go-kit/log"
)

type Handler struct {
	Logger   kitlog.Logger
	Queries  *sqlc.Queries
	Alpaca   *ct.Alpaca
	Notifier *notifier.Notifier
}
