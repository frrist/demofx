# Uber FX Demo: Traditional vs Dependency Injection

This demo showcases the advantages of using Uber's fx dependency injection framework compared to traditional manual dependency wiring in Go. Both examples use the same shared components from the `shared/` package, demonstrating how the wiring approach differs.

## Project Structure

```
.
├── shared/              # Common components used by both approaches
│   ├── config.go        # Configuration struct
│   ├── logger.go        # Logging service
│   ├── database.go      # Database service (mock)
│   ├── user_service.go  # User business logic
│   └── server.go        # HTTP server
├── traditional/         # Manual dependency wiring
│   └── main.go     
├── fx-version/          # Automatic dependency injection with fx
│   └── main.go
├── config.json          # Configuration file
└── README.md
```

## Running the Examples

```bash
# Install dependencies
go mod download

# Run traditional version
go run traditional/main.go

# Run fx version  
go run fx-version/main.go

# Test the endpoints (for either version)
curl http://localhost:9090/user?id=1
curl http://localhost:9090/user?id=2
curl http://localhost:9090/health
curl http://localhost:9090/config
```

## Key Differences: Traditional vs FX

Both implementations use the **exact same components** from the `shared/` package. The only difference is how they're wired together.

### What the Demo Shows

The demo uses a configuration file (`config.json`) that multiple services depend on:
- **Server config**: host and port settings
- **Database config**: connection pool, timeout, cache settings  
- **App config**: environment, log level, feature flags

Each component needs both the config and a logger to function properly.

### 1. Configuration and Dependency Injection
**Traditional** (`traditional/main.go`): Must load config and pass it to EVERY component
```go
// Load config first
config, err := shared.LoadConfig("config.json")
logger := shared.NewLogger(config)  // Pass config

// Every component needs config AND logger passed manually
db := shared.NewDatabase(logger, config)
userService := shared.NewUserService(db, logger, config)
server := shared.NewServer(userService, logger, config)
```

**FX** (`fx-version/main.go`): Config loaded once, injected automatically
```go
fx.Provide(
    provideConfig,      // Loads config once
    provideLogger,      // Gets config automatically
    provideDatabase,    // Gets config and logger automatically
    provideUserService, // Gets all dependencies automatically
)
```

### 2. Lifecycle Management
**Traditional**: Manual initialization and cleanup
```go
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
```

**FX**: Automatic lifecycle hooks
```go
lc.Append(fx.Hook{
    OnStart: func(ctx context.Context) error { /* init */ },
    OnStop: func(ctx context.Context) error { /* cleanup */ },
})
```

### 3. Adding New Dependencies
**Traditional**: Must update every constructor call
```go
// If Database needs a new dependency (e.g., metrics), you must:
// 1. Update NewDatabase signature
// 2. Update EVERY place that calls NewDatabase
// 3. Pass the new dependency through
```

**FX**: Just declare what you need
```go
// Just add the parameter to your constructor
func provideDatabase(lc fx.Lifecycle, logger *shared.Logger, 
    config *shared.Config, metrics *Metrics) *shared.Database {
    // FX automatically provides metrics
}
```

### 4. Dependency Graph Validation
**Traditional**: Runtime errors if dependencies missing
**FX**: Compile-time validation of entire dependency graph

### 5. Built-in Debugging
**Traditional**: Manual logging for debugging
**FX**: Built-in dependency graph visualization and logging
```go
fx.WithLogger(func() fxevent.Logger {
    return &fxevent.ConsoleLogger{W: os.Stdout}
})
```

### 6. Testing Benefits
**Traditional**: Hard to mock dependencies
**FX**: Easy to replace dependencies for testing
```go
fx.New(
    fx.Replace(NewMockDatabase),
    UserModule(),
)
```

### 7. Configuration Impact
The demo shows how configuration affects behavior:
- **Logger**: Environment tag ([STAGING]) in output
- **Database**: Cache enabled/disabled, connection pool settings
- **UserService**: Rate limiting on/off based on feature flag
- **Server**: Binds to configured host:port

Try changing `config.json` and see how both versions adapt!

## When to Use FX

FX is particularly beneficial for:
- Large applications with complex dependency graphs
- Microservices requiring clean separation of concerns
- Applications needing robust lifecycle management
- Teams wanting better testability and modularity

The overhead of fx is minimal, making it suitable even for smaller applications that may grow over time.