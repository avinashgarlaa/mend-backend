package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
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

// controllers/session_controller.go

// GetActiveSession godoc
// @Summary      Retrieve active (unresolved) session for a user
// @Tags         Session
// @Produce      json
// @Param        userId path string true "User ID"
// @Success      200 {object} models.Session
// @Failure      404 {object} map[string]string
// @Router       /api/session/active/:userId [get]
func GetActiveSession(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	sessions := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var session models.Session
	err := sessions.FindOne(ctx, fiber.Map{
		"$or": []fiber.Map{
			{"partnerA": userId},
			{"partnerB": userId},
		},
		"resolved": false,
	}).Decode(&session)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No active session found"})
	}

	return c.JSON(session)
}
