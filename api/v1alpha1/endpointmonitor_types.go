package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HttpJsonCheck defines expected JSON field values from a HTTP response
type HttpJsonCheck struct {
	ExpectedStatusCode int               `json:"expectedStatusCode,omitempty"` // optional override for status code check
	JsonAssertions     map[string]string `json:"jsonAssertions"`               // key: JSONPath-like dot string, value: expected value
}

// EndpointMonitorSpec defines the desired state of EndpointMonitor
type EndpointMonitorSpec struct {
	Driver        string         `json:"driver"`        // ex: "opensearch", "trino", "http", "http-json"
	Endpoint      string         `json:"endpoint"`      // target service URL
	CheckInterval int            `json:"checkInterval"` // in seconds
	Notify        NotifyConfig   `json:"notify"`
	HttpJsonCheck *HttpJsonCheck `json:"httpJsonCheck,omitempty"` // only relevant for driver = "http-json"
}

// NotifyConfig holds notifier configurations
type NotifyConfig struct {
	Slack   *SlackConfig   `json:"slack,omitempty"`
	Email   *EmailConfig   `json:"email,omitempty"`
	Discord *DiscordConfig `json:"discord,omitempty"`
}

// SlackConfig defines Slack notifier config
type SlackConfig struct {
	Enabled    bool     `json:"enabled"`
	WebhookURL string   `json:"webhookUrl"`
	AlertOn    []string `json:"alertOn,omitempty"` // values: "success", "failure"
}

// EmailConfig is placeholder (no-op for now)
type EmailConfig struct {
	Enabled        bool      `json:"enabled"`
	From           string    `json:"from"`
	To             []string  `json:"to"`
	EmailProvider  string    `json:"emailProvider"` // e.g., "ses"
	EmailSecretRef SecretRef `json:"emailSecretRef"`
}

type SecretRef struct {
	Name string `json:"name"`
}

// DiscordConfig defines Discord notifier config
type DiscordConfig struct {
	Enabled    bool     `json:"enabled"`
	WebhookURL string   `json:"webhookUrl"`
	AlertOn    []string `json:"alertOn,omitempty"` // values: "success", "failure"
}

// EndpointMonitorStatus defines the observed state of EndpointMonitor
type EndpointMonitorStatus struct {
	LastCheckedTime metav1.Time `json:"lastCheckedTime,omitempty"`
	LastStatus      string      `json:"lastStatus,omitempty"` // e.g., success/failure
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type EndpointMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointMonitorSpec   `json:"spec,omitempty"`
	Status EndpointMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type EndpointMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EndpointMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EndpointMonitor{}, &EndpointMonitorList{})
}
