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

	if data.UserID == "" || data.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// ✅ Check if user exists
	userColl := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userColl.FindOne(ctx, fiber.Map{"id": data.UserID}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// ✅ Check for existing onboarding for that user
	collection := database.GetCollection("onboarding")
	count, _ := collection.CountDocuments(ctx, fiber.Map{"userId": data.UserID})
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Onboarding already submitted for this user"})
	}

	// ✅ Add timestamp
	data.CreatedAt = time.Now()

	// ✅ Save onboarding
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save onboarding data"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Onboarding data submitted successfully",
	})
}
