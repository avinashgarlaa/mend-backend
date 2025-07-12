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

	// 🧠 AI moderation endpoint (tone, interrupt detection)
	api.Post("/moderate", controllers.ModerateChat)

	// ─────────────────────────────────────────────
	// 🔄 WebSocket Chat Communication
	// ─────────────────────────────────────────────
	app.Use("/ws/:userId/:sessionId", websocket.New(controllers.HandleWebSocket))

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

	// Optional: Health Check or versioning endpoint
	// api.Get("/health", func(c *fiber.Ctx) error {
	// 	return c.SendString("Mend API is running")
	// })
}
