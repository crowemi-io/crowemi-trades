package trader

import "time"

type Activities struct {
	ActivityType    string    `json:"activity_type" omitempty:"true"`
	ActivitySubType string    `json:"activity_subtype" omitempty:"true"`
	Date            time.Time `json:"date" omitempty:"true"`
	NetAmount       float64   `json:"net_amount" omitempty:"true"`
	Description     string    `json:"description" omitempty:"true"`
	Status          string    `json:"status" omitempty:"true"`
	Symbol          string    `json:"symbol" omitempty:"true"`
	PerShareAmount  float64   `json:"per_share_amount" omitempty:"true"`
}
