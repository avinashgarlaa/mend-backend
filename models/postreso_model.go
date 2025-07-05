package models

type PostResolution struct {
	UserID          string `json:"userId" bson:"userId"`                   // ID of the user submitting
	SessionID       string `json:"sessionId" bson:"sessionId"`             // Associated session ID
	Gratitude       string `json:"gratitude" bson:"gratitude"`             // e.g., "I'm grateful for your honesty"
	Reflection      string `json:"reflection" bson:"reflection"`           // e.g., "I felt better after the session"
	BondingActivity string `json:"bondingActivity" bson:"bondingActivity"` // e.g., "Go for a walk together"
	Timestamp       int64  `json:"timestamp" bson:"timestamp"`             // Unix time of submission
}
