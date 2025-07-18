# Endpoint Monitoring Operator
Custom kubernetes monitoring operator to probe any endpoint — REST APIs, databases, DNS records, TCP ports, distributed systems like Trino or OpenSearch — and send alerts on Slack, email, or custom notifiers. Built with extensibility in mind using the Factory Design Pattern.

## Description
Endpoint-monitoring-operator provides a declarative way to monitor your critical application endpoints and infrastructure health — directly from inside your Kubernetes cluster.

Unlike basic uptime checkers, this operator can reach real business endpoints (e.g., /v1/status, /v1/get-users) and assert correctness of responses, not just HTTP 200s.

It’s designed to monitor deeply, not just surface availability:

Supports multiple protocols out of the box :

HTTP: Check if a web service is reachable

HTTP+JSON: Validate nested response fields (e.g., status: "UP" and version: "v2.4.1")

TCP: Check open ports on databases or services

DNS: Verify domain resolution (e.g., google.com)

Ping: Test ICMP reachability (e.g., internal IPs)

Trino: Monitor distributed Trino clusters for query service availability

OpenSearch: Monitor cluster health (e.g., red/yellow/green state)

## Built for Extensibility
The operator is designed using the Factory Pattern to ensure:

New monitoring drivers can be added easily (e.g., Kafka, Redis, gRPC)

New notifiers can be plugged in (e.g., PagerDuty, OpsGenie, SES email)

Minimal changes required to core reconciliation logic

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

## NOTE : This README is for someone who wants to develop this operator. If your goal is just to use the operator - just check `README-INSTALL.md`

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/endpoint-monitoring-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/endpoint-monitoring-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/endpoint-monitoring-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Check the other README file - `README-INSTALL.md`

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
Contributions are welcome. Please raise a PR.

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

