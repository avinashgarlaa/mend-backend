package models

type User struct {
	ID         string   `json:"id" bson:"id"`                 // Unique UUID
	Name       string   `json:"name" bson:"name"`             // User's name
	Gender     string   `json:"gender" bson:"gender"`         // e.g., male, female, non-binary
	Goals      []string `json:"goals" bson:"goals"`           // Relationship goals
	Challenges []string `json:"challenges" bson:"challenges"` // Current challenges
	PartnerID  string   `json:"partnerId" bson:"partnerId"`   // Linked partner's ID
	ColorCode  string   `json:"colorCode" bson:"colorCode"`   // UI color (e.g., pink, blue)
}
