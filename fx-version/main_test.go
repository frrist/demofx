package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/frrist/demofx/shared"
)

// TestUserServiceFX demonstrates fx testing approach
// Notice how fx handles:
// 1. Automatic dependency injection
// 2. Clean dependency replacement
// 3. Lifecycle management
func TestUserServiceFX(t *testing.T) {
	var userService *shared.UserService
	var mockDB *shared.MockDatabase

	// FX SETUP: Just declare what we want!
	app := fxtest.New(
		t,
		// Provide test config
		fx.Provide(func() (*shared.Config, error) {
			return &shared.Config{
				Server:   shared.ServerConfig{Host: "localhost", Port: "8080"},
				Database: shared.DatabaseConfig{Type: "mock"},
				App: shared.AppConfig{
					Environment: "test",
					Features: map[string]bool{
						"cache_enabled":   false,
						"rate_limiting":   false,
						"metrics_enabled": false,
					},
				},
			}, nil
		}),

		// Use all the normal providers
		fx.Provide(
			shared.NewLogger,
			shared.NewMetrics,
			shared.NewUserService,
		),

		// MAGIC: Replace database with mock!
		fx.Provide(func() shared.Database {
			mockDB = shared.NewMockDatabase()
			return mockDB
		}),

		// Extract what we need for testing
		fx.Populate(&userService),
	)

	// FX handles lifecycle automatically!
	app.RequireStart()
	defer app.RequireStop()

	// Create echo instance for testing
	e := echo.New()

	t.Run("successful user lookup", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user?id=test1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := userService.GetUserHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Test User 1")
		assert.Equal(t, "test1", mockDB.LastRequestedID)
	})

	t.Run("user not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user?id=nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := userService.GetUserHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "User not found")
	})

	// FX handles cleanup automatically!
}

// TestDatabaseSwappingFX shows how easy it is to test different implementations
func TestDatabaseSwappingFX(t *testing.T) {
	testCases := []struct {
		name         string
		provideDB    interface{}
		expectedUser string
	}{
		{
			name: "mock database",
			provideDB: func() shared.Database {
				mock := shared.NewMockDatabase()
				mock.Users["1"] = "Mock User"
				return mock
			},
			expectedUser: "Mock User",
		},
		{
			name: "custom mock database",
			provideDB: func() shared.Database {
				mock := shared.NewMockDatabase()
				mock.Users["1"] = "Custom Test User"
				return mock
			},
			expectedUser: "Custom Test User",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var userService *shared.UserService

			// Each test gets its own fx app with different database!
			app := fxtest.New(
				t,
				fx.Provide(
					func() (*shared.Config, error) {
						return &shared.Config{
							Database: shared.DatabaseConfig{Type: "test"},
							App:      shared.AppConfig{Environment: "test"},
						}, nil
					},
					shared.NewLogger,
					shared.NewMetrics,
					shared.NewUserService,
					tc.provideDB, // Just swap the database provider!
				),
				fx.Populate(&userService),
			)

			app.RequireStart()
			defer app.RequireStop()

			// Test with the swapped implementation
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/user?id=1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := userService.GetUserHandler(c)
			assert.NoError(t, err)
			
			if tc.expectedUser != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedUser)
			}
		})
	}
}

// TestIntegrationFX shows how fx makes integration testing elegant
func TestIntegrationFX(t *testing.T) {
	var server *shared.Server
	var mockDB *shared.MockDatabase
	var metrics *shared.Metrics

	// FX MAGIC: Complete integration test with minimal setup!
	app := fxtest.New(
		t,
		// Test configuration
		fx.Provide(func() (*shared.Config, error) {
			return &shared.Config{
				Server:   shared.ServerConfig{Host: "localhost", Port: "0"},
				Database: shared.DatabaseConfig{Type: "mock"},
				App: shared.AppConfig{
					Environment: "test",
					Features:    map[string]bool{"metrics_enabled": true},
				},
			}, nil
		}),

		// All production providers
		fx.Provide(
			shared.NewLogger,
			shared.NewMetrics,
			shared.NewUserService,
			shared.NewServer,
		),

		// Mock database with lifecycle
		fx.Provide(func(lc fx.Lifecycle) shared.Database {
			mockDB = shared.NewMockDatabase()
			
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return mockDB.Initialize()
				},
				OnStop: func(ctx context.Context) error {
					return mockDB.Close()
				},
			})
			
			return mockDB
		}),

		// Extract what we need
		fx.Populate(&server, &metrics),
	)

	// Start the app - fx handles all initialization!
	app.RequireStart()
	defer app.RequireStop()

	// Verify everything is wired correctly
	assert.NotNil(t, server)
	assert.NotNil(t, metrics)
	assert.Equal(t, 1, mockDB.InitializeCalls)

	// We could even start the server and make real HTTP requests!
	// fx makes this trivial compared to traditional approach
}

// TestLifecycleManagementFX shows fx's automatic lifecycle handling
func TestLifecycleManagementFX(t *testing.T) {
	mockDB := shared.NewMockDatabase()
	var db shared.Database
	
	app := fxtest.New(
		t,
		fx.Provide(
			func() (*shared.Config, error) {
				return &shared.Config{
					Database: shared.DatabaseConfig{Type: "mock"},
					App:      shared.AppConfig{Environment: "test"},
				}, nil
			},
			shared.NewLogger,
		),
		
		// Provide database with lifecycle hooks
		fx.Provide(func(lc fx.Lifecycle) shared.Database {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					t.Log("FX: Starting database")
					return mockDB.Initialize()
				},
				OnStop: func(ctx context.Context) error {
					t.Log("FX: Stopping database")
					return mockDB.Close()
				},
			})
			return mockDB
		}),
		
		// Need to populate or invoke to trigger creation
		fx.Populate(&db),
	)

	// Before start
	assert.Equal(t, 0, mockDB.InitializeCalls)
	assert.Equal(t, 0, mockDB.CloseCalls)

	// Start - fx calls Initialize
	app.RequireStart()
	assert.Equal(t, 1, mockDB.InitializeCalls)
	assert.Equal(t, 0, mockDB.CloseCalls)

	// Stop - fx calls Close
	app.RequireStop()
	assert.Equal(t, 1, mockDB.InitializeCalls)
	assert.Equal(t, 1, mockDB.CloseCalls)

	// FX ensures proper cleanup even if we forget!
}

// TestMetricsInjectionFX shows how easy it is to test cross-cutting concerns
func TestMetricsInjectionFX(t *testing.T) {
	var metrics *shared.Metrics
	var userService *shared.UserService
	var mockDB *shared.MockDatabase

	app := fxtest.New(
		t,
		fx.Provide(
			func() (*shared.Config, error) {
				return &shared.Config{
					App: shared.AppConfig{
						Features: map[string]bool{"metrics_enabled": true},
					},
				}, nil
			},
			shared.NewLogger,
			shared.NewMetrics,
			shared.NewUserService,
			func() shared.Database { 
				mockDB = shared.NewMockDatabase()
				return mockDB
			},
		),
		fx.Populate(&metrics, &userService),
	)

	app.RequireStart()
	defer app.RequireStop()

	// Make a request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/user?id=mock", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Before request
	statsBefore := metrics.GetStats()
	assert.Contains(t, statsBefore, "User Lookups: 0")
	assert.Contains(t, statsBefore, "Queries: 0")

	// Make request
	err := userService.GetUserHandler(c)
	require.NoError(t, err)

	// After request - metrics automatically updated!
	statsAfter := metrics.GetStats()
	assert.Contains(t, statsAfter, "User Lookups: 1")
	// Note: Mock database doesn't call metrics.RecordDBQuery() - that's good for isolation!
	// In production, the real database implementations call it
	
	// Verify mock was called
	assert.Equal(t, 1, mockDB.GetUserCalls)

	// This shows how fx automatically injects metrics everywhere needed
	// Without fx, we'd have to manually pass metrics to every component
}
