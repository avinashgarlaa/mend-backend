package controllers

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	openai "github.com/sashabaranov/go-openai"
)

func ModerateVoiceInput(c *fiber.Ctx) error {
	var input struct {
		Transcript string `json:"transcript"`
		Speaker    string `json:"speaker"`
		Context    string `json:"context"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if strings.TrimSpace(input.Transcript) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Transcript is required"})
	}

	// Load from env
	apiKey := os.Getenv("OPENAI_API_KEY")
	endpoint := os.Getenv("OPENAI_ENDPOINT")     // e.g. https://your-resource-name.openai.azure.com/
	deployment := os.Getenv("OPENAI_DEPLOYMENT") // e.g. gpt-35-turbo
	apiVersion := "2023-05-15"                   // required by Azure

	config := openai.DefaultAzureConfig(apiKey, endpoint)
	config.APIVersion = apiVersion

	client := openai.NewClientWithConfig(config)

	// Build moderation prompt
	prompt := `
You are a conversation moderator helping couples communicate better.
Speaker: ` + input.Speaker + `
Transcript: "` + input.Transcript + `"
Context: "` + input.Context + `"

Evaluate this input. Respond in JSON with:
- tone: ["respectful", "hostile", "passive", "supportive", "neutral"]
- empathy: score out of 10
- clarity: score out of 10
- respect: score out of 10
- warning: true/false if this should trigger a warning to the speaker
`

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: deployment,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are a conversation moderator helping partners speak respectfully."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		log.Println("OpenAI API error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI moderation failed"})
	}

	result := resp.Choices[0].Message.Content
	return c.JSON(fiber.Map{"moderation": result})
}
