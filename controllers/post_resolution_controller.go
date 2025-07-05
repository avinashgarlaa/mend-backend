package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
)

// SavePostResolution godoc
// @Summary      Save post-resolution reflection
// @Description  Stores emotional bonding data after a session
// @Tags         Reflection
// @Accept       json
// @Produce      json
// @Param        data body models.PostResolution true "Post-resolution data"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/post-resolution [post]
func SavePostResolution(c *fiber.Ctx) error {
	var data models.PostResolution

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid data format"})
	}

	// Validate required fields
	if data.UserID == "" || data.SessionID == "" || data.Gratitude == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	data.Timestamp = time.Now().Unix()

	collection := database.GetCollection("postResolution")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save post-resolution"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Post-resolution saved successfully"})
}
