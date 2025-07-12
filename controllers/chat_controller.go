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

// HandleWebSocket handles real-time chat messages
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

		// Simple interruption moderation
		if strings.Contains(strings.ToLower(message.Text), "interrupt") {
			c.WriteMessage(websocket.TextMessage, []byte("INTERRUPT: Please wait your turn."))
			continue
		}

		// Save message
		go appendMessageToSessionByID(message.SessionId, message)

		// Broadcast to all clients in session
		for id, conn := range clients {
			if idHasSession(id, sessionId) {
				conn.WriteMessage(websocket.TextMessage, msg)
			}
		}

		// AI moderation & response if needed
		go maybeTriggerTherapistAI(message, sessionId)
	}
}

// Save message to MongoDB session
func appendMessageToSessionByID(sessionId string, msg models.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": sessionId}
	update := bson.M{"$push": bson.M{"messages": msg}}
	sessions := database.GetCollection("sessions")
	_, _ = sessions.UpdateOne(ctx, filter, update)
}

// Check if key belongs to session
func idHasSession(clientKey, sessionId string) bool {
	parts := strings.Split(clientKey, ":")
	return len(parts) == 2 && parts[1] == sessionId
}

// AI moderation & reply trigger
func maybeTriggerTherapistAI(message models.Message, sessionId string) {
	lower := strings.ToLower(message.Text)
	triggerWords := []string{"always", "never", "angry", "hurt", "you don't", "why do you", "not fair"}

	for _, w := range triggerWords {
		if strings.Contains(lower, w) {
			go func() {
				payload := map[string]string{
					"transcript": message.Text,
					"speaker":    message.SpeakerId,
				}
				payloadBytes, _ := json.Marshal(payload)

				resp, err := http.Post("http://localhost:3000/api/moderate", "application/json", bytes.NewBuffer(payloadBytes))
				if err != nil {
					fmt.Println("AI moderation failed:", err)
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				var parsed map[string]interface{}
				_ = json.Unmarshal(body, &parsed)

				reply, _ := parsed["aiReply"].(string)
				if reply == "" {
					return
				}

				aiMessage := models.Message{
					Text:      reply,
					SessionId: sessionId,
					SpeakerId: "AI",
					Timestamp: time.Now().Unix(),
				}

				appendMessageToSessionByID(sessionId, aiMessage)

				aiJson, _ := json.Marshal(aiMessage)
				for id, conn := range clients {
					if idHasSession(id, sessionId) {
						conn.WriteMessage(websocket.TextMessage, aiJson)
					}
				}
			}()
			break
		}
	}
}

// ModerateChat (API)
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
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	endpoint := os.Getenv("OPENAI_ENDPOINT")
	deployment := os.Getenv("OPENAI_DEPLOYMENT")
	if apiKey == "" || endpoint == "" || deployment == "" {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI config missing"})
	}

	prompt := utils.GeneratePrompt(body.Transcript)
	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a kind, empathetic therapist AI guiding respectful conversation between partners."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	jsonData, _ := json.Marshal(payload)
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

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "OpenAI API error", "details": string(bodyBytes)})
	}

	var openaiResp map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &openaiResp)
	reply := extractOpenAIReply(openaiResp)
	if reply == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Empty response from AI"})
	}

	return c.Status(200).JSON(fiber.Map{
		"aiReply":   reply,
		"interrupt": utils.InterruptWarning(body.Speaker),
	})
}

// Extract reply from OpenAI response
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
