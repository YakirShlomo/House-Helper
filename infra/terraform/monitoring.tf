# CloudWatch Dashboard for Application Monitoring
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project_name}-${var.environment}-main"

  dashboard_body = jsonencode({
    widgets = [
      # EKS Cluster Health
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/EKS", "cluster_failed_node_count", { stat = "Average" }],
            [".", "cluster_node_count", { stat = "Average" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "EKS Cluster Nodes"
          yAxis = {
            left = {
              min = 0
            }
          }
        }
      },
      # RDS Performance
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/RDS", "CPUUtilization", { stat = "Average", color = "#ff7f0e" }],
            [".", "DatabaseConnections", { stat = "Average", yAxis = "right" }],
            [".", "FreeableMemory", { stat = "Average", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "RDS Database Metrics"
          yAxis = {
            left = {
              label = "CPU %"
              min   = 0
              max   = 100
            }
            right = {
              label = "Connections / Memory"
            }
          }
        }
      },
      # ElastiCache Redis
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/ElastiCache", "CPUUtilization", { stat = "Average" }],
            [".", "DatabaseMemoryUsagePercentage", { stat = "Average" }],
            [".", "NetworkBytesIn", { stat = "Sum", yAxis = "right" }],
            [".", "NetworkBytesOut", { stat = "Sum", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "ElastiCache Redis Metrics"
          yAxis = {
            left = {
              label = "CPU / Memory %"
              min   = 0
              max   = 100
            }
            right = {
              label = "Network Bytes"
            }
          }
        }
      },
      # MSK Kafka
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/Kafka", "CpuUser", { stat = "Average" }],
            [".", "BytesInPerSec", { stat = "Sum", yAxis = "right" }],
            [".", "BytesOutPerSec", { stat = "Sum", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "MSK Kafka Metrics"
        }
      },
      # Application Load Balancer
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", { stat = "Average" }],
            [".", "RequestCount", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_2XX_Count", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_4XX_Count", { stat = "Sum", yAxis = "right" }],
            [".", "HTTPCode_Target_5XX_Count", { stat = "Sum", yAxis = "right" }]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Application Load Balancer Metrics"
        }
      },
      # S3 Bucket Metrics
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/S3", "BucketSizeBytes", { stat = "Average" }],
            [".", "NumberOfObjects", { stat = "Average", yAxis = "right" }]
          ]
          period = 86400
          stat   = "Average"
          region = var.aws_region
          title  = "S3 Storage Metrics"
        }
      }
    ]
  })
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "application" {
  name              = "/aws/eks/${var.project_name}-${var.environment}/application"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-application-logs"
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/eks/${var.project_name}-${var.environment}/api"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-api-logs"
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "notifier" {
  name              = "/aws/eks/${var.project_name}-${var.environment}/notifier"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-notifier-logs"
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "temporal_worker" {
  name              = "/aws/eks/${var.project_name}-${var.environment}/temporal-worker"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-temporal-worker-logs"
    Environment = var.environment
  }
}

resource "aws_cloudwatch_log_group" "kafka_consumer" {
  name              = "/aws/eks/${var.project_name}-${var.environment}/kafka-consumer"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-kafka-consumer-logs"
    Environment = var.environment
  }
}

# CloudWatch Metric Filters
resource "aws_cloudwatch_log_metric_filter" "error_count" {
  name           = "${var.project_name}-${var.environment}-error-count"
  log_group_name = aws_cloudwatch_log_group.application.name
  pattern        = "[time, request_id, level=ERROR*, ...]"

  metric_transformation {
    name      = "ErrorCount"
    namespace = "${var.project_name}/${var.environment}"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "fatal_count" {
  name           = "${var.project_name}-${var.environment}-fatal-count"
  log_group_name = aws_cloudwatch_log_group.application.name
  pattern        = "[time, request_id, level=FATAL*, ...]"

  metric_transformation {
    name      = "FatalCount"
    namespace = "${var.project_name}/${var.environment}"
    value     = "1"
  }
}

# CloudWatch Alarms for Application Metrics
resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  alarm_name          = "${var.project_name}-${var.environment}-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorCount"
  namespace           = "${var.project_name}/${var.environment}"
  period              = "300"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "Alert when error rate is high"
  treat_missing_data  = "notBreaching"
  alarm_actions       = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "fatal_errors" {
  alarm_name          = "${var.project_name}-${var.environment}-fatal-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "FatalCount"
  namespace           = "${var.project_name}/${var.environment}"
  period              = "60"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "Alert on any fatal errors"
  treat_missing_data  = "notBreaching"
  alarm_actions       = [aws_sns_topic.alerts.arn]
}

# SNS Topic for CloudWatch Alarms
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-${var.environment}-alerts"

  tags = {
    Name        = "${var.project_name}-${var.environment}-alerts"
    Environment = var.environment
  }
}

# SNS Topic Subscription (email)
resource "aws_sns_topic_subscription" "alerts_email" {
  count = var.alert_email != "" ? 1 : 0

  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# CloudWatch Composite Alarm for System Health
resource "aws_cloudwatch_composite_alarm" "system_health" {
  alarm_name          = "${var.project_name}-${var.environment}-system-health"
  alarm_description   = "Composite alarm for overall system health"
  actions_enabled     = true
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]
  insufficient_data_actions = []

  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.rds_cpu.alarm_name}) OR ALARM(${aws_cloudwatch_metric_alarm.redis_cpu.alarm_name}) OR ALARM(${aws_cloudwatch_metric_alarm.msk_cpu.alarm_name}) OR ALARM(${aws_cloudwatch_metric_alarm.high_error_rate.alarm_name})"
}

# CloudWatch Insights Queries
resource "aws_cloudwatch_query_definition" "error_analysis" {
  name = "${var.project_name}-${var.environment}-error-analysis"

  log_group_names = [
    aws_cloudwatch_log_group.application.name,
    aws_cloudwatch_log_group.api.name,
    aws_cloudwatch_log_group.notifier.name
  ]

  query_string = <<EOF
fields @timestamp, @message, level, error, stack_trace
| filter level = "ERROR" or level = "FATAL"
| sort @timestamp desc
| limit 100
EOF
}

resource "aws_cloudwatch_query_definition" "slow_queries" {
  name = "${var.project_name}-${var.environment}-slow-queries"

  log_group_names = [
    aws_cloudwatch_log_group.api.name
  ]

  query_string = <<EOF
fields @timestamp, @message, query, duration_ms
| filter query_type = "database" and duration_ms > 1000
| sort duration_ms desc
| limit 50
EOF
}

resource "aws_cloudwatch_query_definition" "request_latency" {
  name = "${var.project_name}-${var.environment}-request-latency"

  log_group_names = [
    aws_cloudwatch_log_group.api.name
  ]

  query_string = <<EOF
fields @timestamp, method, path, status_code, duration_ms
| stats avg(duration_ms) as avg_latency, max(duration_ms) as max_latency, count() as request_count by path
| sort avg_latency desc
EOF
}

# CloudWatch Event Rules for Automation
resource "aws_cloudwatch_event_rule" "rds_backup_completion" {
  name        = "${var.project_name}-${var.environment}-rds-backup-completion"
  description = "Trigger on RDS backup completion"

  event_pattern = jsonencode({
    source      = ["aws.rds"]
    detail-type = ["RDS DB Snapshot Event"]
    detail = {
      EventCategories = ["backup"]
    }
  })
}

resource "aws_cloudwatch_event_target" "rds_backup_sns" {
  rule      = aws_cloudwatch_event_rule.rds_backup_completion.name
  target_id = "SendToSNS"
  arn       = aws_sns_topic.alerts.arn
}

# CloudWatch Container Insights for EKS
# Enable Container Insights on EKS cluster
resource "null_resource" "enable_container_insights" {
  provisioner "local-exec" {
    command = <<EOF
      aws eks update-cluster-config \
        --region ${var.aws_region} \
        --name ${module.eks.cluster_name} \
        --logging '{"clusterLogging":[{"types":["api","audit","authenticator","controllerManager","scheduler"],"enabled":true}]}'
    EOF
  }

  depends_on = [module.eks]
}
