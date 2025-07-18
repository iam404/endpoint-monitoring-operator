# Endpoint-Monitoring Operator

> A lightweight, extensible Kubernetes Operator that probes *any* endpoint—HTTP/JSON, TCP, DNS, ICMP, Trino, OpenSearch, and more—and routes alerts to Slack or e-mail with a simple Custom Resource.  

![Go](https://img.shields.io/badge/Go-%3E%3D1.23-blue?logo=go)
![License](https://img.shields.io/github/license/LiciousTech/endpoint-monitoring-operator)

---

## Why another monitor?

Traditional uptime checkers only tell you if a port is open. **Endpoint-Monitoring Operator** runs *inside* your cluster, so it can:

* Hit real business URLs such as `/v1/status` that are not exposed publicly.  
* Assert deep JSON fields, not just `HTTP 200`.  
* Validate distributed systems (Trino, OpenSearch) and network primitives (DNS, TCP, Ping).  
* Deliver alerts through pluggable notifiers (Slack today, e-mail and PagerDuty soon). :contentReference[oaicite:0]{index=0}

---

## Supported drivers

| Driver        | Typical use-case                                   |
|---------------|----------------------------------------------------|
| `http`        | Basic status-code check (200/302/…​)               |
| `http-json`   | Validate JSON payload & status code                |
| `tcp`         | Verify a service is listening on a port            |
| `dns`         | Ensure a domain resolves to expected IP(s)         |
| `ping`        | Simple ICMP reachability                           |
| `trino`       | Confirm Trino coordinator is *READY*               |
| `opensearch`  | Check cluster health is `green` / `yellow`         |

Adding a new driver or notifier is only a few lines—everything is wired through a Factory pattern. :contentReference[oaicite:1]{index=1}

---

## CRD quick-look

```yaml
apiVersion: monitoring.licious.app/v1alpha1
kind: EndpointMonitor
metadata:
  name: my-check
spec:
  driver: http-json                 # see table above
  endpoint: https://api.example.com/v1/status
  checkInterval: 30                 # seconds
  httpJsonCheck:                    # driver-specific section
    expectedStatusCode: 200
    jsonAssertions:
      status: "UP"
      version: "v1.2.3"
  notify:
    slack:
      enabled: true
      webhookUrl: https://hooks.slack.com/services/XXX/YYY/ZZZ
      alertOn:                       # optional – defaults to ["failure"]
        - success
        - failure
