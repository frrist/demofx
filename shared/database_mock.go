package shared

import (
	"fmt"
)

// MockDatabase is a test double for the Database interface
type MockDatabase struct {
	// Control behavior
	Users           map[string]string
	ShouldError     bool
	ErrorMessage    string
	InitializeCalls int
	CloseCalls      int
	GetUserCalls    int
	
	// For assertions
	LastRequestedID string
}

// NewMockDatabase creates a new mock database for testing
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Users: map[string]string{
			"test1": "Test User 1",
			"test2": "Test User 2",
			"mock":  "Mock User",
		},
	}
}

// Initialize mock implementation
func (m *MockDatabase) Initialize() error {
	m.InitializeCalls++
	if m.ShouldError {
		return fmt.Errorf("mock initialize error: %s", m.ErrorMessage)
	}
	return nil
}

// Close mock implementation
func (m *MockDatabase) Close() error {
	m.CloseCalls++
	if m.ShouldError {
		return fmt.Errorf("mock close error: %s", m.ErrorMessage)
	}
	return nil
}

// GetUser mock implementation
func (m *MockDatabase) GetUser(id string) (string, error) {
	m.GetUserCalls++
	m.LastRequestedID = id
	
	if m.ShouldError {
		return "", fmt.Errorf("mock error: %s", m.ErrorMessage)
	}
	
	if user, ok := m.Users[id]; ok {
		return user, nil
	}
	
	return "", fmt.Errorf("user not found")
}