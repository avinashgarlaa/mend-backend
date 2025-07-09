package utils

import "fmt"

// GeneratePrompt returns an AI-friendly instruction for conflict resolution
// Enhanced version with tone included
func GeneratePrompt(context string) string {
	return fmt.Sprintf(`You are an emotionally intelligent AI therapist.
The couple is discussing the following issue: "%s"
- Provide a calm, validating response.
- Ask a brief open-ended question to encourage mutual understanding.
- Do not assign blame.
- Keep it under 80 words.`, context)
}

// InterruptWarning returns a gentle reminder when one partner interrupts
func InterruptWarning(partnerName string) string {
	return fmt.Sprintf("Please let %s finish their thought before responding.", partnerName)
}
