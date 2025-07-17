package controllers

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/websocket/v2"
)

var wsClients = make(map[string]*websocket.Conn) // Active user connections (userId -> Conn)

// HandleWebSocket handles real-time chat with AI moderation
func HandleWebSockets(c *websocket.Conn) {
	sessionId := c.Params("sessionId")
	userId := c.Params("userId")

	if sessionId == "" || userId == "" {
		log.Println("❌ Missing sessionId or userId in WebSocket connection")
		return
	}

	// Store client connection
	wsClients[userId] = c
	log.Printf("✅ User %s connected to session %s\n", userId, sessionId)

	defer func() {
		c.Close()
		delete(wsClients, userId)
		log.Printf("👋 User %s disconnected from session\n", userId)
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("❌ Error reading message from %s: %v\n", userId, err)
			break
		}

		var incoming models.ChatMessage
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Printf("❌ Invalid JSON from %s: %v\n", userId, err)
			continue
		}

		// Add sessionId and timestamp to message
		incoming.SessionID = sessionId
		incoming.Timestamp = time.Now()

		// AI Moderation for text messages
		if strings.TrimSpace(incoming.Type) == "text" {
			moderation := utils.ModerateText(incoming.Content, incoming.SpeakerID)
			if moderation.Warning != "" {
				incoming.IsFlagged = true
				incoming.Moderation = moderation.Warning
				log.Printf("⚠️ Moderation warning for user %s: %s\n", userId, moderation.Warning)
			}
		}

		// Save to MongoDB
		collection := database.GetCollection("messages")
		_, err = collection.InsertOne(context.TODO(), incoming)
		if err != nil {
			log.Printf("❌ Failed to insert message for user %s: %v\n", userId, err)
		}

		// Broadcast message to all other users in session
		broadcastMsg, _ := json.Marshal(incoming)
		for id, conn := range wsClients {
			if id != userId {
				if err := conn.WriteMessage(websocket.TextMessage, broadcastMsg); err != nil {
					log.Printf("❌ Error sending message to %s: %v\n", id, err)
				}
			}
		}
	}
}
