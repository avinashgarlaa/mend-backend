package models

type PostResolution struct {
	ID              string `json:"id" bson:"_id"`
	SessionID       string `json:"sessionId" bson:"sessionId"`
	UserID          string `json:"userId" bson:"userId"`
	Gratitude       string `json:"gratitude" bson:"gratitude"`                                 // Freeform reflection
	SharedFeelings  string `json:"sharedFeelings,omitempty" bson:"sharedFeelings,omitempty"`   // Optional
	AttachmentScore int    `json:"attachmentScore,omitempty" bson:"attachmentScore,omitempty"` // Optional 1-5
	Timestamp       int64  `json:"timestamp" bson:"timestamp"`
}
