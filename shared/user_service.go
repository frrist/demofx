package shared

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// UserService handles user-related operations
type UserService struct {
	db           Database
	logger       *Logger
	metrics      *Metrics
	rateLimiting bool
	lastRequest  time.Time
}

// NewUserService creates a new user service
// NOTE: In v2, we added metrics parameter - another breaking change!
// NOTE: In v3, we changed db parameter from *Database to Database interface
func NewUserService(db Database, logger *Logger, config *Config, metrics *Metrics) *UserService {
	return &UserService{
		db:           db,
		logger:       logger,
		metrics:      metrics,
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

	// Track user lookup
	if s.metrics != nil {
		s.metrics.RecordUserLookup()
	}
	
	user, err := s.db.GetUser(userID)
	if err != nil {
		s.logger.Log("USER", fmt.Sprintf("Error fetching user: %v", err))
		return c.String(http.StatusNotFound, "User not found")
	}

	s.logger.Log("USER", fmt.Sprintf("Successfully fetched user: %s", user))
	return c.String(http.StatusOK, fmt.Sprintf("User: %s\n", user))
}
