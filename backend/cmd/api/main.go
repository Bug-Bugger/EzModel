package main

import (
	"log"

	"github.com/Bug-Bugger/ezmodel/internal/api/server"
	"github.com/Bug-Bugger/ezmodel/internal/config"
)

func main() {
	cfg := config.New()

	// Initialize and start server
	srv := server.New(cfg)
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
