package main

import (
	"context"

	"go.uber.org/fx"

	"github.com/frrist/demofx/shared"
)

// Configuration provider - loaded once and injected everywhere
func provideConfig() (*shared.Config, error) {
	return shared.LoadConfig("config.json")
}

// Database with automatic lifecycle and config injection
// NOTE: Just added metrics parameter - fx provides it automatically!
// NEW: Now returns Database interface and selects implementation based on config
// This is the ONLY place we need to change to switch database implementations!
func provideDatabase(lc fx.Lifecycle, logger *shared.Logger, config *shared.Config, metrics *shared.Metrics) shared.Database {
	// FX automatically selects the right database based on config!
	var db shared.Database
	
	switch config.Database.Type {
	case "persistent":
		logger.Log("APP", "Using persistent database")
		db = shared.NewPersistentDatabase(logger, config, metrics)
	default:
		logger.Log("APP", "Using in-memory database")
		db = shared.NewInMemoryDatabase(logger, config, metrics)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.Initialize()
		},
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})

	return db
}

// StartServer registers lifecycle hooks to start/stop the HTTP server
func StartServer(lc fx.Lifecycle, server *shared.Server, logger *shared.Logger, config *shared.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Log("APP", "FX loaded configuration automatically")
			logger.Log("APP", "Environment: "+config.App.Environment)
			logger.Log("APP", "Features: cache="+formatBool(config.App.Features["cache_enabled"])+
				", rate_limiting="+formatBool(config.App.Features["rate_limiting"])+
				", metrics="+formatBool(config.App.Features["metrics_enabled"]))
			logger.Log("APP", "Try: curl http://"+config.Server.Host+":"+config.Server.Port+"/user?id=1")
			logger.Log("APP", "Config: curl http://"+config.Server.Host+":"+config.Server.Port+"/config")
			logger.Log("APP", "Metrics: curl http://"+config.Server.Host+":"+config.Server.Port+"/metrics")

			go func() {
				if err := server.Start(); err != nil {
					logger.Log("SERVER", "Server error: "+err.Error())
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Stop(ctx)
		},
	})
}

func formatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func main() {
	// FX approach: Configuration and dependencies are injected automatically
	// Notice how fx handles:
	// 1. Config is loaded ONCE and injected everywhere needed
	// 2. Logger is created ONCE with config and injected everywhere
	// 3. No manual passing of dependencies through constructors
	// 4. Clean separation between wiring and business logic
	// 5. ADDING METRICS: Just one line! No constructor changes needed!
	// 6. DATABASE SWITCHING: Just update the provider function - zero impact on other code!

	app := fx.New(
		fx.NopLogger,

		// Provide all dependencies
		fx.Provide(
			provideConfig,   // Needs wrapper for config file path, since nothing provides the path param
			provideDatabase, // Needs wrapper for lifecycle hooks
		),

		fx.Provide(
			shared.NewLogger,
			shared.NewMetrics,     // Just add this one line!
			shared.NewUserService, // No changes needed - fx injects metrics automatically
			shared.NewServer,      // No changes needed - fx injects metrics automatically
		),

		// Register the server startup - fx.Invoke runs this function
		fx.Invoke(StartServer),
	)

	// Run the application
	app.Run()
}
