package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
)

// SaveReflection godoc
// @Summary Submit a session reflection
// @Description Stores a post-session reflection entry from a user
// @Tags Reflections
// @Accept json
// @Produce json
// @Param reflection body models.Reflection true "Reflection Data"
// @Success 201 {object} models.Reflection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/reflection [post]
func SaveReflection(c *fiber.Ctx) error {
	var reflection models.Reflection
	if err := c.BodyParser(&reflection); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid reflection data"})
	}

	reflection.ID = reflection.SessionID + "-" + reflection.UserID
	reflection.Timestamp = time.Now().Unix()

	collection := database.GetCollection("reflections")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, reflection)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save reflection"})
	}

	return c.Status(201).JSON(reflection)
}

// GetInsights godoc
// @Summary Get communication insights for a user
// @Description Returns all sessions and reflections related to a user
// @Tags Insights
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/insights/{userId} [get]
func GetInsights(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	sessionCollection := database.GetCollection("sessions")
	reflectionCollection := database.GetCollection("reflections")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := sessionCollection.Find(ctx, fiber.Map{
		"$or": []fiber.Map{
			{"partnerA": userId},
			{"partnerB": userId},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching sessions"})
	}

	var sessions []models.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode sessions"})
	}

	cursor2, err := reflectionCollection.Find(ctx, fiber.Map{"userId": userId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching reflections"})
	}

	var reflections []models.Reflection
	if err := cursor2.All(ctx, &reflections); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode reflections"})
	}

	return c.Status(200).JSON(fiber.Map{
		"sessions":    sessions,
		"reflections": reflections,
	})
}
