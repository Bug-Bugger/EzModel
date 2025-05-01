package main

import (
	"log"

	"github.com/Bug-Bugger/ezmodel/internal/api/server"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/db"
)

func main() {
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
