package crowemi_trades

import "time"

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
	AllowedBatches   int       `bson:"allowed_batches, omitempty"`
	Type             string    `bson:"type, omitempty"`
	SubType          string    `bson:"sub_type, omitempty"`
	IsSuspended      bool      `bson:"is_suspended, omitempty"`
}
