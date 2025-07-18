package driver

import (
	"fmt"
	"net"
	"time"
)

type DNSDriver struct {
	endpoint string
}

func NewDNSDriver(endpoint string) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	return &DNSDriver{endpoint: endpoint}, nil
}

func (d *DNSDriver) Check() (*CheckResult, error) {
	start := time.Now()

	_, err := net.LookupHost(d.endpoint)
	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("DNS check failed: %v", err)
		return result, nil
	}

	result.Success = true
	result.Message = fmt.Sprintf("DNS check successful (response time: %v)", duration)

	return result, nil
}

func (d *DNSDriver) GetEndpoint() string {
	return d.endpoint
}

func (d *DNSDriver) GetType() string {
	return "dns"
}
