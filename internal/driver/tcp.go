package driver

import (
	"fmt"
	"net"
	"time"
)

type TCPDriver struct {
	endpoint string
}

func NewTCPDriver(endpoint string) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	return &TCPDriver{endpoint: endpoint}, nil
}

func (t *TCPDriver) Check() (*CheckResult, error) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", t.endpoint, 10*time.Second)
	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("TCP check failed: %v", err)
		return result, nil
	}

	defer conn.Close()

	result.Success = true
	result.Message = fmt.Sprintf("TCP check successful (response time: %v)", duration)

	return result, nil
}

func (t *TCPDriver) GetEndpoint() string {
	return t.endpoint
}

func (t *TCPDriver) GetType() string {
	return "tcp"
}
