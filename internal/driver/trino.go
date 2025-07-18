package driver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TrinoDriver struct {
	endpoint string
	client   *http.Client
}

type NodeVersion struct {
	Version string `json:"version"`
}

type TrinoInfoResponse struct {
	Uptime      string      `json:"uptime"`
	Coordinator bool        `json:"coordinator"`
	Starting    bool        `json:"starting"`
	Environment string      `json:"environment"`
	NodeVersion NodeVersion `json:"nodeVersion"`
}

func NewTrinoDriver(endpoint string) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	endpoint = strings.TrimSuffix(endpoint, "/")

	return &TrinoDriver{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (t *TrinoDriver) Check() (*CheckResult, error) {
	start := time.Now()

	infoURL := t.endpoint + "/v1/info"

	resp, err := t.client.Get(infoURL)
	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("Trino check failed: %v", err)
		return result, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result.Success = false
		result.Message = fmt.Sprintf("Trino check failed (status: %d, response time: %v)", resp.StatusCode, duration)
		return result, nil
	}

	var trinoInfo TrinoInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&trinoInfo); err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("Trino check failed to parse response: %v", err)
		return result, nil
	}

	// Trino is healthy if:
	// 1. It's a coordinator node (coordinator: true)
	// 2. It has finished starting (starting: false)
	// 3. API is responding (we already checked HTTP 200)
	if trinoInfo.Coordinator && !trinoInfo.Starting {
		result.Success = true
		result.Message = fmt.Sprintf("Trino check successful (coordinator: %t, starting: %t, uptime: %s, version: %s, env: %s, response time: %v)",
			trinoInfo.Coordinator, trinoInfo.Starting, trinoInfo.Uptime, trinoInfo.NodeVersion.Version, trinoInfo.Environment, duration)
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("Trino check failed (coordinator: %t, starting: %t, uptime: %s, version: %s, env: %s, response time: %v)",
			trinoInfo.Coordinator, trinoInfo.Starting, trinoInfo.Uptime, trinoInfo.NodeVersion.Version, trinoInfo.Environment, duration)
	}

	return result, nil
}

func (t *TrinoDriver) GetEndpoint() string {
	return t.endpoint
}

func (t *TrinoDriver) GetType() string {
	return "trino"
}
