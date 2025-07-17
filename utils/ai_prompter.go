package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

// GeneratePrompt returns an AI-friendly instruction for conflict resolution
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

// ModerateText uses OpenAI to analyze a message for tone, respect, and helpfulness
type ModerationResult struct {
	Warning   string `json:"warning,omitempty"`
	Tone      string `json:"tone,omitempty"`
	Respect   string `json:"respect,omitempty"`
	Clarity   string `json:"clarity,omitempty"`
	Empathy   string `json:"empathy,omitempty"`
	IsFlagged bool   `json:"is_flagged"`
}

// ModerateText runs a moderation check using OpenAI on a given message
func ModerateText(message, speaker string) ModerationResult {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	prompt := fmt.Sprintf(`You're a communication coach reviewing a message in a couple's therapy session.

Message from %s:
"%s"

Evaluate the following:
1. Tone: Is it calm, angry, respectful, etc.?
2. Respect: Is the message respectful?
3. Clarity: Is the message clear or vague?
4. Empathy: Does it show understanding of the partner’s feelings?

Also, if the message contains harmful, aggressive, or disrespectful content, provide a short warning.

Respond with JSON:
{
  "tone": "...",
  "respect": "...",
  "clarity": "...",
  "empathy": "...",
  "warning": "...",  // empty if no warning
  "is_flagged": true/false
}
`, speaker, message)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are an AI therapist helping with communication analysis."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.4,
	})
	if err != nil {
		log.Println("Moderation API error:", err)
		return ModerationResult{}
	}

	// Attempt to parse response as JSON
	var result ModerationResult
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		log.Println("Failed to parse moderation result:", err)
		return ModerationResult{}
	}

	return result
}
