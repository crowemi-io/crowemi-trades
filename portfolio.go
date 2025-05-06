package crowemi_trades

type Portfolio struct {
	ID   string `bson:"_id"`
	Type string `bson:"type" json:"type"`
	Name string `bson:"name" json:"name"`
}
