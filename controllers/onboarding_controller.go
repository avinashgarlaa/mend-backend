package controllers

import (
	"context"
	"time"

	"mend/database"

	"github.com/gofiber/fiber/v2"
)

// SaveOnboarding godoc
// @Summary Save onboarding responses
// @Tags Users
// @Accept json
// @Produce json
// @Param data body map[string]interface{} true "Onboarding Data"
// @Success 201 {object} map[string]string
// @Router /api/onboarding [post]
func SaveOnboarding(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	collection := database.GetCollection("onboarding")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB insert failed"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Onboarding saved"})
}
