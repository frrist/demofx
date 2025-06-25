package main

import (
	"log"

	"github.com/frrist/demofx/shared"
)

func main() {
	// Traditional approach: Manual dependency wiring with configuration
	// Notice how we have to:
	// 1. Load configuration first
	// 2. Pass config AND logger to EVERY component
	// 3. Create each dependency in the correct order
	// 4. Handle initialization and cleanup manually
	
	// Step 1: Load configuration
	config, err := shared.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Step 2: Create single logger - needs config
	logger := shared.NewLogger(config)
	
	logger.Log("APP", "Loaded configuration")
	logger.Log("APP", "Environment: " + config.App.Environment)

	// Step 3: Create database - needs logger AND config
	db := shared.NewDatabase(logger, config)

	// Manual initialization
	if err := db.Initialize(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Manual cleanup - easy to forget!
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Step 4: Create user service - needs db, logger, AND config
	userService := shared.NewUserService(db, logger, config)

	// Step 5: Create and start server - needs service, logger, AND config
	server := shared.NewServer(userService, logger, config)

	logger.Log("APP", "Traditional setup complete - all dependencies manually wired")
	logger.Log("APP", "Notice: We had to pass both logger and config to EVERY component")
	logger.Log("APP", "Try: curl http://" + config.Server.Host + ":" + config.Server.Port + "/user?id=1")
	logger.Log("APP", "Config: curl http://" + config.Server.Host + ":" + config.Server.Port + "/config")

	// Server runs forever (blocking)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
