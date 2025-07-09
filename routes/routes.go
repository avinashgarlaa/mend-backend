package routes

import (
	"mend/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// 👤 User Management
	api.Post("/register", controllers.RegisterUser)
	api.Post("/login", controllers.LoginUser)
	api.Get("/user/:id", controllers.GetUser)
	api.Post("/invite", controllers.InvitePartner)
	api.Post("/accept-invite", controllers.AcceptInvite)

	// 🌱 Onboarding
	api.Post("/onboarding", controllers.SubmitOnboarding)

	// 🗣️ Voice Session + AI Moderation
	api.Post("/session", controllers.StartSession)
	api.Get("/session/active/:userId", controllers.GetActiveSession)
	api.Patch("/session/end/:sessionId", controllers.EndSession)
	api.Post("/moderate", controllers.ModerateChat)

	// 🔄 WebSocket Chat
	app.Use("/ws/:userId/:sessionId", websocket.New(controllers.HandleWebSocket))

	// 🧘 Post-Session Features
	api.Post("/reflection", controllers.SaveReflection)
	api.Post("/post-resolution", controllers.SavePostResolution)
	api.Post("/score", controllers.SubmitScore)
	api.Get("/insights/:userId", controllers.GetInsights)
}
