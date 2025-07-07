package controllers

import (
	"context"
	"fmt"
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
// @Param        user body map[string]string true "Name, Email, Password, Gender"
// @Success      201 {object} models.User
// @Failure      400,409,500 {object} map[string]string
// @Router       /api/register [post]
func RegisterUser(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Gender   string `json:"gender"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" || input.Email == "" || input.Password == "" || input.Gender == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
		Gender:    input.Gender,
		ColorCode: "blue",
		CreatedAt: time.Now(),
	}

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save user"})
	}

	user.Password = "" // Don't return password
	return c.Status(201).JSON(user)
}

// SubmitOnboarding godoc
// @Summary      Submit onboarding data
// @Description  Adds goals and challenges to a user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        onboarding body map[string]interface{} true "User Onboarding Data"
// @Success      200 {object} map[string]string
// @Failure      400,404,500 {object} map[string]string
// @Router       /api/onboarding [post]
func SubmitOnboarding(c *fiber.Ctx) error {
	var data struct {
		UserID         string   `json:"userId"`
		Goals          []string `json:"goals"`
		OtherGoal      string   `json:"otherGoal"`
		Challenges     []string `json:"challenges"`
		OtherChallenge string   `json:"otherChallenge"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if data.UserID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing userId"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := fiber.Map{
		"goals":          data.Goals,
		"otherGoal":      data.OtherGoal,
		"challenges":     data.Challenges,
		"otherChallenge": data.OtherChallenge,
	}

	res, err := collection.UpdateOne(ctx, fiber.Map{"id": data.UserID}, fiber.Map{"$set": update})
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

	// Fetch inviter and invitee
	var inviter, invitee models.User
	if err := collection.FindOne(ctx, fiber.Map{"id": body.YourID}).Decode(&inviter); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Inviter not found"})
	}
	if err := collection.FindOne(ctx, fiber.Map{"id": body.PartnerID}).Decode(&invitee); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invitee not found"})
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

	// ✅ Send Email to invitee
	subject := "You’ve Been Invited to Mend 💜"
	bodyHTML := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p><strong>%s</strong> has invited you to join them on Mend, a space to improve communication and connection.</p>
		<p>Your Partner ID: <strong>%s</strong></p>
		<p>Open the app and enter this Partner ID to link accounts.</p>
		<br/>
		<p>With love,<br/>The Mend Team</p>
	`, invitee.Name, inviter.Name, inviter.ID)

	go utils.SendEmail(invitee.Email, subject, bodyHTML)

	return c.JSON(fiber.Map{"message": "Partners linked and invitation email sent"})
}

// AcceptInvite godoc
// @Summary Accept an invitation
// @Description Accepts a partner invite (both users must already exist)
// @Tags Users
// @Accept json
// @Produce json
// @Param accept body map[string]string true "Accept info: yourId, partnerId"
// @Success 200 {object} map[string]string
// @Failure 400,404,500 {object} map[string]string
// @Router /api/accept-invite [post]
func AcceptInvite(c *fiber.Ctx) error {
	type AcceptRequest struct {
		YourID    string `json:"yourId"`
		PartnerID string `json:"partnerId"`
	}

	var body AcceptRequest
	if err := c.BodyParser(&body); err != nil || body.YourID == "" || body.PartnerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify both users exist
	var you, partner models.User
	err1 := collection.FindOne(ctx, fiber.Map{"id": body.YourID}).Decode(&you)
	err2 := collection.FindOne(ctx, fiber.Map{"id": body.PartnerID}).Decode(&partner)

	if err1 != nil || err2 != nil {
		return c.Status(404).JSON(fiber.Map{"error": "One or both users not found"})
	}

	// Verify that inviter already linked you
	if partner.PartnerID != body.YourID {
		return c.Status(400).JSON(fiber.Map{"error": "Invite not initiated by this partner"})
	}

	// Accept the invite: mutual linkage
	_, err := collection.UpdateOne(ctx,
		fiber.Map{"id": body.YourID},
		fiber.Map{"$set": fiber.Map{
			"partnerId": body.PartnerID,
			"invitedBy": body.PartnerID,
		}},
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to accept invite"})
	}

	return c.JSON(fiber.Map{"message": "Invite accepted. You're now linked!"})
}
