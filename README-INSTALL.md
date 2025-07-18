# Installation

You can install the endpoint-monitoring-operator into your Kubernetes cluster with a single command :

```bash
kubectl apply -f https://raw.githubusercontent.com/LiciousTech/endpoint-monitoring-operator/main/dist/install.yaml
```

This manifest sets up everything needed.

# Example usage

You can also refer examples folder in this repo.

## 1. DNS

Let’s say you want to monitor that your domain name is resolving correctly, you can use the dns driver.

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: domain-monitor
  namespace: endpoint-monitoring-operator-system
spec:
  driver: dns
  endpoint: your-domain.com
  checkInterval: 30 # in seconds
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - success  # by default alerts only failures.
        - failure 
```

Apply above manifest.


Check status:

```bash
kubectl get endpointmonitor domain-monitor -n endpoint-monitoring-operator-system -o yaml
```

describe :

```bash
kubectl describe endpointmonitor domain-monitor -n endpoint-monitoring-operator-system
```

## 2. HTTP-JSON

Let's say you have a User Service with an endpoint like:

```bash
GET https://api.mycompany.com/v1/status
```

And it responds with:

```json
{
  "status": "UP",
  "service": "user-service",
  "version": "v2.4.1",
  "env": "prod",
  "dependencies": {
    "database": "ok",
    "redis": "ok"
  }
}
```

EndpointMonitor CR for This

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: user-service-healthcheck
  namespace: endpoint-monitoring-operator-system
spec:
  driver: http-json
  endpoint: https://api.mycompany.com/v1/status
  checkInterval: 30 # seconds
  httpJsonCheck:
    expectedStatusCode: 200 # optional
    jsonAssertions:
      status: "UP"
      service: "user-service"
      env: "prod"
      dependencies.database: "ok"
      dependencies.redis: "ok"
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - failure
```

## 3. TRINO

Want to ensure your Trino cluster is up and not in a starting state?

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: trino-coordinator-monitor
  namespace: endpoint-monitoring-operator-system
spec:
  checkInterval: 300  # 5 minutes
  driver: trino
  endpoint: http://trino.company.com
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>  # alertOn not specified , so by default only reports failures.
```

## 4. TCP

Let’s say you want to monitor if a MySQL port on RDS is accepting connections:

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: rds-mysql-tcp-check
  namespace: endpoint-monitoring-operator-system
spec:
  driver: tcp
  endpoint: mysql.company.com:3306
  checkInterval: 30
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - failure
```

## 5. OpenSearch
If you want to ensure your OpenSearch cluster is healthy (i.e., in green or yellow state), you can use the opensearch driver.

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: opensearch-cluster-health
  namespace: endpoint-monitoring-operator-system
spec:
  checkInterval: 30
  driver: opensearch
  endpoint: http://cluster01.os01.svc.cluster.local:9200
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - failure
```

## 6. HTTP

Use the http driver when you only care about the status code of the response (e.g., 200 OK).

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: trino-basic-http-check
  namespace: endpoint-monitoring-operator-system
spec:
  checkInterval: 60  # check every 1 minute
  driver: http
  endpoint: https://status.my-domain.com/
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - failure
```

## 7. Ping

Use the ping driver when you want to verify basic network reachability (ICMP) to a host.

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: ping-google
  namespace: endpoint-monitoring-operator-system
spec:
  checkInterval: 60  # every 1 minute
  driver: ping
  endpoint: 8.8.8.8
  notify:
    slack:
      enabled: true
      webhookUrl: <slack-webhook-url>
      alertOn:
        - failure
```