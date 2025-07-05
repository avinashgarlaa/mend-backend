package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocket clients map
var clients = make(map[string]*websocket.Conn)

// SetupWebSocket adds the real-time chat endpoint
func SetupWebSocket(app *fiber.App) {
	app.Use("/ws/:userId", websocket.New(func(c *websocket.Conn) {
		userId := c.Params("userId")
		clients[userId] = c
		defer func() {
			c.Close()
			delete(clients, userId)
		}()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			var message models.Message
			if err := json.Unmarshal(msg, &message); err != nil {
				continue
			}
			message.Timestamp = time.Now().Unix()

			// Broadcast to other users
			for id, conn := range clients {
				if id != userId {
					conn.WriteMessage(websocket.TextMessage, msg)
				}
			}

			// Persist in DB
			go appendMessageToSession(userId, message)
		}
	}))
}

// Appends message to active session in DB
func appendMessageToSession(userId string, msg models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessions := database.GetCollection("sessions")
	filter := fiber.Map{
		"$or": []fiber.Map{
			{"partnerA": userId},
			{"partnerB": userId},
		},
		"resolved": false,
	}
	update := fiber.Map{
		"$push": fiber.Map{"messages": msg},
	}
	_, _ = sessions.UpdateOne(ctx, filter, update)
}

// ModerateChat godoc
// @Summary      GPT-4 AI moderation of transcript
// @Description  Returns AI feedback + tone warning
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        input body map[string]string true "Transcript, Speaker"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/moderate [post]
func ModerateChat(c *fiber.Ctx) error {
	type ChatRequest struct {
		Transcript string `json:"transcript"`
		Context    string `json:"context"`
		Speaker    string `json:"speaker"`
	}

	var body ChatRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if body.Transcript == "" || body.Speaker == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing transcript or speaker"})
	}

	prompt := utils.GeneratePrompt(body.Transcript)
	openaiKey := os.Getenv("OPENAI_API_KEY")

	payload := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful AI couples therapist."},
			{"role": "user", "content": prompt},
		},
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to prepare GPT request"})
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiKey)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI request failed"})
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var gptResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &gptResponse); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse GPT response"})
	}

	reply := gptResponse["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": utils.InterruptWarning(body.Speaker),
	})
}
