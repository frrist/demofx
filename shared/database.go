package shared

import (
	"fmt"
	"time"
)

// Database provides mock database functionality
type Database struct {
	logger         *Logger
	config         *DatabaseConfig
	users          map[string]string
	cache          map[string]string
	cacheEnabled   bool
}

// NewDatabase creates a new database instance
func NewDatabase(logger *Logger, config *Config) *Database {
	return &Database{
		logger:       logger,
		config:       &config.Database,
		cacheEnabled: config.App.Features["cache_enabled"],
		users: map[string]string{
			"1": "Alice",
			"2": "Bob",
			"3": "Charlie",
		},
		cache: make(map[string]string),
	}
}

// Initialize sets up the database connection (mock)
func (d *Database) Initialize() error {
	d.logger.Log("DATABASE", fmt.Sprintf("Initializing database with max connections: %d, timeout: %ds", 
		d.config.MaxConnections, d.config.Timeout))
	
	if d.cacheEnabled {
		d.logger.Log("DATABASE", fmt.Sprintf("Cache enabled with size: %d", d.config.CacheSize))
	}
	
	// Mock initialization with timeout
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Close shuts down the database connection
func (d *Database) Close() error {
	d.logger.Log("DATABASE", "Closing database connection...")
	// Mock cleanup logic
	return nil
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(id string) (string, error) {
	// Check cache first if enabled
	if d.cacheEnabled {
		if cached, ok := d.cache[id]; ok {
			d.logger.Log("DATABASE", fmt.Sprintf("Cache hit for user ID: %s", id))
			return cached, nil
		}
	}
	
	d.logger.Log("DATABASE", fmt.Sprintf("Fetching user with ID: %s from database", id))
	
	// Simulate database query with configured timeout
	time.Sleep(50 * time.Millisecond)
	
	if name, ok := d.users[id]; ok {
		// Store in cache if enabled
		if d.cacheEnabled && len(d.cache) < d.config.CacheSize {
			d.cache[id] = name
			d.logger.Log("DATABASE", fmt.Sprintf("Cached user %s", id))
		}
		return name, nil
	}
	return "", fmt.Errorf("user not found")
}