package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HttpJsonCheck defines expected JSON field values from a HTTP response
type HttpJsonCheck struct {
	ExpectedStatusCode int               `json:"expectedStatusCode,omitempty"` // optional override for status code check
	JsonAssertions     map[string]string `json:"jsonAssertions"`               // key: JSONPath-like dot string, value: expected value
}

type SmtpCheck struct {
	Helo            string    `json:"helo,omitempty"`
	Tls             bool      `json:"tls,omitempty"`
	StartTls        bool      `json:"startTls,omitempty"`
	EmailSecretRef  SecretRef `json:"emailSecretRef,omitempty"`
	VerifyAssertion string    `json:"verifyAssertion,omitempty"`
	FromAssertion   string    `json:"fromAssertion,omitempty"`
	ToAssertion     string    `json:"toAssertion,omitempty"`
}

// EndpointMonitorSpec defines the desired state of EndpointMonitor
type EndpointMonitorSpec struct {
	Driver        string         `json:"driver"`        // ex: "opensearch", "trino", "http", "http-json"
	Endpoint      string         `json:"endpoint"`      // target service URL
	CheckInterval int            `json:"checkInterval"` // in seconds
	Notify        NotifyConfig   `json:"notify"`
	HttpJsonCheck *HttpJsonCheck `json:"httpJsonCheck,omitempty"` // only relevant for driver = "http-json"
	SmtpCheck     *SmtpCheck     `json:"smtpCheck,omitempty"`     // only relevant for driver = "smtp"
}

// NotifyConfig holds notifier configurations
type NotifyConfig struct {
	Slack   *SlackConfig   `json:"slack,omitempty"`
	Email   *EmailConfig   `json:"email,omitempty"`
	Discord *DiscordConfig `json:"discord,omitempty"`
	Webhook *WebhookConfig `json:"webhook,omitempty"`
}

// WebhookConfig defines Webhook notifier config
type WebhookConfig struct {
	Enabled       bool      `json:"enabled"`
	WebhookURL    string    `json:"webhookUrl"`
	Authorization SecretRef `json:"authorization,omitempty"` // Bearer or other token needed
	ContentType   string    `json:"contentType,omitempty"`   // "application/json", "application/x-www-form-urlencoded", ...
	Contents      string    `json:"contents,omitempty"`      // template, if empty will use GET instead of POST
	AlertOn       []string  `json:"alertOn,omitempty"`       // values: "success", "failure", "change"
}

// SlackConfig defines Slack notifier config
type SlackConfig struct {
	Enabled    bool     `json:"enabled"`
	WebhookURL string   `json:"webhookUrl"`
	AlertOn    []string `json:"alertOn,omitempty"` // values: "success", "failure", "change"
}

// EmailConfig is placeholder (no-op for now)
type EmailConfig struct {
	Enabled        bool      `json:"enabled"`
	From           string    `json:"from,omitempty"`
	To             []string  `json:"to"`
	EmailProvider  string    `json:"emailProvider,omitempty"` // e.g., "ses", "smtp"
	EmailSecretRef SecretRef `json:"emailSecretRef,omitempty"`
	Host           string    `json:"host,omitempty"`    // Host:Port
	Subject        string    `json:"subject,omitempty"` // subject template
	Body           string    `json:"body,omitempty"`    // message body template
	AlertOn        []string  `json:"alertOn,omitempty"` // values: "success", "failure"
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
	LastCheckedTime  metav1.Time `json:"lastCheckedTime,omitempty"`
	LastStatus       string      `json:"lastStatus,omitempty"` // e.g., success/failure
	LastStatusChange metav1.Time `json:"lastStatusChange,omitempty"`
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
