package driver

import "time"

// CheckResult represents the result of a health check
type CheckResult struct {
	Success      bool
	ResponseTime time.Duration
	Error        error
	Message      string
}

// Driver interface for different monitoring types
type Driver interface {
	Check() (*CheckResult, error)
	GetEndpoint() string
	GetType() string
}
