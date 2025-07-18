package driver

import (
	"fmt"
	"net/http"
	"time"
)

type HTTPDriver struct {
	endpoint string
	client   *http.Client
}

func NewHTTPDriver(endpoint string) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	return &HTTPDriver{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (h *HTTPDriver) Check() (*CheckResult, error) {
	start := time.Now()

	resp, err := h.client.Get(h.endpoint)
	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("HTTP check failed: %v", err)
		return result, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = fmt.Sprintf("HTTP check successful (status: %d, response time: %v)", resp.StatusCode, duration)
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("HTTP check failed (status: %d, response time: %v)", resp.StatusCode, duration)
	}

	return result, nil
}

func (h *HTTPDriver) GetEndpoint() string {
	return h.endpoint
}

func (h *HTTPDriver) GetType() string {
	return "http"
}
