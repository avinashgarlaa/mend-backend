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

// StartSession godoc
// @Summary      Start a new session between partners
// @Description  Creates a session document in MongoDB and notifies partner via email
// @Tags         Session
// @Accept       json
// @Produce      json
// @Param        session body models.Session true "Session Info"
// @Success      201 {object} models.Session
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/session [post]
func StartSession(c *fiber.Ctx) error {
	var session models.Session
	if err := c.BodyParser(&session); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid session data"})
	}
	if session.PartnerA == "" || session.PartnerB == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Both partner IDs required"})
	}

	// Session setup
	session.ID = utils.GeneratePartnerID()
	session.CreatedAt = time.Now().Unix()
	session.Resolved = false
	session.Messages = []models.Message{}
	session.ScoreA = models.CommunicationScore{}
	session.ScoreB = models.CommunicationScore{}

	// Insert session
	collection := database.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, session)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create session"})
	}

	// Fetch partnerB's email and PartnerA's name
	usersColl := database.GetCollection("users")

	var partnerB, partnerA models.User
	errA := usersColl.FindOne(ctx, bson.M{"id": session.PartnerA}).Decode(&partnerA)
	errB := usersColl.FindOne(ctx, bson.M{"id": session.PartnerB}).Decode(&partnerB)
	if errA == nil && errB == nil {
		subject := "New Mend Session Started 💬"
		body := fmt.Sprintf(`
			<h2>Hi %s,</h2>
			<p><strong>%s</strong> has started a new Mend session with you.</p>
			<p>Please open the app to join and continue your conversation.</p>
			<p><i>Session ID:</i> <strong>%s</strong></p>
			<br/>
			<p>With love,<br/>The Mend Team</p>
		`, partnerB.Name, partnerA.Name, session.ID)

		go utils.SendEmail(partnerB.Email, subject, body)
	}

	return c.Status(201).JSON(session)
}

// GetActiveSession godoc
// @Summary Get active (unresolved) session for a user
// @Tags Session
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} models.Session
// @Failure 404 {object} map[string]string
// @Router /api/session/active/{userId} [get]
func GetActiveSession(c *fiber.Ctx) error {
	userId := c.Params("userId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"partnerA": userId},
			{"partnerB": userId},
		},
		"resolved": false,
	}

	var session models.Session
	err := database.GetCollection("sessions").FindOne(ctx, filter).Decode(&session)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "No active session"})
	}
	return c.JSON(session)
}

// EndSession godoc
// @Summary      Mark a session as resolved
// @Description  Updates the session's resolved field to true and notifies partner via email
// @Tags         Session
// @Accept       json
// @Produce      json
// @Param        sessionId path string true "Session ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/session/end/{sessionId} [patch]
func EndSession(c *fiber.Ctx) error {
	sessionId := c.Params("sessionId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": sessionId}
	update := bson.M{"$set": bson.M{"resolved": true}}

	// Update session
	sessionsColl := database.GetCollection("sessions")
	result, err := sessionsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to end session"})
	}
	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Session not found"})
	}

	// Fetch session to send email
	var session models.Session
	err = sessionsColl.FindOne(ctx, filter).Decode(&session)
	if err == nil {
		usersColl := database.GetCollection("users")

		var partnerA, partnerB models.User
		errA := usersColl.FindOne(ctx, bson.M{"id": session.PartnerA}).Decode(&partnerA)
		errB := usersColl.FindOne(ctx, bson.M{"id": session.PartnerB}).Decode(&partnerB)
		if errA == nil && errB == nil {
			subject := "Your Mend Session Has Ended 💜"
			body := fmt.Sprintf(`
				<h2>Hi %s,</h2>
				<p>Your session with <strong>%s</strong> has just ended.</p>
				<p>You can review insights and reflections inside the app.</p>
				<p><i>Session ID:</i> <strong>%s</strong></p>
				<br/>
				<p>With warmth,<br/>The Mend Team</p>
			`, partnerB.Name, partnerA.Name, session.ID)

			go utils.SendEmail(partnerB.Email, subject, body)
		}
	}

	return c.JSON(fiber.Map{"message": "Session ended successfully"})
}
