# VPC Outputs
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "CIDR block of the VPC"
  value       = module.vpc.vpc_cidr_block
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = module.vpc.public_subnets
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = module.vpc.private_subnets
}

output "database_subnet_ids" {
  description = "IDs of database subnets"
  value       = module.vpc.database_subnets
}

# EKS Outputs
output "eks_cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "eks_cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_oidc_provider_arn" {
  description = "ARN of the OIDC Provider for EKS"
  value       = module.eks.oidc_provider_arn
}

output "configure_kubectl" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

# RDS Outputs
output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.db_instance_endpoint
}

output "rds_port" {
  description = "RDS instance port"
  value       = module.rds.db_instance_port
}

output "rds_database_name" {
  description = "Name of the database"
  value       = module.rds.db_instance_name
}

output "rds_username" {
  description = "Master username for RDS"
  value       = module.rds.db_instance_username
  sensitive   = true
}

output "rds_password_secret_arn" {
  description = "ARN of Secrets Manager secret containing RDS password"
  value       = aws_secretsmanager_secret.rds_password.arn
}

# ElastiCache Outputs
output "elasticache_endpoint" {
  description = "ElastiCache cluster endpoint"
  value       = aws_elasticache_replication_group.redis.configuration_endpoint_address
}

output "elasticache_port" {
  description = "ElastiCache port"
  value       = aws_elasticache_replication_group.redis.port
}

# MSK Outputs
output "msk_bootstrap_brokers" {
  description = "MSK bootstrap brokers"
  value       = aws_msk_cluster.kafka.bootstrap_brokers
}

output "msk_bootstrap_brokers_tls" {
  description = "MSK bootstrap brokers (TLS)"
  value       = aws_msk_cluster.kafka.bootstrap_brokers_tls
}

output "msk_zookeeper_connect_string" {
  description = "MSK Zookeeper connection string"
  value       = aws_msk_cluster.kafka.zookeeper_connect_string
}

output "msk_cluster_arn" {
  description = "MSK cluster ARN"
  value       = aws_msk_cluster.kafka.arn
}

# S3 Outputs
output "s3_assets_bucket" {
  description = "Name of S3 bucket for assets"
  value       = aws_s3_bucket.assets.id
}

output "s3_backups_bucket" {
  description = "Name of S3 bucket for backups"
  value       = aws_s3_bucket.backups.id
}

output "s3_logs_bucket" {
  description = "Name of S3 bucket for logs"
  value       = aws_s3_bucket.logs.id
}

# ECR Outputs
output "ecr_api_repository_url" {
  description = "URL of ECR repository for API service"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_notifier_repository_url" {
  description = "URL of ECR repository for notifier service"
  value       = aws_ecr_repository.notifier.repository_url
}

output "ecr_temporal_worker_repository_url" {
  description = "URL of ECR repository for temporal worker"
  value       = aws_ecr_repository.temporal_worker.repository_url
}

output "ecr_temporal_api_repository_url" {
  description = "URL of ECR repository for temporal API"
  value       = aws_ecr_repository.temporal_api.repository_url
}

output "ecr_kafka_consumer_repository_url" {
  description = "URL of ECR repository for Kafka consumer"
  value       = aws_ecr_repository.kafka_consumer.repository_url
}

# IAM Outputs
output "eks_node_role_arn" {
  description = "ARN of IAM role for EKS nodes"
  value       = module.eks.eks_managed_node_groups["main"].iam_role_arn
}

output "alb_controller_role_arn" {
  description = "ARN of IAM role for AWS Load Balancer Controller"
  value       = aws_iam_role.alb_controller.arn
}

output "external_dns_role_arn" {
  description = "ARN of IAM role for External DNS"
  value       = aws_iam_role.external_dns.arn
}

# Secrets Manager Outputs
output "jwt_secret_arn" {
  description = "ARN of JWT secret in Secrets Manager"
  value       = aws_secretsmanager_secret.jwt_secret.arn
}

output "fcm_credentials_secret_arn" {
  description = "ARN of FCM credentials secret"
  value       = aws_secretsmanager_secret.fcm_credentials.arn
}

output "apns_credentials_secret_arn" {
  description = "ARN of APNS credentials secret"
  value       = aws_secretsmanager_secret.apns_credentials.arn
}

# CloudWatch Outputs
output "cloudwatch_log_group_eks" {
  description = "CloudWatch log group for EKS"
  value       = aws_cloudwatch_log_group.eks.name
}

output "cloudwatch_log_group_api" {
  description = "CloudWatch log group for API service"
  value       = aws_cloudwatch_log_group.api.name
}
