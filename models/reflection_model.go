package models

type Reflection struct {
	ID        string `json:"id" bson:"_id"`              // Composite ID: sessionID-userID
	SessionID string `json:"sessionId" bson:"sessionId"` // Which session
	UserID    string `json:"userId" bson:"userId"`       // Who submitted
	Text      string `json:"text" bson:"text"`           // Reflection content
	Timestamp int64  `json:"timestamp" bson:"timestamp"` // Unix time
}
