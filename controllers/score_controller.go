package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
)

// SubmitScore godoc
// @Summary Submit communication score
// @Description Partner submits feedback for a session (used for insights)
// @Tags Communication
// @Accept json
// @Produce json
// @Param score body models.CommunicationScore true "Score Data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/score [post]
func SubmitScore(c *fiber.Ctx) error {
	var score models.CommunicationScore

	if err := c.BodyParser(&score); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if score.SessionID == "" || score.PartnerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Session ID and Partner ID required"})
	}

	score.CreatedAt = time.Now().Unix()

	sessionCollection := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Determine whether to update scoreA or scoreB
	updateField := "scoreA"
	if score.PartnerID != "" {
		// Check if this PartnerID is B (assumes partnerA and partnerB are saved in session)
		session := struct {
			PartnerA string `bson:"partnerA"`
			PartnerB string `bson:"partnerB"`
		}{}

		err := sessionCollection.FindOne(ctx, fiber.Map{"_id": score.SessionID}).Decode(&session)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to find session"})
		}

		if score.PartnerID == session.PartnerB {
			updateField = "scoreB"
		}
	}

	// Push the score into the session document under scoreA or scoreB
	_, err := sessionCollection.UpdateOne(ctx,
		fiber.Map{"_id": score.SessionID},
		fiber.Map{
			"$set": fiber.Map{
				updateField: score,
			},
		},
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update session with score"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Score saved to session"})
}
