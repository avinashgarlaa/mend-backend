package models

type CommunicationScore struct {
	SessionID      string `json:"sessionId" bson:"sessionId"` // Which session this score is for
	PartnerID      string `json:"partnerId" bson:"partnerId"` // Who submitted this
	Empathy        int    `json:"empathy" bson:"empathy"`
	Listening      int    `json:"listening" bson:"listening"`
	Clarity        int    `json:"clarity" bson:"clarity"`
	Respect        int    `json:"respect" bson:"respect"`
	Responsiveness int    `json:"responsiveness" bson:"responsiveness"`
	OpenMindedness int    `json:"openMindedness" bson:"openMindedness"`
	CreatedAt      int64  `json:"createdAt" bson:"createdAt"`
}
