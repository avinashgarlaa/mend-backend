package controllers

import (
	"context"
	"time"

	"mend/database"
	"mend/models"
	"mend/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Creates a user with basic details
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user body map[string]string true "Name, Email, Password"
// @Success      201 {object} models.User
// @Failure      400,409,500 {object} map[string]string
// @Router       /api/register [post]

func RegisterUser(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" || input.Email == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check for duplicate email
	count, _ := collection.CountDocuments(ctx, fiber.Map{"email": input.Email})
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Email already registered"})
	}

	hashed := utils.HashPassword(input.Password)

	user := models.User{
		ID:        utils.GeneratePartnerID(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  hashed,
		ColorCode: "blue",
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save user"})
	}

	user.Password = "" // Hide before returning
	return c.Status(201).JSON(user)
}

// SubmitOnboarding godoc
// @Summary      Submit onboarding data
// @Description  Adds gender, goals, challenges to a user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        onboarding body models.User true "User Onboarding"
// @Success      200 {object} map[string]string
// @Failure      400,404,500 {object} map[string]string
// @Router       /api/onboarding [post]
func SubmitOnboarding(c *fiber.Ctx) error {
	var data models.User
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if data.ID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := fiber.Map{
		"gender":         data.Gender,
		"goals":          data.Goals,
		"otherGoal":      data.OtherGoal,
		"challenges":     data.Challenges,
		"otherChallenge": data.OtherChallenge,
	}

	res, err := collection.UpdateOne(ctx, fiber.Map{"id": data.ID}, fiber.Map{"$set": update})
	if err != nil || res.ModifiedCount == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	return c.JSON(fiber.Map{"message": "Onboarding complete"})
}

// LoginUser godoc
// @Summary Login a user
// @Description Logs in user by email and password
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Login credentials (email & password)"
// @Success 200 {object} models.User
// @Failure 400,401,404 {object} map[string]string
// @Router /api/login [post]
func LoginUser(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Password == "" {
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

	// ✅ Validate password using bcrypt
	if !utils.CheckPassword(req.Password, user.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
	}

	user.Password = "" // Hide password before returning
	return c.JSON(user)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Fetches user info by user ID
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing userId",
		})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"id": userId}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	user.Password = "" // 🔐 Hide password if any (optional field)

	return c.JSON(user)
}

// InvitePartner godoc
// @Summary Link two users as partners
// @Description Stores partnership and inviter reference
// @Tags Users
// @Accept json
// @Produce json
// @Param invite body map[string]string true "Invite info: yourId, partnerId"
// @Success 200 {object} map[string]string
// @Failure 400,404,500 {object} map[string]string
// @Router /api/invite [post]
func InvitePartner(c *fiber.Ctx) error {
	type InviteRequest struct {
		YourID    string `json:"yourId"`
		PartnerID string `json:"partnerId"`
	}

	var body InviteRequest
	if err := c.BodyParser(&body); err != nil || body.YourID == "" || body.PartnerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid invite payload"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if both users exist
	var inviter, invitee models.User
	err1 := collection.FindOne(ctx, fiber.Map{"id": body.YourID}).Decode(&inviter)
	err2 := collection.FindOne(ctx, fiber.Map{"id": body.PartnerID}).Decode(&invitee)
	if err1 != nil || err2 != nil {
		return c.Status(404).JSON(fiber.Map{"error": "One or both users not found"})
	}

	// Update inviter
	_, err := collection.UpdateOne(ctx,
		fiber.Map{"id": body.YourID},
		fiber.Map{"$set": fiber.Map{"partnerId": body.PartnerID}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update inviter"})
	}

	// Update invitee
	_, err = collection.UpdateOne(ctx,
		fiber.Map{"id": body.PartnerID},
		fiber.Map{"$set": fiber.Map{
			"partnerId": body.YourID,
			"invitedBy": body.YourID,
		}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update partner"})
	}

	return c.JSON(fiber.Map{"message": "Partners linked successfully"})
}
