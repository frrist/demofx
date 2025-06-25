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
	// 5. NOW WITH METRICS: Update EVERY constructor call!
	
	// Step 1: Load configuration
	config, err := shared.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Step 2: Create single logger - needs config
	logger := shared.NewLogger(config)
	
	logger.Log("APP", "Loaded configuration")
	logger.Log("APP", "Environment: " + config.App.Environment)
	
	// Step 3: Create metrics collector - NEW DEPENDENCY!
	metrics := shared.NewMetrics(config)
	logger.Log("APP", "Created metrics collector")

	// Step 4: Create database - NOW needs logger, config, AND metrics!
	// BREAKING CHANGE: Had to update constructor call
	db := shared.NewDatabase(logger, config, metrics)

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

	// Step 5: Create user service - NOW needs db, logger, config, AND metrics!
	// BREAKING CHANGE: Had to update constructor call
	userService := shared.NewUserService(db, logger, config, metrics)

	// Step 6: Create and start server - NOW needs service, logger, config, AND metrics!
	// BREAKING CHANGE: Had to update constructor call
	server := shared.NewServer(userService, logger, config, metrics)

	logger.Log("APP", "Traditional setup complete - all dependencies manually wired")
	logger.Log("APP", "Notice: We had to update EVERY constructor call to add metrics!")
	logger.Log("APP", "Try: curl http://" + config.Server.Host + ":" + config.Server.Port + "/user?id=1")
	logger.Log("APP", "Config: curl http://" + config.Server.Host + ":" + config.Server.Port + "/config")
	logger.Log("APP", "Metrics: curl http://" + config.Server.Host + ":" + config.Server.Port + "/metrics")

	// Server runs forever (blocking)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
