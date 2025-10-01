package main

import (
	"log"
	"os"

	"github.com/Bug-Bugger/ezmodel/internal/api/server"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	// Determine which .env file to load based on environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "development" // Default to development
	}

	var envFile string
	if env == "production" {
		envFile = "../.env.prod"
	} else {
		envFile = "../.env.dev"
	}

	// Load environment-specific .env file from project root
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Warning: No %s file found or error loading it. Using default values or environment variables.", envFile)
	}

	// Load configuration
	cfg := config.New()

	// Connect to database
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the SQL DB instance from GORM
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer sqlDB.Close()

	// Initialize and start server
	srv := server.New(cfg, database)
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
