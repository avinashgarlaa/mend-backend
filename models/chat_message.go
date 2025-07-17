package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID  string             `bson:"sessionId" json:"sessionId"`
	SpeakerID  string             `bson:"speakerId" json:"speakerId"`
	Content    string             `bson:"content" json:"content"`
	Type       string             `bson:"type" json:"type"` // text/audio
	IsFlagged  bool               `bson:"isFlagged,omitempty" json:"isFlagged,omitempty"`
	Moderation string             `bson:"moderation,omitempty" json:"moderation,omitempty"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
}
