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
func provideDatabase(lc fx.Lifecycle, logger *shared.Logger, config *shared.Config) *shared.Database {
	db := shared.NewDatabase(logger, config)

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
				", rate_limiting="+formatBool(config.App.Features["rate_limiting"]))
			logger.Log("APP", "Try: curl http://"+config.Server.Host+":"+config.Server.Port+"/user?id=1")
			logger.Log("APP", "Config: curl http://"+config.Server.Host+":"+config.Server.Port+"/config")

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

	app := fx.New(
		fx.NopLogger,

		// Provide all dependencies
		fx.Provide(
			provideConfig,   // Needs wrapper for config file path
			provideDatabase, // Needs wrapper for lifecycle hooks
		),

		fx.Provide(
			shared.NewLogger,      // Can use directly!
			shared.NewUserService, // Can use directly!
			shared.NewServer,      // Can use directly!
		),

		// Register the server startup - fx.Invoke runs this function
		fx.Invoke(StartServer),
	)

	// Run the application
	app.Run()
}
