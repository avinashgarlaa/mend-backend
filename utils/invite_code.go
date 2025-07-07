package utils

import (
	"github.com/google/uuid"
)

// GeneratePartnerID returns a unique UUID for user/session/partner IDs
func GeneratePartnerID() string {
	return uuid.NewString()
}
