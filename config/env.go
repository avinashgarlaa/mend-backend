package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Only load from .env if not in Render (or production)
	if os.Getenv("RENDER") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ .env file not found (expected in production)")
		} else {
			log.Println("✅ Loaded .env for local dev")
		}
	}
}
