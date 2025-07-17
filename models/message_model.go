package models

import (
	"time"
)

type Messages struct {
	SessionID string    `json:"sessionId" bson:"sessionId"`
	SpeakerID string    `json:"speakerId" bson:"speakerId"`
	Text      string    `json:"text" bson:"text"`
	IsAI      bool      `json:"isAI" bson:"isAI"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
