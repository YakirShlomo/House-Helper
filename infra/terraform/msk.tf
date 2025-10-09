# MSK Kafka Cluster
resource "aws_msk_cluster" "kafka" {
  cluster_name           = "${var.project_name}-${var.environment}-kafka"
  kafka_version          = var.msk_kafka_version
  number_of_broker_nodes = var.msk_number_of_broker_nodes

  broker_node_group_info {
    instance_type   = var.msk_instance_type
    client_subnets  = module.vpc.private_subnets
    security_groups = [aws_security_group.msk.id]

    storage_info {
      ebs_storage_info {
        volume_size            = var.msk_ebs_volume_size
        provisioned_throughput {
          enabled           = true
          volume_throughput = 250
        }
      }
    }

    connectivity_info {
      public_access {
        type = "DISABLED"
      }
    }
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.msk.arn
    
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.kafka.arn
    revision = aws_msk_configuration.kafka.latest_revision
  }

  client_authentication {
    sasl {
      scram = true
    }
    unauthenticated = false
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.msk.name
      }
      s3 {
        enabled = true
        bucket  = aws_s3_bucket.logs.id
        prefix  = "msk-logs"
      }
    }
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-kafka"
  }
}

# MSK Configuration
resource "aws_msk_configuration" "kafka" {
  name              = "${var.project_name}-${var.environment}-kafka-config"
  kafka_versions    = [var.msk_kafka_version]
  server_properties = <<PROPERTIES
auto.create.topics.enable=true
default.replication.factor=3
min.insync.replicas=2
num.io.threads=8
num.network.threads=5
num.partitions=3
num.replica.fetchers=2
replica.lag.time.max.ms=30000
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
socket.send.buffer.bytes=102400
unclean.leader.election.enable=false
zookeeper.session.timeout.ms=18000
log.retention.hours=168
log.retention.bytes=1073741824
log.segment.bytes=1073741824
PROPERTIES

  lifecycle {
    create_before_destroy = true
  }
}

# KMS Key for MSK encryption
resource "aws_kms_key" "msk" {
  description             = "KMS key for MSK encryption"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "${var.project_name}-${var.environment}-msk-kms"
  }
}

resource "aws_kms_alias" "msk" {
  name          = "alias/${var.project_name}-${var.environment}-msk"
  target_key_id = aws_kms_key.msk.key_id
}

# Security Group for MSK
resource "aws_security_group" "msk" {
  name        = "${var.project_name}-${var.environment}-msk-sg"
  description = "Security group for MSK Kafka cluster"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Kafka plaintext from EKS"
    from_port       = 9092
    to_port         = 9092
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  ingress {
    description     = "Kafka TLS from EKS"
    from_port       = 9094
    to_port         = 9094
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  ingress {
    description     = "Kafka SASL/SCRAM from EKS"
    from_port       = 9096
    to_port         = 9096
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  ingress {
    description = "Zookeeper from VPC"
    from_port   = 2181
    to_port     = 2181
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-msk-sg"
  }
}

# CloudWatch Log Group for MSK
resource "aws_cloudwatch_log_group" "msk" {
  name              = "/aws/msk/${var.project_name}-${var.environment}"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name = "${var.project_name}-${var.environment}-msk-logs"
  }
}

# MSK SCRAM Secret Association
resource "aws_msk_scram_secret_association" "kafka" {
  cluster_arn     = aws_msk_cluster.kafka.arn
  secret_arn_list = [aws_secretsmanager_secret.msk_credentials.arn]

  depends_on = [aws_secretsmanager_secret_version.msk_credentials]
}

# Secrets Manager for MSK credentials
resource "aws_secretsmanager_secret" "msk_credentials" {
  name        = "${var.project_name}-${var.environment}-msk-credentials"
  description = "MSK Kafka SCRAM credentials"
  kms_key_id  = aws_kms_key.msk.id

  tags = {
    Name = "${var.project_name}-${var.environment}-msk-credentials"
  }
}

resource "random_password" "msk_password" {
  length  = 32
  special = true
}

resource "aws_secretsmanager_secret_version" "msk_credentials" {
  secret_id = aws_secretsmanager_secret.msk_credentials.id
  secret_string = jsonencode({
    username = "house-helper-app"
    password = random_password.msk_password.result
  })
}

# CloudWatch Alarms for MSK
resource "aws_cloudwatch_metric_alarm" "msk_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-msk-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CpuUser"
  namespace           = "AWS/Kafka"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors MSK CPU utilization"
  alarm_actions       = [aws_sns_topic.msk_notifications.arn]

  dimensions = {
    "Cluster Name" = aws_msk_cluster.kafka.cluster_name
  }
}

resource "aws_cloudwatch_metric_alarm" "msk_disk_usage" {
  alarm_name          = "${var.project_name}-${var.environment}-msk-high-disk"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "KafkaDataLogsDiskUsed"
  namespace           = "AWS/Kafka"
  period              = "300"
  statistic           = "Average"
  threshold           = "85"
  alarm_description   = "This metric monitors MSK disk usage"
  alarm_actions       = [aws_sns_topic.msk_notifications.arn]

  dimensions = {
    "Cluster Name" = aws_msk_cluster.kafka.cluster_name
  }
}

resource "aws_cloudwatch_metric_alarm" "msk_offline_partitions" {
  alarm_name          = "${var.project_name}-${var.environment}-msk-offline-partitions"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "OfflinePartitionsCount"
  namespace           = "AWS/Kafka"
  period              = "60"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "This metric monitors MSK offline partitions"
  alarm_actions       = [aws_sns_topic.msk_notifications.arn]

  dimensions = {
    "Cluster Name" = aws_msk_cluster.kafka.cluster_name
  }
}

# SNS Topic for MSK notifications
resource "aws_sns_topic" "msk_notifications" {
  name = "${var.project_name}-${var.environment}-msk-notifications"

  tags = {
    Name = "${var.project_name}-${var.environment}-msk-notifications"
  }
}
