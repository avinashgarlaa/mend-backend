package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// SaveReflection handles user or AI-generated reflection
func SaveReflection(c *fiber.Ctx) error {
	var reflection models.Reflection
	if err := c.BodyParser(&reflection); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid reflection data"})
	}

	if reflection.UserID == "" || reflection.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId or sessionId"})
	}

	// 🧠 If no reflection text, generate with AI
	if reflection.Text == "" {
		transcript, err := fetchSessionTranscript(reflection.SessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch session messages"})
		}

		aiText, err := generateAIReflection(transcript)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "AI generation failed", "details": err.Error()})
		}

		reflection.Text = aiText
	}

	// Generate ID and save
	reflection.ID = reflection.SessionID + "-" + reflection.UserID
	reflection.Timestamp = time.Now().Unix()

	collection := database.GetCollection("reflections")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, reflection)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save reflection"})
	}

	return c.Status(201).JSON(reflection)
}

// fetchSessionTranscript gets all messages from a session
func fetchSessionTranscript(sessionId string) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.GetCollection("sessions")
	var session models.Session
	err := collection.FindOne(ctx, bson.M{"_id": sessionId}).Decode(&session)
	if err != nil {
		return nil, err
	}

	return session.Messages, nil
}

// generateAIReflection calls OpenAI to summarize session
func generateAIReflection(messages []models.Message) (string, error) {
	// Format messages for prompt
	var transcript string
	for _, m := range messages {
		speaker := m.SpeakerId
		if speaker == "AI" {
			speaker = "Therapist AI"
		}
		transcript += fmt.Sprintf("%s: %s\n", speaker, m.Text)
	}

	prompt := fmt.Sprintf(`
You are a relationship therapist AI. Given the following chat transcript between two partners, write a gentle, insightful reflection summarizing what was discussed, areas of emotional concern, and any progress made.

Transcript:
%s

Please return a 3-5 sentence therapist-style reflection.
`, transcript)

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a compassionate therapist AI that helps couples reflect on their communication."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	endpoint := os.Getenv("OPENAI_ENDPOINT")
	deployment := os.Getenv("OPENAI_DEPLOYMENT")

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", endpoint, deployment)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	// Extract AI reply
	choices := result["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

// GetInsights godoc
// @Summary      Get communication insights for a user
// @Description  Returns sessions, reflections, scores, and emotional bonding data for a user
// @Tags         Insights
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/insights/{userId} [get]
func GetInsights(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessionsCollection := database.GetCollection("sessions")
	reflectionCollection := database.GetCollection("reflections")
	postResCollection := database.GetCollection("postResolution")

	// 🧾 Get all sessions where the user is involved
	sessionCursor, err := sessionsCollection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"partnerA": userId},
			{"partnerB": userId},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching sessions"})
	}

	var sessions []models.Session
	if err := sessionCursor.All(ctx, &sessions); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode sessions"})
	}

	// 💬 Get reflections written by the user
	reflectionCursor, err := reflectionCollection.Find(ctx, bson.M{"userId": userId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching reflections"})
	}

	var reflections []models.Reflection
	if err := reflectionCursor.All(ctx, &reflections); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode reflections"})
	}

	// ❤️ Get post-resolution feedback (emotional bonding data)
	postResCursor, err := postResCollection.Find(ctx, bson.M{"userId": userId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching post-resolution entries"})
	}

	var postRes []models.PostResolution
	if err := postResCursor.All(ctx, &postRes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode post-resolution entries"})
	}

	// 🔍 Assemble insights (more features like AI scoring summary or timeline trend can be added here)
	return c.Status(200).JSON(fiber.Map{
		"sessions":     sessions,
		"reflections":  reflections,
		"postFeedback": postRes,
	})
}
