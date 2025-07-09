package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// WebSocket clients: key = userId:sessionId
var clients = make(map[string]*websocket.Conn)

// handleWebSocket handles real-time chat messages
func HandleWebSocket(c *websocket.Conn) {
	userId := c.Params("userId")
	sessionId := c.Params("sessionId")
	clientKey := userId + ":" + sessionId

	clients[clientKey] = c
	defer func() {
		c.Close()
		delete(clients, clientKey)
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

		// 🛡️ Simple interruption moderation
		if strings.Contains(strings.ToLower(message.Text), "interrupt") {
			c.WriteMessage(websocket.TextMessage, []byte("INTERRUPT: Please wait your turn."))
			continue
		}

		// 💾 Save to DB
		go appendMessageToSessionByID(message.SessionId, message)

		// 📤 Broadcast to other clients in the session
		for id, conn := range clients {
			if id != clientKey && idHasSession(id, sessionId) {
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}

// appendMessageToSessionByID saves a message to MongoDB session by sessionId
func appendMessageToSessionByID(sessionId string, msg models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": sessionId}
	update := bson.M{"$push": bson.M{"messages": msg}}

	sessions := database.GetCollection("sessions")
	_, _ = sessions.UpdateOne(ctx, filter, update)
}

// idHasSession checks if a client ID belongs to a session
func idHasSession(clientKey, sessionId string) bool {
	parts := strings.Split(clientKey, ":")
	return len(parts) == 2 && parts[1] == sessionId
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

	// ✅ Safely extract the message content
	choices, ok := gptResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid GPT response: no choices"})
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid GPT message structure"})
	}

	reply, ok := message["content"].(string)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid GPT reply content"})
	}

	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": utils.InterruptWarning(body.Speaker),
	})
}
