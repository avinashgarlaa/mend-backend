package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
)

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user account in the system.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.User true "User Info"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Router /api/register [post]
func RegisterUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user.ID = utils.GeneratePartnerID()
	user.ColorCode = "blue" // Default for Partner A — you can flip this based on order

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save user"})
	}

	return c.Status(201).JSON(user)
}

// InvitePartner godoc
// @Summary Send an invite code to partner
// @Description Links two users as partners in the system.
// @Tags Users
// @Accept json
// @Produce json
// @Param invite body map[string]string true "Invite Info (yourId and partnerId)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/invite [post]
func InvitePartner(c *fiber.Ctx) error {
	type InviteRequest struct {
		YourID    string `json:"yourId"`
		PartnerID string `json:"partnerId"`
	}

	var body InviteRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid invite payload"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update both users with partner IDs
	_, err := collection.UpdateOne(ctx,
		map[string]interface{}{"id": body.YourID},
		map[string]interface{}{"$set": map[string]string{"partnerId": body.PartnerID}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update inviter"})
	}

	_, err = collection.UpdateOne(ctx,
		map[string]interface{}{"id": body.PartnerID},
		map[string]interface{}{"$set": map[string]string{"partnerId": body.YourID}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update partner"})
	}

	return c.Status(200).JSON(fiber.Map{"message": "Partners linked successfully"})
}
