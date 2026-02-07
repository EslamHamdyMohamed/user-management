package main

import (
	"log"

	"user-management/internal/api"
	"user-management/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and setup server
	server := api.NewServer(cfg)

	if err := server.Setup(); err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}

	// Run server with graceful shutdown
	if err := server.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
