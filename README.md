# Uber FX Demo: Traditional vs Dependency Injection

This demo showcases the advantages of using Uber's fx dependency injection framework compared to traditional manual dependency wiring in Go. Both examples use the same shared components from the `shared/` package, demonstrating how the wiring approach differs.

## Project Structure

```
.
├── shared/                      # Common components used by both approaches
│   ├── config.go                # Configuration struct with database type
│   ├── logger.go                # Logging service
│   ├── database_interface.go    # Database interface
│   ├── database_inmemory.go     # In-memory database implementation
│   ├── database_persistent.go   # File-based persistent database
│   ├── metrics.go               # Metrics collection service
│   ├── user_service.go          # User business logic
│   └── server.go                # HTTP server with Echo framework
├── traditional/                 # Manual dependency wiring
│   └── main.go     
├── fx-version/                  # Automatic dependency injection with fx
│   └── main.go
├── config.json                  # Configuration file
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
curl http://localhost:9090/metrics
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

### 3. Adding New Dependencies (e.g., Metrics)
**Traditional**: Must update every constructor call
```go
// Before: db := shared.NewDatabase(logger, config)
// After:  db := shared.NewDatabase(logger, config, metrics)
// Must update EVERY constructor in the chain!
```

**FX**: Just add one line
```go
fx.Provide(
    shared.NewMetrics,  // Just add this line!
    // All other providers automatically get metrics injected
)
```

### 4. Swapping Implementations (e.g., Database Types)
**Traditional**: Conditional logic in main function
```go
// Must add switch/if statements in main
var db shared.Database
switch config.Database.Type {
case "persistent":
    db = shared.NewPersistentDatabase(logger, config, metrics)
default:
    db = shared.NewInMemoryDatabase(logger, config, metrics)
}
```

**FX**: Update only the provider function
```go
// Only change the provideDatabase function
// No other code needs to change!
func provideDatabase(...) shared.Database {
    switch config.Database.Type {
    case "persistent":
        return shared.NewPersistentDatabase(...)
    default:
        return shared.NewInMemoryDatabase(...)
    }
}
```

### 5. Dependency Graph Validation
**Traditional**: Runtime errors if dependencies missing
**FX**: Compile-time validation of entire dependency graph

### 6. Built-in Debugging
**Traditional**: Manual logging for debugging
**FX**: Built-in dependency graph visualization and logging
```go
fx.WithLogger(func() fxevent.Logger {
    return &fxevent.ConsoleLogger{W: os.Stdout}
})
```

### 7. Testing Benefits

See the test files for detailed examples:
- `traditional/main_test.go` - Shows manual dependency wiring complexity
- `fx-version/main_test.go` - Shows clean fx testing approach

**Traditional**: Manual dependency wiring, hard to mock
```go
// Must manually create all dependencies in order
config := &shared.Config{...}
logger := shared.NewLogger(config)
metrics := shared.NewMetrics(config)
mockDB := shared.NewMockDatabase()
// Don't forget to initialize!
err := mockDB.Initialize()
// Manual cleanup required
defer mockDB.Close()
```

**FX**: Automatic injection, easy mocking
```go
app := fxtest.New(t,
    fx.Provide(
        provideTestConfig,
        shared.NewLogger,
        shared.NewMetrics,
        func() shared.Database { return mockDB }, // Easy mock injection!
    ),
    fx.Populate(&userService),
)
// Lifecycle handled automatically!
app.RequireStart()
defer app.RequireStop()
```

Key testing advantages with fx:
- Swap implementations with a single line
- Automatic lifecycle management (no forgotten cleanup)
- Test different configurations easily
- Integration tests with minimal setup
- Each test gets isolated dependency graph

### 8. Configuration Impact
The demo shows how configuration affects behavior:
- **Logger**: Environment tag ([STAGING]) in output
- **Database**: 
  - Type selection (inmemory vs persistent)
  - Cache enabled/disabled, connection pool settings
- **UserService**: Rate limiting on/off based on feature flag
- **Server**: Binds to configured host:port
- **Metrics**: Tracks HTTP requests, DB queries, cache hits/misses

Try changing `config.json` (e.g., set `"type": "inmemory"`) and see how both versions adapt!

## When to Use FX

FX is particularly beneficial for:
- Large applications with complex dependency graphs
- Microservices requiring clean separation of concerns
- Applications needing robust lifecycle management
- Teams wanting better testability and modularity

The overhead of fx is minimal, making it suitable even for smaller applications that may grow over time.