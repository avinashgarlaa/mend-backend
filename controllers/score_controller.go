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

// SubmitScore handles manual or AI-generated communication scores
func SubmitScore(c *fiber.Ctx) error {
	var score models.CommunicationScore
	if err := c.BodyParser(&score); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if score.SessionID == "" || score.PartnerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing sessionId or partnerId"})
	}

	score.CreatedAt = time.Now().Unix()

	// 🧠 Auto-generate score via AI if fields are zero
	if score.Empathy == 0 && score.Respect == 0 && score.Listening == 0 {
		messages, err := fetchSessionMessages(score.SessionID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch session messages"})
		}
		aiScore, err := generateAIScore(messages)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "AI scoring failed", "details": err.Error()})
		}
		aiScore.SessionID = score.SessionID
		aiScore.PartnerID = score.PartnerID
		aiScore.CreatedAt = time.Now().Unix()
		score = aiScore
	}

	// 💾 Save score to session document
	sessions := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session struct {
		PartnerA string `bson:"partnerA"`
		PartnerB string `bson:"partnerB"`
	}
	if err := sessions.FindOne(ctx, bson.M{"_id": score.SessionID}).Decode(&session); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Session not found"})
	}

	updateField := "scoreA"
	if score.PartnerID == session.PartnerB {
		updateField = "scoreB"
	}

	if _, err := sessions.UpdateOne(ctx,
		bson.M{"_id": score.SessionID},
		bson.M{"$set": bson.M{updateField: score}},
	); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save score"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Score saved successfully",
		"score":   score,
	})
}

// fetchSessionMessages retrieves messages for a given session
func fetchSessionMessages(sessionId string) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var session models.Session
	if err := database.GetCollection("sessions").
		FindOne(ctx, bson.M{"_id": sessionId}).
		Decode(&session); err != nil {
		return nil, err
	}
	return session.Messages, nil
}

// generateAIScore evaluates communication quality using OpenAI
func generateAIScore(messages []models.Message) (models.CommunicationScore, error) {
	var transcript string
	for _, msg := range messages {
		speaker := msg.SpeakerId
		if speaker == "AI" {
			speaker = "Therapist AI"
		}
		transcript += fmt.Sprintf("%s: %s\n", speaker, msg.Text)
	}

	prompt := fmt.Sprintf(`
You are a therapist AI evaluating a conversation between two people. Based on the transcript below, rate their communication on a scale of 1 to 5 in these areas:

- Empathy
- Listening
- Respect
- Clarity
- Conflict Resolution

Then summarize the emotional tone in 1-2 lines.

Respond in this exact JSON format:

{
  "empathy": 4,
  "listening": 5,
  "respect": 4,
  "clarity": 5,
  "conflictResolution": 4,
  "summary": "The tone was respectful and both parties were attentive to each other."
}

Transcript:
%s`, transcript)

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a therapist AI evaluating communication quality."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.6,
	}
	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf(
		"%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview",
		os.Getenv("OPENAI_ENDPOINT"),
		os.Getenv("OPENAI_DEPLOYMENT"),
	)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return models.CommunicationScore{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var openaiResp map[string]interface{}
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return models.CommunicationScore{}, err
	}

	var content = extractOpenAIText(openaiResp)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return models.CommunicationScore{}, err
	}

	return models.CommunicationScore{
		Empathy:            int(parsed["empathy"].(float64)),
		Listening:          int(parsed["listening"].(float64)),
		Respect:            int(parsed["respect"].(float64)),
		Clarity:            int(parsed["clarity"].(float64)),
		ConflictResolution: int(parsed["conflictResolution"].(float64)),
		Summary:            parsed["summary"].(string),
	}, nil
}

// extractOpenAIText parses message content from OpenAI response
func extractOpenAIText(resp map[string]interface{}) string {
	choices := resp["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	return message["content"].(string)
}

func autoGenerateScore(sessionId string, partnerId string) {
	messages, err := fetchSessionMessages(sessionId)
	if err != nil {
		fmt.Println("❌ Failed to fetch messages:", err)
		return
	}

	score, err := generateAIScore(messages)
	if err != nil {
		fmt.Println("❌ Failed to generate AI score:", err)
		return
	}

	score.SessionID = sessionId
	score.PartnerID = partnerId
	score.CreatedAt = time.Now().Unix()

	sessionsColl := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var session struct {
		PartnerA string `bson:"partnerA"`
		PartnerB string `bson:"partnerB"`
	}
	if err := sessionsColl.FindOne(ctx, bson.M{"_id": sessionId}).Decode(&session); err != nil {
		fmt.Println("❌ Session not found:", err)
		return
	}

	updateField := "scoreA"
	if partnerId == session.PartnerB {
		updateField = "scoreB"
	}

	_, err = sessionsColl.UpdateOne(ctx,
		bson.M{"_id": sessionId},
		bson.M{"$set": bson.M{updateField: score}},
	)
	if err != nil {
		fmt.Println("❌ Failed to save AI score:", err)
	}
}
