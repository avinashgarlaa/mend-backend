package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

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

func SubmitOnboarding(c *fiber.Ctx) error {
	var data models.Onboarding

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	collection := database.GetCollection("onboarding")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save onboarding data"})
	}

	return c.JSON(fiber.Map{
		"message": "Onboarding data submitted successfully",
	})
}
