package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
// @Summary      OpenAI moderation of transcript
// @Description  Returns AI feedback + tone warning
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        input body map[string]string true "Transcript, Context (optional), Speaker"
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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input body"})
	}

	if body.Transcript == "" || body.Speaker == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields: transcript or speaker"})
	}

	apiKey := os.Getenv("OPENAI_API_KEY")        // For Azure, set AZURE_OPENAI_KEY and use it below
	endpoint := os.Getenv("OPENAI_ENDPOINT")     // For Azure, e.g., https://your-resource-name.openai.azure.com
	deployment := os.Getenv("OPENAI_DEPLOYMENT") // For Azure: e.g., "gpt-35-turbo"

	if apiKey == "" || endpoint == "" || deployment == "" {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI config missing"})
	}

	// 🧠 Build prompt
	prompt := utils.GeneratePrompt(body.Transcript)

	// 🧾 OpenAI Payload
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a relationship therapist AI helping moderate couple conversations with respectful tone and helpful suggestions."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to encode OpenAI payload"})
	}

	// 🌐 OpenAI or Azure endpoint
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", endpoint, deployment)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create OpenAI request"})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI request failed"})
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read OpenAI response"})
	}

	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "OpenAI API error", "details": string(bodyBytes)})
	}

	var openaiResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &openaiResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse OpenAI response"})
	}

	// ✅ Extract reply from OpenAI
	reply := extractOpenAIReply(openaiResp)
	if reply == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Empty response from AI"})
	}

	// 🔎 Check for interruption warning
	interruptWarning := utils.InterruptWarning(body.Speaker)

	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": interruptWarning,
	})
}

// extractOpenAIReply extracts content from chat completion
func extractOpenAIReply(resp map[string]interface{}) string {
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}

	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return ""
	}

	message, ok := first["message"].(map[string]interface{})
	if !ok {
		return ""
	}

	content, ok := message["content"].(string)
	if !ok {
		return ""
	}

	return content
}
