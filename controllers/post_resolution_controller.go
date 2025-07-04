package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"

	"github.com/gofiber/fiber/v2"
)

// SavePostResolution godoc
// @Summary Save post-resolution reflection
// @Tags Reflection
// @Accept json
// @Produce json
// @Param data body models.PostResolution true "Post-resolution data"
// @Success 201 {object} map[string]string
// @Router /api/post-resolution [post]
func SavePostResolution(c *fiber.Ctx) error {
	var data models.PostResolution

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
	}

	data.Timestamp = time.Now().Unix()
	collection := database.GetCollection("postResolution")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save post-resolution"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Saved successfully"})
}
