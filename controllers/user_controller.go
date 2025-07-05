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
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/register [post]
func RegisterUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if user.Name == "" || user.Gender == "" || user.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing name, gender, or email"})
	}

	user.ID = utils.GeneratePartnerID()
	user.ColorCode = "blue" // Default color for first user

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if email already exists
	count, _ := collection.CountDocuments(ctx, fiber.Map{"email": user.Email})
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Email already registered"})
	}

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save user"})
	}

	return c.Status(201).JSON(user)
}

// LoginUser godoc
// @Summary Login a user
// @Description Logs in user by email lookup.
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Login credentials"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/login [post]
func LoginUser(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email string `json:"email"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid login payload"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, fiber.Map{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Returns user information by user ID.
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/user/{id} [get]
func GetUser(c *fiber.Ctx) error {
	userId := c.Params("id")
	if userId == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, fiber.Map{"id": userId}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// InvitePartner godoc
// @Summary Link partners
// @Description Links two users as partners by ID and stores inviter.
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

	// Update inviter (YourID) with partnerId
	_, err := collection.UpdateOne(ctx,
		map[string]interface{}{"id": body.YourID},
		map[string]interface{}{"$set": map[string]string{"partnerId": body.PartnerID}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update inviter"})
	}

	// Update invitee (PartnerID) with partnerId and invitedBy
	_, err = collection.UpdateOne(ctx,
		map[string]interface{}{"id": body.PartnerID},
		map[string]interface{}{"$set": map[string]string{
			"partnerId": body.YourID,
			"invitedBy": body.YourID, // ✅ Store who invited them
		}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update partner"})
	}

	return c.Status(200).JSON(fiber.Map{"message": "Partners linked successfully"})
}
