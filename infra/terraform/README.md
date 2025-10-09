# Terraform AWS Infrastructure

This directory contains Terraform configuration for deploying the House Helper application to AWS.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.6
- [AWS CLI](https://aws.amazon.com/cli/) configured with appropriate credentials
- An S3 bucket for Terraform state storage (created manually before first deployment)

## Infrastructure Components

The Terraform configuration provisions the following AWS resources:

### Networking (vpc.tf)
- VPC with public, private, and database subnets across 3 availability zones
- NAT Gateway (single for non-prod, one per AZ for prod)
- VPC Flow Logs to CloudWatch
- VPC Endpoints for S3 and ECR (cost optimization)

### Compute (eks.tf)
- EKS Kubernetes cluster (v1.28)
- Managed node group with t3.medium/large instances
- Cluster addons: CoreDNS, kube-proxy, VPC CNI, EBS CSI driver
- IRSA (IAM Roles for Service Accounts) enabled
- IAM roles for AWS Load Balancer Controller and External DNS

### Database (rds.tf)
- RDS PostgreSQL 16.1 with Multi-AZ deployment
- Automated backups (7-day retention for prod)
- Performance Insights and Enhanced Monitoring
- Read replica for production environment
- Credentials stored in AWS Secrets Manager

### Cache (elasticache.tf)
- ElastiCache Redis 7.0 replication group
- Multi-AZ with automatic failover (production)
- Encryption at rest and in transit
- SCRAM authentication with Secrets Manager integration

### Messaging (msk.tf)
- MSK (Managed Streaming for Apache Kafka) cluster
- Kafka 3.6.0 with multiple brokers across AZs
- Encryption at rest with KMS and in transit with TLS
- SASL/SCRAM authentication
- CloudWatch and S3 logging

### Storage (s3.tf)
- S3 bucket for application assets with versioning and CORS
- S3 bucket for backups with KMS encryption and lifecycle policies
- S3 bucket for logs with retention policies

### Container Registry (ecr.tf)
- ECR repositories for all microservices (API, Notifier, Temporal Worker/API, Kafka Consumer)
- Image scanning on push
- Lifecycle policies to manage image retention
- Pull-through cache for public registries (optional)

### IAM (iam.tf)
- GitHub Actions OIDC provider for CI/CD (optional)
- Service account roles for application pods (IRSA)
- Roles for RDS Enhanced Monitoring and VPC Flow Logs
- Policies for S3, Secrets Manager, and CloudWatch access

### Secrets Management (secrets.tf)
- Application secrets (JWT, API keys, OAuth credentials)
- External service credentials (Twilio, SendGrid, Sentry, Datadog)
- Temporal workflow secrets
- KMS encryption for all secrets
- External Secrets Operator integration for Kubernetes

### Monitoring (monitoring.tf)
- CloudWatch Dashboard with metrics for all services
- Log groups for application and infrastructure components
- Metric filters for error tracking
- CloudWatch Alarms for CPU, memory, disk usage
- SNS topics for alert notifications
- CloudWatch Insights query definitions

## Directory Structure

```
infra/terraform/
├── main.tf              # Provider configuration and backend
├── variables.tf         # Input variables
├── outputs.tf           # Output values
├── vpc.tf              # VPC and networking
├── eks.tf              # EKS cluster
├── rds.tf              # RDS PostgreSQL
├── elasticache.tf      # ElastiCache Redis
├── msk.tf              # MSK Kafka
├── s3.tf               # S3 buckets
├── ecr.tf              # Container registries
├── iam.tf              # IAM roles and policies
├── secrets.tf          # Secrets Manager
├── monitoring.tf       # CloudWatch monitoring
├── terraform.tfvars    # Variable values (create from example)
└── README.md           # This file
```

## Initial Setup

1. **Create S3 bucket for Terraform state:**

```bash
aws s3api create-bucket \
  --bucket house-helper-terraform-state \
  --region us-east-1

aws s3api put-bucket-versioning \
  --bucket house-helper-terraform-state \
  --versioning-configuration Status=Enabled

aws s3api put-bucket-encryption \
  --bucket house-helper-terraform-state \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'
```

2. **Create DynamoDB table for state locking:**

```bash
aws dynamodb create-table \
  --table-name house-helper-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

3. **Create terraform.tfvars file:**

```hcl
# Basic Configuration
project_name = "house-helper"
environment  = "dev"
aws_region   = "us-east-1"

# Networking
vpc_cidr             = "10.0.0.0/16"
availability_zones   = ["us-east-1a", "us-east-1b", "us-east-1c"]

# EKS Configuration
eks_cluster_version    = "1.28"
eks_node_instance_type = "t3.medium"
eks_node_desired_size  = 2
eks_node_min_size      = 1
eks_node_max_size      = 10

# RDS Configuration
rds_instance_class          = "db.t3.medium"
rds_allocated_storage       = 100
rds_max_allocated_storage   = 500
rds_backup_retention_period = 7
rds_enable_read_replica     = false

# ElastiCache Configuration
elasticache_node_type         = "cache.t3.medium"
elasticache_num_cache_nodes   = 2

# MSK Configuration
msk_instance_type           = "kafka.t3.small"
msk_number_of_broker_nodes  = 3
msk_ebs_volume_size         = 100

# Monitoring
cloudwatch_log_retention_days = 30
alert_email                   = "alerts@example.com"

# Application Configuration
app_log_level            = "info"
max_upload_size_mb       = 50
session_timeout_minutes  = 30

# GitHub Actions (optional)
enable_github_actions_oidc = true
github_org                 = "your-org"
github_repo                = "House-Helper"
```

## Deployment

### Initialize Terraform

```bash
cd infra/terraform
terraform init
```

### Plan the deployment

```bash
terraform plan -out=tfplan
```

### Apply the configuration

```bash
terraform apply tfplan
```

### Destroy infrastructure (when needed)

```bash
terraform destroy
```

## Environments

The configuration supports multiple environments (dev, staging, prod) using workspaces:

```bash
# Create and switch to production workspace
terraform workspace new prod
terraform workspace select prod

# Deploy to production
terraform apply -var-file=prod.tfvars
```

## Outputs

After successful deployment, Terraform outputs important values:

- VPC ID and subnet IDs
- EKS cluster endpoint and configuration command
- RDS endpoint and Secrets Manager ARN
- ElastiCache Redis endpoint
- MSK Kafka bootstrap brokers
- S3 bucket names
- ECR repository URLs
- IAM role ARNs

View outputs:

```bash
terraform output
```

Get specific output:

```bash
terraform output -raw eks_cluster_endpoint
```

## Connecting to EKS Cluster

After deployment, configure kubectl:

```bash
aws eks update-kubeconfig \
  --region $(terraform output -raw aws_region) \
  --name $(terraform output -raw eks_cluster_name)
```

Verify connection:

```bash
kubectl get nodes
kubectl get pods --all-namespaces
```

## Cost Optimization

For development environments:

1. Use smaller instance types:
   - EKS nodes: t3.medium
   - RDS: db.t3.medium
   - ElastiCache: cache.t3.medium
   - MSK: kafka.t3.small

2. Reduce redundancy:
   - Single NAT Gateway
   - Disable read replicas
   - Reduce backup retention

3. Use spot instances for EKS nodes (optional):
   ```hcl
   capacity_type = "SPOT"
   ```

For production environments:

1. Enable Multi-AZ deployments
2. Use larger instance types
3. Enable read replicas
4. Increase backup retention
5. Enable CloudWatch Container Insights

## Security Considerations

1. **Secrets Management:**
   - All sensitive data stored in AWS Secrets Manager
   - KMS encryption enabled for all secrets
   - Use External Secrets Operator in Kubernetes

2. **Network Security:**
   - Private subnets for all workloads
   - Security groups with least privilege access
   - VPC Flow Logs enabled
   - VPC endpoints for AWS services

3. **Data Encryption:**
   - Encryption at rest for RDS, ElastiCache, S3, MSK
   - Encryption in transit with TLS
   - KMS key rotation enabled

4. **Access Control:**
   - IAM roles with least privilege
   - IRSA for pod-level permissions
   - GitHub Actions OIDC for CI/CD (no long-lived credentials)

5. **Monitoring:**
   - CloudWatch Alarms for anomalies
   - VPC Flow Logs for network analysis
   - Container Insights for EKS
   - Image scanning in ECR

## Maintenance

### Upgrading EKS

1. Update `eks_cluster_version` in variables
2. Plan and apply changes
3. Update node groups separately if needed

### Rotating Secrets

1. For production, enable automatic rotation:
   ```hcl
   secrets_rotation_lambda_arn = "arn:aws:lambda:..."
   ```

2. Manual rotation:
   ```bash
   aws secretsmanager rotate-secret --secret-id <secret-name>
   ```

### Backup and Recovery

1. **RDS Backups:**
   - Automated daily backups
   - Manual snapshots before major changes
   - Point-in-time recovery available

2. **S3 Backups:**
   - Versioning enabled
   - Lifecycle policies for cost optimization
   - Cross-region replication (optional)

## Troubleshooting

### EKS Node Issues

```bash
# Check node status
kubectl get nodes

# Describe node
kubectl describe node <node-name>

# Check node group in AWS
aws eks describe-nodegroup \
  --cluster-name $(terraform output -raw eks_cluster_name) \
  --nodegroup-name main
```

### RDS Connection Issues

```bash
# Get RDS endpoint
terraform output rds_endpoint

# Test connection from EKS pod
kubectl run -it --rm psql \
  --image=postgres:16 \
  --restart=Never \
  -- psql -h <rds-endpoint> -U house_helper -d house_helper
```

### Secrets Access Issues

```bash
# Verify secrets exist
aws secretsmanager list-secrets

# Get secret value
aws secretsmanager get-secret-value \
  --secret-id $(terraform output -raw rds_credentials_secret_arn)
```

## Support

For issues or questions:
1. Check AWS service health: https://status.aws.amazon.com/
2. Review CloudWatch Logs and Alarms
3. Check Terraform state: `terraform show`
4. Validate configuration: `terraform validate`

## License

Copyright © 2024 House Helper. All rights reserved.
