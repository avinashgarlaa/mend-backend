package models

import "time"

type Onboarding struct {
	UserID            string    `json:"userId" bson:"userId"`
	Name              string    `json:"name" bson:"name"`
	Gender            string    `json:"gender" bson:"gender"`
	RelationshipGoals []string  `json:"relationshipGoals" bson:"relationshipGoals"`
	OtherGoal         string    `json:"otherGoal,omitempty" bson:"otherGoal,omitempty"`
	CurrentChallenges []string  `json:"currentChallenges" bson:"currentChallenges"`
	OtherChallenge    string    `json:"otherChallenge,omitempty" bson:"otherChallenge,omitempty"`
	CreatedAt         time.Time `json:"createdAt" bson:"createdAt"`
}
