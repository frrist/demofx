package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistentDatabase provides file-based persistent database functionality
// This implementation demonstrates an alternative to the in-memory database
type PersistentDatabase struct {
	logger       *Logger
	config       *DatabaseConfig
	metrics      *Metrics
	mu           sync.RWMutex
	dataFile     string
	users        map[string]string
	cache        map[string]string
	cacheEnabled bool
}

// NewPersistentDatabase creates a new persistent database instance
func NewPersistentDatabase(logger *Logger, config *Config, metrics *Metrics) *PersistentDatabase {
	return &PersistentDatabase{
		logger:       logger,
		config:       &config.Database,
		metrics:      metrics,
		cacheEnabled: config.App.Features["cache_enabled"],
		dataFile:     filepath.Join(os.TempDir(), "demo_users.json"),
		users:        make(map[string]string),
		cache:        make(map[string]string),
	}
}

// Initialize sets up the database and loads data from file
func (d *PersistentDatabase) Initialize() error {
	d.logger.Log("DATABASE", fmt.Sprintf("Initializing PERSISTENT database with file: %s", d.dataFile))
	d.logger.Log("DATABASE", fmt.Sprintf("Max connections: %d, timeout: %ds", 
		d.config.MaxConnections, d.config.Timeout))
	
	if d.cacheEnabled {
		d.logger.Log("DATABASE", fmt.Sprintf("Cache enabled with size: %d", d.config.CacheSize))
	}
	
	// Try to load existing data
	if err := d.loadData(); err != nil {
		// If file doesn't exist, create initial data
		d.logger.Log("DATABASE", "No existing data found, creating initial dataset")
		d.users = map[string]string{
			"1": "Alice",
			"2": "Bob", 
			"3": "Charlie",
			"4": "Diana",      // Additional users in persistent DB
			"5": "Edward",
			"6": "Fiona",
		}
		// Save initial data
		if err := d.saveData(); err != nil {
			return fmt.Errorf("failed to save initial data: %w", err)
		}
	}
	
	// Simulate longer initialization for persistent DB
	time.Sleep(200 * time.Millisecond)
	return nil
}

// loadData reads user data from file
func (d *PersistentDatabase) loadData() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	data, err := os.ReadFile(d.dataFile)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, &d.users)
}

// saveData writes user data to file
func (d *PersistentDatabase) saveData() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	data, err := json.MarshalIndent(d.users, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(d.dataFile, data, 0644)
}

// Close saves data and shuts down the database
func (d *PersistentDatabase) Close() error {
	d.logger.Log("DATABASE", "Saving data before closing persistent database...")
	if err := d.saveData(); err != nil {
		d.logger.Log("DATABASE", fmt.Sprintf("Error saving data: %v", err))
		return err
	}
	d.logger.Log("DATABASE", "Persistent database closed successfully")
	return nil
}

// GetUser retrieves a user by ID
func (d *PersistentDatabase) GetUser(id string) (string, error) {
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
	
	d.logger.Log("DATABASE", fmt.Sprintf("Fetching user with ID: %s from persistent storage", id))
	
	// Simulate slower persistent database query
	time.Sleep(100 * time.Millisecond)
	
	d.mu.RLock()
	name, ok := d.users[id]
	d.mu.RUnlock()
	
	if ok {
		// Store in cache if enabled
		if d.cacheEnabled && len(d.cache) < d.config.CacheSize {
			d.cache[id] = name
			d.logger.Log("DATABASE", fmt.Sprintf("Cached user %s", id))
		}
		return name, nil
	}
	return "", fmt.Errorf("user not found")
}