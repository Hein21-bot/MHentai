package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"mhentai-backend/internal/database"
	"mhentai-backend/internal/router"
	"mhentai-backend/internal/storage"
)

func main() {
	// Load .env if present (silently ignored if missing)
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env")
	}

	port := getEnv("PORT", "8080")
	adminToken := getEnv("ADMIN_TOKEN", "admin123")

	database.Init()
	storage.InitR2()

	r := router.New(adminToken)
	log.Printf("Server starting on port %s (admin token: %s)", port, adminToken)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
