package routes

import (
	"mend/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// 👤 User Management
	api.Post("/register", controllers.RegisterUser) // Create user with name, gender, email
	api.Post("/login", controllers.LoginUser)       // Login with email
	api.Get("/user/:id", controllers.GetUser)       // Fetch user by ID
	api.Post("/invite", controllers.InvitePartner)
	api.Post("/accept-invite", controllers.AcceptInvite)
	// Link two partners

	// 🌱 Onboarding
	api.Post("/onboarding", controllers.SubmitOnboarding) // Add goals, challenges, etc.

	// 🗣️ Voice Session + AI Moderation
	api.Post("/session", controllers.StartSession)
	app.Get("/api/session/active/:userId", controllers.GetActiveSession) // Start session between users
	api.Post("/moderate", controllers.ModerateChat)                      // Moderate message via GPT

	// 🔄 Real-time Chat (WebSocket)
	controllers.SetupWebSocket(app) // GET /ws/:userId

	// 🧘 Post-Session Features
	api.Post("/reflection", controllers.SaveReflection)          // Submit reflection
	api.Post("/post-resolution", controllers.SavePostResolution) // Gratitude/Bonding form
	api.Post("/score", controllers.SubmitScore)                  // Communication scoring
	api.Get("/insights/:userId", controllers.GetInsights)        // Combined insights endpoint
}
