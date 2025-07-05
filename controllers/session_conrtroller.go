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
