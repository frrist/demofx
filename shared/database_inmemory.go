package shared

import (
	"fmt"
	"time"
)

// InMemoryDatabase provides in-memory database functionality
type InMemoryDatabase struct {
	logger         *Logger
	config         *DatabaseConfig
	metrics        *Metrics
	users          map[string]string
	cache          map[string]string
	cacheEnabled   bool
}

// NewInMemoryDatabase creates a new in-memory database instance
func NewInMemoryDatabase(logger *Logger, config *Config, metrics *Metrics) *InMemoryDatabase {
	return &InMemoryDatabase{
		logger:       logger,
		config:       &config.Database,
		metrics:      metrics,
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
func (d *InMemoryDatabase) Initialize() error {
	d.logger.Log("DATABASE", fmt.Sprintf("Initializing IN-MEMORY database with max connections: %d, timeout: %ds", 
		d.config.MaxConnections, d.config.Timeout))
	
	if d.cacheEnabled {
		d.logger.Log("DATABASE", fmt.Sprintf("Cache enabled with size: %d", d.config.CacheSize))
	}
	
	// Mock initialization with timeout
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Close shuts down the database connection
func (d *InMemoryDatabase) Close() error {
	d.logger.Log("DATABASE", "Closing database connection...")
	// Mock cleanup logic
	return nil
}

// GetUser retrieves a user by ID
func (d *InMemoryDatabase) GetUser(id string) (string, error) {
	// Track the database query
	if d.metrics != nil {
		d.metrics.RecordDBQuery()
	}
	
	// Check cache first if enabled
	if d.cacheEnabled {
		if cached, ok := d.cache[id]; ok {
			d.logger.Log("DATABASE", fmt.Sprintf("Cache hit for user ID: %s", id))
			if d.metrics != nil {
				d.metrics.RecordCacheHit()
			}
			return cached, nil
		}
		// Cache miss
		if d.metrics != nil {
			d.metrics.RecordCacheMiss()
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