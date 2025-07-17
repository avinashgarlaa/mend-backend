package routes

import (
	"mend/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// SetupRoutes defines all backend API and WebSocket routes
func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// ─────────────────────────────────────────────
	// 👤 User Authentication & Relationship
	// ─────────────────────────────────────────────
	api.Post("/register", controllers.RegisterUser)
	api.Post("/login", controllers.LoginUser)
	api.Get("/user/:id", controllers.GetUser)
	api.Post("/invite", controllers.InvitePartner)
	api.Post("/accept-invite", controllers.AcceptInvite)

	// ─────────────────────────────────────────────
	// 🌱 Onboarding Data
	// ─────────────────────────────────────────────
	api.Post("/onboarding", controllers.SubmitOnboarding)

	// ─────────────────────────────────────────────
	// 🗣️ Session Management + AI Moderation
	// ─────────────────────────────────────────────
	api.Post("/session", controllers.StartSession)
	api.Get("/session/active/:userId", controllers.GetActiveSession)
	api.Patch("/session/end/:sessionId", controllers.EndSession)
	api.Get("/session/score/:sessionId", controllers.GetSessionScore)
	api.Post("/moderate", controllers.ModerateChat)

	// ─────────────────────────────────────────────
	// 🔄 WebSocket Chat Communication
	// ─────────────────────────────────────────────
	// Legacy text-based chat WebSocket
	app.Get("/ws/:userId/:sessionId", websocket.New(controllers.HandleWebSocket))

	// New voice chat + AI moderation WebSocket
	app.Use("/ws-voice/:sessionId/:userId", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws-voice/:sessionId/:userId", websocket.New(controllers.WebSocketHandler2))

	// ─────────────────────────────────────────────
	// 🧘 Post-Session Reflections & Scores
	// ─────────────────────────────────────────────
	api.Post("/reflection", controllers.SaveReflection)
	api.Post("/post-resolution", controllers.SavePostResolution)
	api.Post("/score", controllers.SubmitScore)

	// ─────────────────────────────────────────────
	// 📊 Communication Insights
	// ─────────────────────────────────────────────
	api.Get("/insights/:userId", controllers.GetInsights)
}
