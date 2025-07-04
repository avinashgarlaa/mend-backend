package models

type Reflection struct {
	ID           string `json:"id" bson:"_id"`                    // Reflection ID (UUID)
	SessionID    string `json:"sessionId" bson:"sessionId"`       // Associated session
	UserID       string `json:"userId" bson:"userId"`             // Reflecting user
	Gratitude    string `json:"gratitude" bson:"gratitude"`       // "Thank you for..."
	Appreciation string `json:"appreciation" bson:"appreciation"` // "I liked that you..."
	Commitment   string `json:"commitment" bson:"commitment"`     // "Going forward, I will..."
	Timestamp    int64  `json:"timestamp" bson:"timestamp"`       // Unix time
}
