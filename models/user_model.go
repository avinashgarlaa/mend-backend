package models

import "time"

type User struct {
	ID             string    `json:"id" bson:"id"`
	Name           string    `json:"name" bson:"name"`
	Email          string    `json:"email" bson:"email"`
	Password       string    `json:"password,omitempty" bson:"password,omitempty"`
	Gender         string    `json:"gender,omitempty" bson:"gender,omitempty"`
	Goals          []string  `json:"goals,omitempty" bson:"goals,omitempty"`
	OtherGoal      string    `json:"otherGoal,omitempty" bson:"otherGoal,omitempty"`
	Challenges     []string  `json:"challenges,omitempty" bson:"challenges,omitempty"`
	OtherChallenge string    `json:"otherChallenge,omitempty" bson:"otherChallenge,omitempty"`
	ColorCode      string    `json:"colorCode,omitempty" bson:"colorCode,omitempty"`
	PartnerID      string    `json:"partnerId,omitempty" bson:"partnerId,omitempty"`
	InvitedBy      string    `json:"invitedBy,omitempty" bson:"invitedBy,omitempty"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
}
