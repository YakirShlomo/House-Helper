# Grafana Dashboards for House Helper

This directory contains Grafana dashboard definitions for monitoring the House Helper application.

## Dashboards

### 1. API Performance Dashboard (`api-performance.json`)

Monitors API service performance metrics:
- Request rate (requests/second)
- Response times (P50, P95, P99)
- Error rate
- HTTP status codes distribution
- Endpoint-specific metrics

### 2. Infrastructure Dashboard (`infrastructure.json`)

Monitors infrastructure health:
- CPU usage per service
- Memory usage per service
- Network I/O
- Disk usage
- Pod restart count

### 3. Database Dashboard (`database.json`)

Monitors PostgreSQL database:
- Query performance
- Connection pool status
- Cache hit ratio
- Slow queries
- Transaction rate
- Database size

### 4. Business Metrics Dashboard (`business-metrics.json`)

Monitors business KPIs:
- Active users
- Tasks created/completed
- User engagement
- Feature adoption
- Points awarded

### 5. Alerts Dashboard (`alerts.json`)

Shows active alerts and their status:
- Critical alerts
- Warning alerts
- Alert history
- Alert response time

## Usage

### Import Dashboards

```bash
# Using Grafana API
for file in *.json; do
  curl -X POST \
    -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
    -H "Content-Type: application/json" \
    -d @"$file" \
    http://grafana.example.com/api/dashboards/db
done
```

### Using Terraform

```hcl
resource "grafana_dashboard" "api_performance" {
  config_json = file("${path.module}/dashboards/api-performance.json")
}

resource "grafana_dashboard" "infrastructure" {
  config_json = file("${path.module}/dashboards/infrastructure.json")
}
```

### Using Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  api-performance.json: |
    {{ .Files.Get "dashboards/api-performance.json" | indent 4 }}
```

## Dashboard Configuration

All dashboards are configured with:
- Auto-refresh: 30 seconds
- Time range: Last 1 hour (default)
- Templating for multi-environment support
- Variables: `$environment`, `$service`, `$namespace`

## Data Sources

Dashboards expect the following data sources:
- **Prometheus**: `prometheus` (default)
- **Loki**: `loki` (for logs)
- **Tempo**: `tempo` (for traces)

## Variables

All dashboards support these variables:

```yaml
variables:
  - name: environment
    type: custom
    options: [development, staging, production]
    current: production
    
  - name: namespace
    type: query
    query: label_values(kube_namespace_labels, namespace)
    current: house-helper-prod
    
  - name: service
    type: query
    query: label_values(container_cpu_usage_seconds_total{namespace="$namespace"}, pod)
    current: all
```

## Panels

### Common Panel Types

**Graph Panel:**
```json
{
  "type": "graph",
  "title": "Request Rate",
  "targets": [
    {
      "expr": "rate(http_requests_total{namespace=\"$namespace\"}[5m])",
      "legendFormat": "{{service}}"
    }
  ]
}
```

**Stat Panel:**
```json
{
  "type": "stat",
  "title": "Error Rate",
  "targets": [
    {
      "expr": "rate(http_requests_total{status=~\"5..\",namespace=\"$namespace\"}[5m])",
      "format": "time_series"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "percentunit",
      "thresholds": {
        "steps": [
          { "value": 0, "color": "green" },
          { "value": 0.01, "color": "yellow" },
          { "value": 0.05, "color": "red" }
        ]
      }
    }
  }
}
```

**Table Panel:**
```json
{
  "type": "table",
  "title": "Top Endpoints",
  "targets": [
    {
      "expr": "topk(10, rate(http_requests_total{namespace=\"$namespace\"}[5m]))",
      "format": "table"
    }
  ]
}
```

## Alerting

Dashboards include alert annotations:

```json
{
  "annotations": {
    "list": [
      {
        "datasource": "prometheus",
        "enable": true,
        "expr": "ALERTS{namespace=\"$namespace\"}",
        "iconColor": "red",
        "name": "Alerts",
        "step": "60s",
        "tagKeys": "alertname,severity",
        "textFormat": "{{alertname}}",
        "titleFormat": "Alert"
      }
    ]
  }
}
```

## Best Practices

1. **Use Templates**: Define dashboard templates for consistent look and feel
2. **Add Descriptions**: Include panel descriptions for context
3. **Set Thresholds**: Configure color thresholds for quick visual feedback
4. **Use Variables**: Make dashboards reusable across environments
5. **Group Panels**: Organize related panels into rows
6. **Add Links**: Link to related dashboards and documentation

## Troubleshooting

### Dashboard Not Loading

```bash
# Check Grafana logs
kubectl logs -n monitoring deployment/grafana

# Verify data source
curl -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  http://grafana.example.com/api/datasources
```

### No Data in Panels

```bash
# Verify Prometheus is scraping metrics
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Check targets
open http://localhost:9090/targets

# Test query
curl http://localhost:9090/api/v1/query?query=up
```

### Slow Dashboard Loading

- Reduce time range
- Increase refresh interval
- Optimize queries (use recording rules)
- Add query limits

## Maintenance

### Update Dashboards

```bash
# Export existing dashboard
curl -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  http://grafana.example.com/api/dashboards/uid/${DASHBOARD_UID} \
  | jq '.dashboard' > dashboard.json

# Modify and re-import
curl -X POST \
  -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  -H "Content-Type: application/json" \
  -d @dashboard.json \
  http://grafana.example.com/api/dashboards/db
```

### Version Control

- Keep dashboards in Git
- Use meaningful commit messages
- Review changes in PRs
- Tag dashboard versions

## Examples

### Simple Query

```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / 
rate(http_requests_total[5m])

# P95 latency
histogram_quantile(0.95, 
  rate(http_request_duration_seconds_bucket[5m])
)
```

### Complex Query

```promql
# Requests per endpoint, sorted by error rate
topk(10,
  sum by (endpoint) (
    rate(http_requests_total{status=~"5.."}[5m])
  ) / 
  sum by (endpoint) (
    rate(http_requests_total[5m])
  )
)
```

## Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Query Functions](https://prometheus.io/docs/prometheus/latest/querying/functions/)
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)

## License

Copyright © 2024 House Helper. All rights reserved.
