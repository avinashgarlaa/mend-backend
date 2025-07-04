package models

type CommunicationScore struct {
	UserID         string  `json:"userId" bson:"userId"`
	SessionID      string  `json:"sessionId" bson:"sessionId"`
	Empathy        float64 `json:"empathy" bson:"empathy"`
	Listening      float64 `json:"listening" bson:"listening"`
	Clarity        float64 `json:"clarity" bson:"clarity"`
	Respect        float64 `json:"respect" bson:"respect"`
	Responsiveness float64 `json:"responsiveness" bson:"responsiveness"`
	OpenMindedness float64 `json:"openMindedness" bson:"openMindedness"`
	CreatedAt      int64   `json:"createdAt" bson:"createdAt"`
}
