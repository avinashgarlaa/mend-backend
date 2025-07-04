package main

import (
	"mend/config"
	"mend/database"
	"mend/routes"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load .env variables
	config.LoadEnv()

	// Connect to MongoDB
	database.ConnectDB()

	// Initialize Fiber app
	app := fiber.New()

	// Setup all API routes
	routes.SetupRoutes(app)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	app.Listen(":" + port)
}
