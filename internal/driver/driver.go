package driver

import "time"

// CheckResult represents the result of a health check.
type CheckResult struct {
	Success      bool
	ResponseTime time.Duration
	StatusCode   int
	Error        error
	ErrorMessage string
	Message      string
}

// Driver interface for different monitoring types.
type Driver interface {
	Check() (*CheckResult, error)
	GetEndpoint() string
	GetType() string
}
