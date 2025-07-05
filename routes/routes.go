// routes/routes.go

package routes

import (
	"mend/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// 👤 User & Onboarding
	api.Post("/register", controllers.RegisterUser)
	api.Post("/invite", controllers.InvitePartner)
	api.Post("/onboarding", controllers.SubmitOnboarding)
	api.Post("/login", controllers.LoginUser)
	api.Get("/user/:id", controllers.GetUser) // FIXED: "/api/user/:id" → redundant

	// 🎙️ Session & AI Chat
	api.Post("/session", controllers.StartSession)
	api.Post("/moderate", controllers.ModerateChat)

	// 🔄 Real-time WebSocket Messaging (outside /api group)
	controllers.SetupWebSocket(app) // GET /ws/:userId

	// 🧘 Post-Session Flow
	api.Post("/reflection", controllers.SaveReflection)
	api.Post("/post-resolution", controllers.SavePostResolution)
	api.Post("/score", controllers.SubmitScore)
	api.Get("/insights/:userId", controllers.GetInsights)
}
