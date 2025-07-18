package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
)

type HTTPJSONDriver struct {
	endpoint   string
	client     *http.Client
	assertions map[string]string
}

func NewHTTPJSONDriver(endpoint string, check *v1.HttpJsonCheck) (Driver, error) {
	if endpoint == "" || check == nil {
		return nil, fmt.Errorf("invalid endpoint or config for http-json")
	}

	return &HTTPJSONDriver{
		endpoint:   endpoint,
		assertions: check.JsonAssertions,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (h *HTTPJSONDriver) Check() (*CheckResult, error) {
	start := time.Now()

	resp, err := h.client.Get(h.endpoint)
	duration := time.Since(start)

	result := &CheckResult{ResponseTime: duration}
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("HTTP-JSON request failed: %v", err)
		return result, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = "failed to read HTTP response body"
		return result, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		result.Success = false
		result.Error = err
		result.Message = "invalid JSON response"
		return result, nil
	}

	// Simple assertion logic
	for path, expected := range h.assertions {
		actual, ok := resolveJSONPath(payload, path)
		if !ok || fmt.Sprintf("%v", actual) != expected {
			result.Success = false
			result.Message = fmt.Sprintf("assertion failed at '%s': expected '%s', got '%v'", path, expected, actual)
			return result, nil
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("HTTP-JSON check successful (response time: %v)", duration)
	return result, nil
}

func (h *HTTPJSONDriver) GetEndpoint() string {
	return h.endpoint
}

func (h *HTTPJSONDriver) GetType() string {
	return "http-json"
}

// resolveJSONPath traverses a map[string]interface{} based on dot-separated path (e.g., "cluster.status")
func resolveJSONPath(data map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = data

	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
