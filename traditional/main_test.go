package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/frrist/demofx/shared"
)

// TestUserServiceTraditional demonstrates traditional testing approach
// Notice how we have to:
// 1. Manually create all dependencies in the right order
// 2. Pass mocks through multiple layers
// 3. Handle all wiring ourselves
func TestUserServiceTraditional(t *testing.T) {
	// MANUAL SETUP: Create test config
	config := &shared.Config{
		Server: shared.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
		Database: shared.DatabaseConfig{
			Type:           "inmemory",
			MaxConnections: 10,
			Timeout:        30,
			CacheSize:      100,
		},
		App: shared.AppConfig{
			Environment: "test",
			LogLevel:    "debug",
			Features: map[string]bool{
				"cache_enabled":   false,
				"rate_limiting":   false,
				"metrics_enabled": false,
			},
		},
	}

	// MANUAL SETUP: Create logger
	logger := shared.NewLogger(config)

	// MANUAL SETUP: Create metrics (even though we don't need it for this test)
	metrics := shared.NewMetrics(config)

	// MANUAL SETUP: Create mock database
	mockDB := shared.NewMockDatabase()

	// MANUAL SETUP: Initialize mock (easy to forget!)
	err := mockDB.Initialize()
	require.NoError(t, err)

	// MANUAL SETUP: Create user service with all dependencies
	userService := shared.NewUserService(mockDB, logger, config, metrics)

	// MANUAL SETUP: Create echo instance for testing
	e := echo.New()

	t.Run("successful user lookup", func(t *testing.T) {
		// Setup HTTP request
		req := httptest.NewRequest(http.MethodGet, "/user?id=test1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := userService.GetUserHandler(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Test User 1")
		assert.Equal(t, "test1", mockDB.LastRequestedID)
		assert.Equal(t, 1, mockDB.GetUserCalls)
	})

	t.Run("user not found", func(t *testing.T) {
		// Reset call count
		mockDB.GetUserCalls = 0

		// Setup HTTP request
		req := httptest.NewRequest(http.MethodGet, "/user?id=nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := userService.GetUserHandler(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "User not found")
		assert.Equal(t, "nonexistent", mockDB.LastRequestedID)
		assert.Equal(t, 1, mockDB.GetUserCalls)
	})

	t.Run("missing user id", func(t *testing.T) {
		// Setup HTTP request without ID
		req := httptest.NewRequest(http.MethodGet, "/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute
		err := userService.GetUserHandler(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Missing user ID")
	})

	// MANUAL CLEANUP: Don't forget to close!
	err = mockDB.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, mockDB.CloseCalls)
}

// TestDatabaseSwappingTraditional shows how testing different implementations is complex
func TestDatabaseSwappingTraditional(t *testing.T) {
	config := &shared.Config{
		Server: shared.ServerConfig{Host: "localhost", Port: "8080"},
		Database: shared.DatabaseConfig{
			Type:           "inmemory", // We'll change this
			MaxConnections: 10,
			Timeout:        30,
			CacheSize:      100,
		},
		App: shared.AppConfig{
			Environment: "test",
			LogLevel:    "debug",
			Features:    map[string]bool{"cache_enabled": false},
		},
	}

	logger := shared.NewLogger(config)
	metrics := shared.NewMetrics(config)

	// TEST COMPLEXITY: We need to duplicate the database selection logic from main!
	testCases := []struct {
		name   string
		dbType string
		setup  func() shared.Database
	}{
		{
			name:   "mock database",
			dbType: "mock",
			setup: func() shared.Database {
				return shared.NewMockDatabase()
			},
		},
		{
			name:   "in-memory database",
			dbType: "inmemory",
			setup: func() shared.Database {
				return shared.NewInMemoryDatabase(logger, config, metrics)
			},
		},
		// Can't easily test persistent without side effects!
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// MANUAL: Update config
			config.Database.Type = tc.dbType

			// MANUAL: Create database based on type
			db := tc.setup()

			// MANUAL: Initialize
			err := db.Initialize()
			require.NoError(t, err)

			// MANUAL: Create service with database
			userService := shared.NewUserService(db, logger, config, metrics)

			// Test the service
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/user?id=1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = userService.GetUserHandler(c)
			assert.NoError(t, err)

			// MANUAL: Cleanup
			err = db.Close()
			assert.NoError(t, err)
		})
	}
}

// TestIntegrationSetupTraditional shows the pain of integration testing
func TestIntegrationSetupTraditional(t *testing.T) {
	// PROBLEM: To test the server, we need to wire EVERYTHING manually!
	
	config := &shared.Config{
		Server:   shared.ServerConfig{Host: "localhost", Port: "0"}, // Use port 0 for testing
		Database: shared.DatabaseConfig{Type: "mock"},
		App: shared.AppConfig{
			Environment: "test",
			Features:    map[string]bool{"metrics_enabled": true},
		},
	}

	logger := shared.NewLogger(config)
	metrics := shared.NewMetrics(config)
	mockDB := shared.NewMockDatabase()
	
	// Initialize everything manually
	err := mockDB.Initialize()
	require.NoError(t, err)
	defer mockDB.Close()

	userService := shared.NewUserService(mockDB, logger, config, metrics)
	server := shared.NewServer(userService, logger, config, metrics)

	// Can't easily test the server without starting it!
	// This shows how traditional approach makes integration testing harder
	
	// Just verify we can create everything
	assert.NotNil(t, server)
	assert.Equal(t, 1, mockDB.InitializeCalls)
}