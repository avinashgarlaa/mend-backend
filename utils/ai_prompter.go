package utils

import "fmt"

// GeneratePrompt returns an AI-friendly instruction for conflict resolution
// Enhanced version with tone included
func GeneratePrompt(transcript string) string {
	return fmt.Sprintf(`You're a licensed relationship therapist. Here's a message from a couple's conversation:

"%s"

Your role is to:
1. Detect if there's emotional tension, conflict, or misunderstanding.
2. Respond therapeutically — encourage empathy, ask reflective questions, or help de-escalate.
3. Use a warm, calm tone. Be brief but impactful.

Provide only your therapeutic message response.`, transcript)
}

// InterruptWarning returns a gentle reminder when one partner interrupts
func InterruptWarning(partnerName string) string {
	return fmt.Sprintf("Please let %s finish their thought before responding.", partnerName)
}
