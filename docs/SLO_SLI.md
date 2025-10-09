# Service Level Objectives (SLOs) and Service Level Indicators (SLIs)

## Overview

This document defines the Service Level Objectives (SLOs) and Service Level Indicators (SLIs) for the House Helper application. SLOs represent our commitment to service quality, while SLIs are the metrics we use to measure our performance against those objectives.

## SLO Framework

We follow Google's SRE principles for defining SLOs:
- **SLI**: Service Level Indicator - A quantitative measure of service level
- **SLO**: Service Level Objective - A target value or range for an SLI
- **SLA**: Service Level Agreement - A business contract with consequences for missing SLOs

### Error Budget

Error budget = 100% - SLO

For a 99.9% availability SLO:
- Error budget: 0.1%
- Maximum downtime per 30 days: 43.2 minutes
- Maximum downtime per 365 days: 8.76 hours

## API Service SLOs

### 1. Availability

**SLO**: 99.9% of API requests should succeed (non-5xx responses)

**SLI**: 
```promql
sum(rate(http_requests_total{status!~"5.."}[30d]))
/
sum(rate(http_requests_total[30d]))
```

**Error Budget**:
- 30-day window: 43.2 minutes of downtime
- 7-day window: 10.08 minutes of downtime

**Measurement Window**: 30 days rolling

**Alert Threshold**: 99.85% (consume 50% of error budget)

### 2. Latency

**SLO**: 95% of API requests should complete within 500ms

**SLI**:
```promql
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket[30d])) by (le)
)
```

**Error Budget**:
- 5% of requests can exceed 500ms
- On 1M requests/day: 50K requests can be slow

**Measurement Window**: 30 days rolling

**Alert Threshold**: P95 > 450ms (90% of budget consumed)

### 3. Freshness

**SLO**: 99% of data should be less than 5 seconds old

**SLI**:
```promql
sum(rate(data_freshness_seconds_bucket{le="5"}[30d]))
/
sum(rate(data_freshness_seconds_bucket[30d]))
```

**Measurement Window**: 30 days rolling

## Database SLOs

### 1. Query Performance

**SLO**: 99% of database queries should complete within 100ms

**SLI**:
```promql
histogram_quantile(0.99,
  sum(rate(pg_stat_statements_total_time_bucket[30d])) by (le)
)
```

**Error Budget**:
- 1% of queries can exceed 100ms

**Measurement Window**: 30 days rolling

### 2. Connection Availability

**SLO**: Database connections should be available 99.99% of the time

**SLI**:
```promql
sum(pg_stat_database_numbackends) / sum(pg_settings_max_connections)
```

**Error Budget**:
- 30-day window: 4.32 minutes of unavailability

**Measurement Window**: 30 days rolling

### 3. Data Durability

**SLO**: 99.999% of committed transactions should be durable

**SLI**:
```promql
1 - (
  sum(pg_stat_database_xact_rollback)
  /
  sum(pg_stat_database_xact_commit + pg_stat_database_xact_rollback)
)
```

**Measurement Window**: 30 days rolling

## Notification Service SLOs

### 1. Delivery Success

**SLO**: 99% of notifications should be delivered successfully

**SLI**:
```promql
sum(rate(notifications_delivered_total[30d]))
/
sum(rate(notifications_sent_total[30d]))
```

**Error Budget**:
- 1% of notifications can fail
- On 100K notifications/day: 1K can fail

**Measurement Window**: 30 days rolling

### 2. Delivery Latency

**SLO**: 95% of notifications should be delivered within 10 seconds

**SLI**:
```promql
histogram_quantile(0.95,
  sum(rate(notification_delivery_duration_seconds_bucket[30d])) by (le)
)
```

**Measurement Window**: 30 days rolling

## Temporal Workflow SLOs

### 1. Workflow Success Rate

**SLO**: 99.5% of workflows should complete successfully

**SLI**:
```promql
sum(rate(temporal_workflow_completed_total[30d]))
/
sum(rate(temporal_workflow_started_total[30d]))
```

**Error Budget**:
- 0.5% of workflows can fail
- On 10K workflows/day: 50 can fail

**Measurement Window**: 30 days rolling

### 2. Workflow Execution Time

**SLO**: 90% of workflows should complete within their scheduled time + 10%

**SLI**:
```promql
histogram_quantile(0.90,
  sum(rate(temporal_workflow_execution_time_bucket[30d])) by (le)
)
```

**Measurement Window**: 30 days rolling

## Business Metrics SLOs

### 1. User Engagement

**SLO**: 80% of active users should create at least one task per week

**SLI**:
```promql
count(increase(tasks_created_total[7d]) > 0) by (user_id)
/
count(user_last_active[7d]) by (user_id)
```

**Measurement Window**: 7 days rolling

### 2. Task Completion Rate

**SLO**: 70% of created tasks should be completed within their due date

**SLI**:
```promql
sum(rate(tasks_completed_on_time_total[30d]))
/
sum(rate(tasks_created_total[30d]))
```

**Measurement Window**: 30 days rolling

## SLI Implementation

### Instrumentation

All services should instrument the following:

```go
// Request counter
httpRequests = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// Request duration histogram
httpDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)

// Data freshness gauge
dataFreshness = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "data_freshness_seconds",
        Help: "Age of data in seconds",
    },
    []string{"resource"},
)
```

### Recording Rules

Optimize queries with recording rules:

```yaml
# prometheus-rules.yml
groups:
- name: sli_recording_rules
  interval: 1m
  rules:
  # API availability
  - record: api:availability:ratio_30d
    expr: |
      sum(rate(http_requests_total{status!~"5.."}[30d]))
      /
      sum(rate(http_requests_total[30d]))

  # API latency P95
  - record: api:latency:p95_30d
    expr: |
      histogram_quantile(0.95,
        sum(rate(http_request_duration_seconds_bucket[30d])) by (le)
      )

  # Error budget remaining
  - record: api:error_budget:remaining_30d
    expr: |
      1 - (
        (1 - api:availability:ratio_30d) / (1 - 0.999)
      )
```

## SLO Monitoring Dashboard

### Grafana Dashboard Panels

**Availability Panel:**
```json
{
  "title": "API Availability (30d)",
  "targets": [{
    "expr": "api:availability:ratio_30d",
    "legendFormat": "Availability"
  }],
  "thresholds": [
    { "value": 0.999, "color": "green" },
    { "value": 0.9985, "color": "yellow" },
    { "value": 0, "color": "red" }
  ]
}
```

**Error Budget Panel:**
```json
{
  "title": "Error Budget Remaining (30d)",
  "targets": [{
    "expr": "api:error_budget:remaining_30d * 100",
    "legendFormat": "Error Budget %"
  }],
  "thresholds": [
    { "value": 50, "color": "green" },
    { "value": 25, "color": "yellow" },
    { "value": 0, "color": "red" }
  ]
}
```

**Burn Rate Panel:**
```json
{
  "title": "Error Budget Burn Rate",
  "targets": [{
    "expr": "1 - api:availability:ratio_1h",
    "legendFormat": "1h burn rate"
  }, {
    "expr": "1 - api:availability:ratio_6h",
    "legendFormat": "6h burn rate"
  }]
}
```

## Alerting on SLOs

### Multi-Window, Multi-Burn-Rate Alerts

```yaml
# High burn rate (fast burn)
- alert: ErrorBudgetFastBurn
  expr: |
    (1 - api:availability:ratio_1h) / (1 - 0.999) > 14.4
    and
    (1 - api:availability:ratio_5m) / (1 - 0.999) > 14.4
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Error budget burning at 14.4x rate"
    description: "At this rate, entire 30-day error budget will be consumed in 2 hours"

# Medium burn rate (slow burn)
- alert: ErrorBudgetSlowBurn
  expr: |
    (1 - api:availability:ratio_6h) / (1 - 0.999) > 6
    and
    (1 - api:availability:ratio_30m) / (1 - 0.999) > 6
  for: 15m
  labels:
    severity: warning
  annotations:
    summary: "Error budget burning at 6x rate"
    description: "At this rate, entire 30-day error budget will be consumed in 5 days"
```

## SLO Review Process

### Weekly Review

- Review current SLI values
- Check error budget consumption
- Identify trends
- Plan improvements

### Monthly Review

- Validate SLO targets
- Review incidents and their impact on SLOs
- Update error budget policies
- Adjust SLO targets if needed

### Quarterly Review

- Comprehensive SLO review
- Update SLO documentation
- Refine measurement methodology
- Set SLO roadmap

## Error Budget Policy

### When Error Budget is Healthy (> 50%)

- Continue feature development
- Accept some risk in deployments
- Focus on innovation

### When Error Budget is Low (< 50%)

- Freeze feature development
- Focus on reliability improvements
- Increase testing and validation
- Conduct blameless postmortems

### When Error Budget is Exhausted (< 10%)

- Stop all feature deployments
- Emergency reliability improvements only
- Root cause analysis for all incidents
- Implement preventive measures

## Calculating Error Budget

```python
# Python example
def calculate_error_budget(slo_target, actual_availability, days=30):
    """
    Calculate remaining error budget
    
    Args:
        slo_target: Target availability (e.g., 0.999 for 99.9%)
        actual_availability: Actual availability (e.g., 0.9995)
        days: Measurement window in days
    
    Returns:
        Remaining error budget as percentage
    """
    max_error_rate = 1 - slo_target
    actual_error_rate = 1 - actual_availability
    
    if actual_error_rate >= max_error_rate:
        return 0.0
    
    remaining = 1 - (actual_error_rate / max_error_rate)
    return remaining * 100

# Example
slo = 0.999
actual = 0.9995
budget = calculate_error_budget(slo, actual)
print(f"Error budget remaining: {budget:.2f}%")
```

## Best Practices

1. **Start Simple**: Begin with a few critical SLOs
2. **User-Centric**: Focus on user-visible metrics
3. **Measurable**: Ensure SLIs are easily measurable
4. **Achievable**: Set realistic SLO targets
5. **Document**: Keep SLO documentation up to date
6. **Review**: Regularly review and adjust SLOs
7. **Communicate**: Share SLO status with stakeholders
8. **Act**: Use error budget to guide decision-making

## References

- [Google SRE Book - Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
- [Google SRE Workbook - Implementing SLOs](https://sre.google/workbook/implementing-slos/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)

## License

Copyright © 2024 House Helper. All rights reserved.
