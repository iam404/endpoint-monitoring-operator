# Contributing to Endpoint Monitoring Operator

First off, thank you for considering contributing to the endpoint-monitoring-operator! 
This project is designed to be extensible, reliable, and easy to build upon.

## Overview

This is a Kubernetes operator that lets you declaratively monitor external services like: HTTP / HTTP-JSON APIs, TCP ports, DNS resolutions, Trino clusters, OpenSearch, Ping checks and more.

It follows the Factory Design Pattern to easily support new drivers (monitoring mechanisms) and notifiers (e.g., Slack, email, etc.).


## Project Structure

```csharp
api/                      # CRD schema definitions
config/                   # Install manifests, RBAC, kustomize
internal/
  ├── driver/             # All drivers go here (http, dns, ping, etc.)
  ├── notifier/           # Slack, email, etc.
pkg/
  └── factory/            # Driver/Notifier creation logic
controllers/              # Core reconciliation logic
dist/                     # install.yaml and release manifests
```


## Add a New Driver

To monitor a new type of system (e.g., Kafka, Redis, Mongo):

Create a new Go file in `internal/driver/<yourdriver>.go`

Implement the Driver interface:

```go
type Driver interface {
    Check() (*CheckResult, error)
    GetEndpoint() string
    GetType() string
}
```

Add your driver to the DriverFactory in `pkg/factory/factory.go`

## Add a New Notifier

Want to send alerts to Discord, PagerDuty, SES?

Add a new notifier in `internal/notifier/<notifier>.go`

Implement the Notifier interface:

```go
type Notifier interface {
    SendAlert(status string, msg string) error
}
```

Add the config struct in `api/v1alpha1/endpointmonitor_types.go`

Add support in NewNotifier() in `pkg/factory/factory.go`


## Code Style

1. Use idiomatic Go formatting (go fmt)

2. Follow Kubernetes naming conventions for types and CRDs

3. Keep logic minimal in the Reconciler – offload work to factory/driver layers

4. No hardcoded secrets/tokens; use Kubernetes SecretRef


## Security & Secrets

Run Trivy to scan for secrets in the image:

```bash
trivy image --scanners secret <your-image>
```

## Pull Requests
Always create a feature branch (e.g., feat/kafka-driver)

Write clear PR titles and descriptions

Link to relevant issues if applicable

Follow conventional commit standards - https://www.conventionalcommits.org/en/v1.0.0/

