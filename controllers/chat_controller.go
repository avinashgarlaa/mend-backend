// StartModeratedSession godoc
// @Summary Start a moderated AI voice session
// @Tags Chat
// @Accept json
// @Produce json
// @Param session body models.Session true "Session Info"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/moderate [post]
// SubmitReflection godoc
// @Summary Submit reflection after a session
// @Tags Chat
// @Accept json
// @Produce json
// @Param reflection body models.Reflection true "Reflection info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/reflection [post]

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
)

// POST /api/session
func StartSession(c *fiber.Ctx) error {
	var session models.Session
	if err := c.BodyParser(&session); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid session data"})
	}

	session.ID = utils.GeneratePartnerID()
	session.CreatedAt = time.Now().Unix()
	session.Resolved = false

	collection := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, session)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	return c.Status(201).JSON(session)
}

// POST /api/moderate
func ModerateChat(c *fiber.Ctx) error {
	type ChatRequest struct {
		Transcript string `json:"transcript"`
		Context    string `json:"context"`
		Speaker    string `json:"speaker"`
	}

	var body ChatRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	prompt := utils.GeneratePrompt(body.Transcript)
	openaiKey := os.Getenv("OPENAI_API_KEY")

	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a supportive AI couples therapist."},
			{"role": "user", "content": prompt},
		},
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", ioutil.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to prepare request"})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiKey)

	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI request failed"})
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var gptResponse map[string]interface{}
	json.Unmarshal(bodyBytes, &gptResponse)

	reply := gptResponse["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": utils.InterruptWarning(body.Speaker),
	})
}
