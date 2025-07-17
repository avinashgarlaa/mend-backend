package controllers

import (
	"encoding/json"
	"log"
	"sync"

	"mend/ai"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SessionConn stores user connections within a session
type SessionConn struct {
	Conns map[string]*websocket.Conn
	Lock  sync.RWMutex
}

var (
	sessions     = make(map[string]*SessionConn)
	sessionsLock sync.RWMutex
)

// HandleWebSocket2 upgrades to WebSocket
func HandleWebSocket2(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocketHandler2 handles real-time messages
func WebSocketHandler2(c *websocket.Conn) {
	sessionId := c.Params("sessionId")
	userId := c.Params("userId")

	// Register user connection
	sessionsLock.Lock()
	if sessions[sessionId] == nil {
		sessions[sessionId] = &SessionConn{
			Conns: make(map[string]*websocket.Conn),
		}
	}
	sessions[sessionId].Lock.Lock()
	sessions[sessionId].Conns[userId] = c
	sessions[sessionId].Lock.Unlock()
	sessionsLock.Unlock()

	defer func() {
		log.Println("Disconnected:", userId)
		sessionsLock.Lock()
		sessions[sessionId].Lock.Lock()
		delete(sessions[sessionId].Conns, userId)
		sessions[sessionId].Lock.Unlock()
		sessionsLock.Unlock()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Broadcast to all other participants
		sessions[sessionId].Lock.RLock()
		for otherId, conn := range sessions[sessionId].Conns {
			if otherId == userId {
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("Write error:", err)
			}
		}
		sessions[sessionId].Lock.RUnlock()

		// AI moderation logic for transcript messages
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err == nil {
			if msgType, ok := data["type"].(string); ok && msgType == "transcript" {
				text, _ := data["text"].(string)
				speaker, _ := data["speaker"].(string)
				go handleAIModeration(sessionId, text, speaker)
			}
		}
	}
}

func handleAIModeration(sessionId string, transcript, speaker string) {
	warning := ai.ModerateTranscript(transcript, speaker)
	if warning == "" {
		return
	}

	resp := map[string]interface{}{
		"type":    "ai_warning",
		"message": warning,
	}
	msg, _ := json.Marshal(resp)

	sessions[sessionId].Lock.RLock()
	defer sessions[sessionId].Lock.RUnlock()

	for _, conn := range sessions[sessionId].Conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("AI warning write error:", err)
		}
	}
}
