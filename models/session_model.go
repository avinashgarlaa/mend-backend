package models

type Message struct {
	SpeakerID string `json:"speakerId" bson:"speakerId"` // ID of the speaker (user ID)
	Text      string `json:"text" bson:"text"`           // What was said
	Timestamp int64  `json:"timestamp" bson:"timestamp"` // Unix timestamp
}

type Score struct {
	Empathy        int `json:"empathy" bson:"empathy"`
	Listening      int `json:"listening" bson:"listening"`
	Clarity        int `json:"clarity" bson:"clarity"`
	Respect        int `json:"respect" bson:"respect"`
	Responsiveness int `json:"responsiveness" bson:"responsiveness"`
	OpenMindedness int `json:"openMindedness" bson:"openMindedness"`
}

type Session struct {
	ID        string    `json:"id" bson:"_id"`              // Session ID (UUID)
	PartnerA  string    `json:"partnerA" bson:"partnerA"`   // User ID A
	PartnerB  string    `json:"partnerB" bson:"partnerB"`   // User ID B
	Messages  []Message `json:"messages" bson:"messages"`   // All spoken messages
	ScoreA    Score     `json:"scoreA" bson:"scoreA"`       // Score for Partner A
	ScoreB    Score     `json:"scoreB" bson:"scoreB"`       // Score for Partner B
	CreatedAt int64     `json:"createdAt" bson:"createdAt"` // Session timestamp
	Resolved  bool      `json:"resolved" bson:"resolved"`   // Conflict resolved status
}
