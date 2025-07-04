package routes

import (
	"mend/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// User routes
	api.Post("/register", controllers.RegisterUser)
	api.Post("/invite", controllers.InvitePartner)

	// Onboarding route
	api.Post("/onboarding", controllers.SubmitOnboarding)

	api.Post("/score", controllers.SubmitScore)
	api.Post("/post-resolution", controllers.SavePostResolution)

	// Chat session & AI moderation
	api.Post("/session", controllers.StartSession)
	api.Post("/moderate", controllers.ModerateChat)

	// Reflections & Insights
	api.Post("/reflection", controllers.SaveReflection)
	api.Get("/insights/:userId", controllers.GetInsights)
}
