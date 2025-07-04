package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
)

// SubmitOnboarding godoc
// @Summary      Submit onboarding form
// @Description  Stores onboarding details for a user
// @Tags         Onboarding
// @Accept       json
// @Produce      json
// @Param        onboarding body models.Onboarding true "Onboarding Data"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/onboarding [post]
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Onboarding data submitted successfully",
	})
}
