package models

type User struct {
	ID         string   `json:"id" bson:"id"`
	Name       string   `json:"name" bson:"name"`
	Gender     string   `json:"gender" bson:"gender"`
	Email      string   `json:"email" bson:"email"`
	Goals      []string `json:"goals,omitempty" bson:"goals,omitempty"`
	Challenges []string `json:"challenges,omitempty" bson:"challenges,omitempty"`
	PartnerID  string   `json:"partnerId,omitempty" bson:"partnerId,omitempty"`
	ColorCode  string   `json:"colorCode" bson:"colorCode"`
	InvitedBy  string   `json:"invitedBy,omitempty" bson:"invitedBy,omitempty"` // ✅ NEW
}
