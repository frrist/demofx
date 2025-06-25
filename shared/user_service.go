package shared

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// UserService handles user-related operations
type UserService struct {
	db           *Database
	logger       *Logger
	rateLimiting bool
	lastRequest  time.Time
}

// NewUserService creates a new user service
func NewUserService(db *Database, logger *Logger, config *Config) *UserService {
	return &UserService{
		db:           db,
		logger:       logger,
		rateLimiting: config.App.Features["rate_limiting"],
	}
}

// GetUserHandler handles HTTP requests for user data
func (s *UserService) GetUserHandler(c echo.Context) error {
	// Simple rate limiting if enabled
	if s.rateLimiting {
		now := time.Now()
		if s.lastRequest.Add(100 * time.Millisecond).After(now) {
			s.logger.Log("USER", "Rate limit exceeded")
			return c.String(http.StatusTooManyRequests, "Too many requests")
		}
		s.lastRequest = now
	}
	
	userID := c.QueryParam("id")
	if userID == "" {
		s.logger.Log("USER", "Missing user ID in request")
		return c.String(http.StatusBadRequest, "Missing user ID")
	}

	user, err := s.db.GetUser(userID)
	if err != nil {
		s.logger.Log("USER", fmt.Sprintf("Error fetching user: %v", err))
		return c.String(http.StatusNotFound, "User not found")
	}

	s.logger.Log("USER", fmt.Sprintf("Successfully fetched user: %s", user))
	return c.String(http.StatusOK, fmt.Sprintf("User: %s\n", user))
}
