# Production Environment Configuration
# Use this file for production deployments: terraform apply -var-file=prod.tfvars

# Basic Configuration
project_name = "house-helper"
environment  = "prod"
aws_region   = "us-east-1"

# Networking
vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]

# EKS Configuration - Production sizing
eks_cluster_version    = "1.28"
eks_node_instance_type = "t3.large"
eks_node_desired_size  = 4
eks_node_min_size      = 2
eks_node_max_size      = 20
eks_disk_size          = 100

# RDS Configuration - Production with high availability
rds_instance_class          = "db.r6g.xlarge"
rds_allocated_storage       = 200
rds_max_allocated_storage   = 1000
rds_backup_retention_period = 30
rds_enable_read_replica     = true
rds_max_connections         = 500
rds_connection_timeout      = 10
rds_connection_pool_size    = 50

# ElastiCache Configuration - Production with Multi-AZ
elasticache_node_type       = "cache.r6g.large"
elasticache_num_cache_nodes = 3

# MSK Configuration - Production with more brokers
msk_kafka_version          = "3.6.0"
msk_instance_type          = "kafka.m5.large"
msk_number_of_broker_nodes = 6
msk_ebs_volume_size        = 500

# Monitoring - Extended retention for production
cloudwatch_log_retention_days = 90
alert_email                   = "production-alerts@example.com"

# Application Configuration
app_log_level           = "warn"
max_upload_size_mb      = 100
session_timeout_minutes = 60

# S3 Configuration
s3_bucket_size_alert_threshold = 536870912000  # 500 GB

# ECR Configuration
enable_ecr_pull_through_cache = true

# GitHub Actions
enable_github_actions_oidc = true
github_org                 = "your-org"
github_repo                = "House-Helper"

# Note: Sensitive values should be set via environment variables:
# export TF_VAR_firebase_server_key="..."
# export TF_VAR_expo_push_token_secret="..."
# export TF_VAR_google_oauth_client_id="..."
# export TF_VAR_google_oauth_secret="..."
# export TF_VAR_twilio_account_sid="..."
# export TF_VAR_twilio_auth_token="..."
# export TF_VAR_twilio_phone_number="..."
# export TF_VAR_sendgrid_api_key="..."
# export TF_VAR_sentry_dsn="..."
# export TF_VAR_datadog_api_key="..."
