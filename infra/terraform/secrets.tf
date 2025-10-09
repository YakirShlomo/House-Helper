# Application Secrets
resource "aws_secretsmanager_secret" "app_secrets" {
  name        = "${var.project_name}-${var.environment}-app-secrets"
  description = "Application secrets for House Helper"
  kms_key_id  = aws_kms_key.secrets.id

  tags = {
    Name        = "${var.project_name}-${var.environment}-app-secrets"
    Environment = var.environment
  }
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = true
}

resource "random_password" "api_key" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret_version" "app_secrets" {
  secret_id = aws_secretsmanager_secret.app_secrets.id
  secret_string = jsonencode({
    jwt_secret              = random_password.jwt_secret.result
    api_key                 = random_password.api_key.result
    firebase_server_key     = var.firebase_server_key
    expo_push_token_secret  = var.expo_push_token_secret
    google_oauth_client_id  = var.google_oauth_client_id
    google_oauth_secret     = var.google_oauth_secret
  })
}

# External Services Secrets
resource "aws_secretsmanager_secret" "external_services" {
  name        = "${var.project_name}-${var.environment}-external-services"
  description = "External service credentials and API keys"
  kms_key_id  = aws_kms_key.secrets.id

  tags = {
    Name        = "${var.project_name}-${var.environment}-external-services"
    Environment = var.environment
  }
}

resource "aws_secretsmanager_secret_version" "external_services" {
  secret_id = aws_secretsmanager_secret.external_services.id
  secret_string = jsonencode({
    twilio_account_sid   = var.twilio_account_sid
    twilio_auth_token    = var.twilio_auth_token
    twilio_phone_number  = var.twilio_phone_number
    sendgrid_api_key     = var.sendgrid_api_key
    sentry_dsn           = var.sentry_dsn
    datadog_api_key      = var.datadog_api_key
  })
}

# Temporal Secrets
resource "aws_secretsmanager_secret" "temporal_secrets" {
  name        = "${var.project_name}-${var.environment}-temporal-secrets"
  description = "Temporal workflow engine secrets"
  kms_key_id  = aws_kms_key.secrets.id

  tags = {
    Name        = "${var.project_name}-${var.environment}-temporal-secrets"
    Environment = var.environment
  }
}

resource "random_password" "temporal_admin_password" {
  length  = 32
  special = true
}

resource "aws_secretsmanager_secret_version" "temporal_secrets" {
  secret_id = aws_secretsmanager_secret.temporal_secrets.id
  secret_string = jsonencode({
    admin_username = "temporal-admin"
    admin_password = random_password.temporal_admin_password.result
    namespace      = var.project_name
  })
}

# KMS Key for Secrets Manager
resource "aws_kms_key" "secrets" {
  description             = "KMS key for Secrets Manager encryption"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "${var.project_name}-${var.environment}-secrets-kms"
  }
}

resource "aws_kms_alias" "secrets" {
  name          = "alias/${var.project_name}-${var.environment}-secrets"
  target_key_id = aws_kms_key.secrets.key_id
}

# KMS Key Policy to allow Secrets Manager to use the key
resource "aws_kms_key_policy" "secrets" {
  key_id = aws_kms_key.secrets.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow Secrets Manager to use the key"
        Effect = "Allow"
        Principal = {
          Service = "secretsmanager.amazonaws.com"
        }
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = "*"
      },
      {
        Sid    = "Allow EKS pods to decrypt secrets"
        Effect = "Allow"
        Principal = {
          AWS = [
            aws_iam_role.app_pods.arn,
            aws_iam_role.temporal_workflows.arn
          ]
        }
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })
}

# Secret Rotation Lambda (placeholder - would need actual implementation)
# For production, implement proper secret rotation for RDS, Redis, etc.
resource "aws_secretsmanager_secret_rotation" "rds_credentials" {
  count = var.environment == "prod" ? 1 : 0

  secret_id           = aws_secretsmanager_secret.rds_credentials.id
  rotation_lambda_arn = var.secrets_rotation_lambda_arn

  rotation_rules {
    automatically_after_days = 30
  }
}

# CloudWatch Alarms for Secrets Manager
resource "aws_cloudwatch_metric_alarm" "secrets_access_denied" {
  alarm_name          = "${var.project_name}-${var.environment}-secrets-access-denied"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "UserErrorCount"
  namespace           = "AWS/SecretsManager"
  period              = "300"
  statistic           = "Sum"
  threshold           = "5"
  alarm_description   = "Alert when there are repeated access denied errors to secrets"
  treat_missing_data  = "notBreaching"

  dimensions = {
    SecretId = aws_secretsmanager_secret.app_secrets.id
  }
}

# SSM Parameters for non-sensitive configuration
resource "aws_ssm_parameter" "app_config" {
  name        = "/${var.project_name}/${var.environment}/config/app"
  description = "Application configuration"
  type        = "String"
  value = jsonencode({
    log_level           = var.app_log_level
    max_upload_size_mb  = var.max_upload_size_mb
    session_timeout_min = var.session_timeout_minutes
    enable_debug        = var.environment != "prod"
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-app-config"
    Environment = var.environment
  }
}

resource "aws_ssm_parameter" "db_config" {
  name        = "/${var.project_name}/${var.environment}/config/database"
  description = "Database configuration"
  type        = "String"
  value = jsonencode({
    max_connections     = var.rds_max_connections
    connection_timeout  = var.rds_connection_timeout
    pool_size           = var.rds_connection_pool_size
    enable_query_log    = var.environment != "prod"
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-db-config"
    Environment = var.environment
  }
}

resource "aws_ssm_parameter" "redis_config" {
  name        = "/${var.project_name}/${var.environment}/config/redis"
  description = "Redis configuration"
  type        = "String"
  value = jsonencode({
    max_retries         = 3
    retry_delay_ms      = 100
    connection_timeout  = 5000
    command_timeout     = 3000
    enable_cluster_mode = false
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-redis-config"
    Environment = var.environment
  }
}

resource "aws_ssm_parameter" "kafka_config" {
  name        = "/${var.project_name}/${var.environment}/config/kafka"
  description = "Kafka configuration"
  type        = "String"
  value = jsonencode({
    consumer_group      = "${var.project_name}-${var.environment}-consumers"
    session_timeout_ms  = 30000
    heartbeat_interval  = 3000
    max_poll_records    = 500
    enable_auto_commit  = false
    auto_offset_reset   = "earliest"
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-kafka-config"
    Environment = var.environment
  }
}

# External Secrets Operator integration (for Kubernetes)
# This allows pods to sync secrets from AWS Secrets Manager
resource "aws_iam_policy" "external_secrets_operator" {
  name        = "${var.project_name}-${var.environment}-external-secrets-operator"
  description = "Allow External Secrets Operator to read secrets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:ListSecrets"
        ]
        Resource = [
          aws_secretsmanager_secret.app_secrets.arn,
          aws_secretsmanager_secret.external_services.arn,
          aws_secretsmanager_secret.temporal_secrets.arn,
          aws_secretsmanager_secret.rds_credentials.arn,
          aws_secretsmanager_secret.redis_auth_token.arn,
          aws_secretsmanager_secret.msk_credentials.arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = aws_kms_key.secrets.arn
      }
    ]
  })
}

resource "aws_iam_role" "external_secrets_operator" {
  name = "${var.project_name}-${var.environment}-external-secrets-operator"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = module.eks.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${module.eks.oidc_provider}:sub" = "system:serviceaccount:external-secrets:external-secrets-operator"
            "${module.eks.oidc_provider}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-external-secrets-operator"
  }
}

resource "aws_iam_role_policy_attachment" "external_secrets_operator" {
  role       = aws_iam_role.external_secrets_operator.name
  policy_arn = aws_iam_policy.external_secrets_operator.arn
}
