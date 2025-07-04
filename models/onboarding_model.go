package models

type Onboarding struct {
	UserID string `json:"userId" bson:"userId"`
	Name   string `json:"name" bson:"name"`
	Gender string `json:"gender" bson:"gender"`

	RelationshipGoals []string `json:"relationshipGoals" bson:"relationshipGoals"`     // multiple-choice
	OtherGoal         string   `json:"otherGoal,omitempty" bson:"otherGoal,omitempty"` // optional custom input

	CurrentChallenges []string `json:"currentChallenges" bson:"currentChallenges"`               // multiple-choice
	OtherChallenge    string   `json:"otherChallenge,omitempty" bson:"otherChallenge,omitempty"` // optional custom input
}
