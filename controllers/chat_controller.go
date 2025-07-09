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
// @Summary      Gemini AI moderation of transcript
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

	// 🧠 Generate AI-friendly prompt
	prompt := utils.GeneratePrompt(body.Transcript)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Gemini API key not configured"})
	}

	// 📦 Gemini JSON request payload
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encode payload"})
	}

	// 🌐 Send request to Gemini API
	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key="+apiKey,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create Gemini request"})
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gemini request failed"})
	}
	defer resp.Body.Close()

	// 📥 Parse response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read Gemini response"})
	}

	var geminiResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse Gemini response"})
	}

	// ✅ Extract reply text from Gemini structure
	candidates, ok := geminiResp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid Gemini response: no candidates"})
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid candidate structure"})
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Missing content in Gemini response"})
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Missing parts in Gemini response"})
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid part structure"})
	}

	reply, ok := part["text"].(string)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid reply text"})
	}

	// 🧠 Return AI reply + interruption warning
	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": utils.InterruptWarning(body.Speaker),
	})
}
