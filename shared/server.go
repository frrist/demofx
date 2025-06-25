package shared

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents the HTTP server
type Server struct {
	echo    *echo.Echo
	logger  *Logger
	config  *ServerConfig
	metrics *Metrics
}

// NewServer creates a new HTTP server with the given handlers
// NOTE: In v2, we added metrics parameter - yet another breaking change!
func NewServer(userService *UserService, logger *Logger, config *Config, metrics *Metrics) *Server {
	e := echo.New()
	
	// Disable Echo's default logger
	e.HideBanner = true
	e.HidePort = true
	
	// Add middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	
	// Metrics middleware - track all HTTP requests
	if metrics != nil {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				start := time.Now()
				err := next(c)
				duration := time.Since(start)
				
				// Record metrics
				path := c.Path()
				if path == "" {
					path = c.Request().URL.Path
				}
				metrics.RecordHTTPRequest(path, duration)
				
				return err
			}
		})
	}
	
	// Custom logger middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			logger.Log("HTTP", fmt.Sprintf("%s %s %d", c.Request().Method, c.Request().URL.Path, c.Response().Status))
			return err
		}
	})
	
	// Register routes
	e.GET("/user", userService.GetUserHandler)
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	
	// Add config endpoint to show configuration in use
	e.GET("/config", func(c echo.Context) error {
		return c.JSON(http.StatusOK, config)
	})
	
	// Add metrics endpoint
	e.GET("/metrics", func(c echo.Context) error {
		if metrics == nil {
			return c.String(http.StatusNotFound, "Metrics not enabled")
		}
		return c.String(http.StatusOK, metrics.GetStats())
	})

	return &Server{
		echo:    e,
		logger:  logger,
		config:  &config.Server,
		metrics: metrics,
	}
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	s.logger.Log("SERVER", fmt.Sprintf("Starting server on %s", addr))
	return s.echo.Start(addr)
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Log("SERVER", "Stopping server...")
	return s.echo.Shutdown(ctx)
}