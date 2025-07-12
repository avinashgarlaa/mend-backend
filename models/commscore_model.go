package models

type CommunicationScore struct {
	SessionID          string `json:"sessionId" bson:"sessionId"`
	PartnerID          string `json:"partnerId" bson:"partnerId"`
	Empathy            int    `json:"empathy" bson:"empathy"`
	Listening          int    `json:"listening" bson:"listening"`
	Respect            int    `json:"respect" bson:"respect"`
	Clarity            int    `json:"clarity" bson:"clarity"`
	ConflictResolution int    `json:"conflictResolution" bson:"conflictResolution"`
	Summary            string `json:"summary,omitempty" bson:"summary,omitempty"`
	CreatedAt          int64  `json:"createdAt" bson:"createdAt"`
}
