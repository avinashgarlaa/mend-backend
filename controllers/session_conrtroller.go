package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// StartSession godoc
// @Summary      Start a new session between partners
// @Description  Creates a session document in MongoDB
// @Tags         Session
// @Accept       json
// @Produce      json
// @Param        session body models.Session true "Session Info"
// @Success      201 {object} models.Session
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/session [post]
func StartSession(c *fiber.Ctx) error {
	var session models.Session
	if err := c.BodyParser(&session); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid session data"})
	}

	if session.PartnerA == "" || session.PartnerB == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Both partner IDs required"})
	}

	session.ID = utils.GeneratePartnerID()
	session.CreatedAt = time.Now().Unix()
	session.Resolved = false
	session.Messages = []models.Message{}
	session.ScoreA = models.CommunicationScore{}
	session.ScoreB = models.CommunicationScore{}

	collection := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, session)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	return c.Status(201).JSON(session)
}

// GetActiveSession godoc
// @Summary Get active (unresolved) session for a user
// @Tags Session
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} models.Session
// @Failure 404 {object} map[string]string
// @Router /api/session/active/{userId} [get]
func GetActiveSession(c *fiber.Ctx) error {
	userId := c.Params("userId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := fiber.Map{
		"$or": []fiber.Map{
			{"partnerA": userId},
			{"partnerB": userId},
		},
		"resolved": false,
	}

	var session models.Session
	err := database.GetCollection("sessions").FindOne(ctx, filter).Decode(&session)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No active session"})
	}
	return c.JSON(session)
}

// EndSession godoc
// @Summary      Mark a session as resolved
// @Description  Updates the session's resolved field to true
// @Tags         Session
// @Accept       json
// @Produce      json
// @Param        sessionId path string true "Session ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/session/end/{sessionId} [patch]
func EndSession(c *fiber.Ctx) error {
	sessionId := c.Params("sessionId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": sessionId}
	update := bson.M{"$set": bson.M{"resolved": true}}

	result, err := database.GetCollection("sessions").UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to end session"})
	}
	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	return c.JSON(fiber.Map{"message": "Session ended successfully"})
}
