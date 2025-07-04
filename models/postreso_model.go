package models

type PostResolution struct {
	UserID          string `json:"userId" bson:"userId"`
	SessionID       string `json:"sessionId" bson:"sessionId"`
	Gratitude       string `json:"gratitude" bson:"gratitude"`
	Reflection      string `json:"reflection" bson:"reflection"`
	BondingActivity string `json:"bondingActivity" bson:"bondingActivity"`
	Timestamp       int64  `json:"timestamp" bson:"timestamp"`
}
