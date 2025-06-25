package shared

// Database defines the interface for user data storage
type Database interface {
	Initialize() error
	Close() error
	GetUser(id string) (string, error)
}