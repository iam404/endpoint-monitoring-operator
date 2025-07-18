package factory

import (
	"fmt"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/driver"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier/email"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier/slack"
)

// NotifierFactory creates notifiers based on configuration
type NotifierFactory struct{}

// NewNotifier creates a notifier instance based on the configuration
func NewNotifier(config *v1alpha1.NotifyConfig) (notifier.Notifier, error) {
	factory := &NotifierFactory{}
	return factory.CreateNotifier(config)
}

// CreateNotifier implements the factory pattern for notifiers
func (f *NotifierFactory) CreateNotifier(config *v1alpha1.NotifyConfig) (notifier.Notifier, error) {
	if config == nil {
		return nil, fmt.Errorf("notify config is nil")
	}

	var notifiers []notifier.Notifier

	if config.Slack != nil && config.Slack.Enabled {
		slackNotifier, err := slack.New(config.Slack)
		if err != nil {
			return nil, fmt.Errorf("failed to create Slack notifier: %w", err)
		}
		notifiers = append(notifiers, slackNotifier)
	}

	if config.Email != nil && config.Email.Enabled {
		emailNotifier, err := email.New(config.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to create Email notifier: %w", err)
		}
		notifiers = append(notifiers, emailNotifier)
	}

	if len(notifiers) == 0 {
		return nil, fmt.Errorf("no notifiers enabled")
	}

	return &CompositeNotifier{notifiers: notifiers}, nil
}

// CompositeNotifier handles multiple notification channels
type CompositeNotifier struct {
	notifiers []notifier.Notifier
}

// SendAlert sends alerts to all configured notifiers
func (c *CompositeNotifier) SendAlert(status string, msg string) error {
	var errs []error
	for _, n := range c.notifiers {
		if err := n.SendAlert(status, msg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send alerts: %v", errs)
	}
	return nil
}

// DriverFactory creates monitoring drivers based on configuration
type DriverFactory struct{}

// NewDriver creates a driver instance based on the driver type
func NewDriver(driverType string, endpoint string, monitor *v1alpha1.EndpointMonitor) (driver.Driver, error) {
	factory := &DriverFactory{}
	return factory.CreateDriver(driverType, endpoint, monitor)
}

// CreateDriver implements the factory pattern for drivers
func (f *DriverFactory) CreateDriver(driverType string, endpoint string, monitor *v1alpha1.EndpointMonitor) (driver.Driver, error) {
	switch driverType {
	case "http":
		return driver.NewHTTPDriver(endpoint)
	case "http-json":
		return driver.NewHTTPJSONDriver(endpoint, monitor.Spec.HttpJsonCheck)
	case "tcp":
		return driver.NewTCPDriver(endpoint)
	case "dns":
		return driver.NewDNSDriver(endpoint)
	case "ping":
		return driver.NewPingDriver(endpoint)
	case "trino":
		return driver.NewTrinoDriver(endpoint)
	case "opensearch":
		return driver.NewOpenSearchDriver(endpoint)
	default:
		return nil, fmt.Errorf("unsupported driver type: %s", driverType)
	}
}
