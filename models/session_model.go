package models

type Message struct {
	SpeakerID string `json:"speakerId" bson:"speakerId"` // Who spoke
	Text      string `json:"text" bson:"text"`           // Transcript
	Timestamp int64  `json:"timestamp" bson:"timestamp"` // Unix time
}

type Session struct {
	ID        string             `json:"id" bson:"_id"`              // UUID
	PartnerA  string             `json:"partnerA" bson:"partnerA"`   // User A
	PartnerB  string             `json:"partnerB" bson:"partnerB"`   // User B
	Messages  []Message          `json:"messages" bson:"messages"`   // Chat transcript
	ScoreA    CommunicationScore `json:"scoreA" bson:"scoreA"`       // A's score
	ScoreB    CommunicationScore `json:"scoreB" bson:"scoreB"`       // B's score
	CreatedAt int64              `json:"createdAt" bson:"createdAt"` // Session time
	Resolved  bool               `json:"resolved" bson:"resolved"`   // Has reflection happened
}
