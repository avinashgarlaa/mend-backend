package utils

import "fmt"

// GeneratePrompt returns an AI-friendly instruction for conflict resolution
func GeneratePrompt(context string) string {
	return fmt.Sprintf(`You are a calm, emotionally intelligent AI therapist. 
The couple is discussing this: "%s". Ask an open-ended, empathetic question 
to help them communicate respectfully and reach resolution.`, context)
}

// InterruptWarning returns a gentle reminder when one partner interrupts
func InterruptWarning(partnerName string) string {
	return fmt.Sprintf("Please let %s finish their thought before responding.", partnerName)
}
