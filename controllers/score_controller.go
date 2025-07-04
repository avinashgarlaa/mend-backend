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
// @Tags Communication
// @Accept json
// @Produce json
// @Param score body models.CommunicationScore true "Score Data"
// @Success 201 {object} map[string]string
// @Router /api/score [post]
func SubmitScore(c *fiber.Ctx) error {
	var score models.CommunicationScore

	if err := c.BodyParser(&score); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	score.CreatedAt = time.Now().Unix()

	collection := database.GetCollection("scores")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, score)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save score"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Score submitted"})
}
