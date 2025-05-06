package crowemi_trades

import (
	"context"
	"time"

	"github.com/crowemi-io/crowemi-go-utils/db"
)

type Watchlist struct {
	ID               string    `bson:"_id, omitempty"`
	CreatedAt        time.Time `bson:"created_at, omitempty"`
	CreatedAtSession string    `bson:"created_at_session, omitempty"`
	UpdatedAt        time.Time `bson:"updated_at, omitempty"`
	UpdateAtSession  string    `bson:"updated_at_session, omitempty"`
	Symbol           string    `bson:"symbol, omitempty"`
	IsActive         bool      `bson:"is_active, omitempty"`
	ExtendedHours    bool      `bson:"extended_hours, omitempty"`
	BatchSize        int       `bson:"batch_size, omitempty"`
	AllowedBatches   int       `bson:"total_allowed_batches, omitempty"`
	Type             string    `bson:"type, omitempty"`
	SubType          string    `bson:"sub_type, omitempty"`
	IsSuspended      bool      `bson:"is_suspend, omitempty"`
}

func GetWatchlists(mongoClient *db.MongoClient) ([]Watchlist, error) {
	// Implement the logic to get allowable investment
	res, err := db.GetMany[Watchlist](context.TODO(), mongoClient, "watchlists", nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}
