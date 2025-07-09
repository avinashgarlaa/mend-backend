package models

type Message struct {
	SpeakerId string `json:"speakerId" bson:"speakerId"`
	SessionId string `json:"sessionId" bson:"sessionId"`
	Text      string `json:"text" bson:"text"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
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
