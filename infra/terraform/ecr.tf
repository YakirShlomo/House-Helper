# ECR Repository for API Service
resource "aws_ecr_repository" "api" {
  name                 = "${var.project_name}/${var.environment}/api"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-api"
    Service     = "api"
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Repository for Notifier Service
resource "aws_ecr_repository" "notifier" {
  name                 = "${var.project_name}/${var.environment}/notifier"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-notifier"
    Service     = "notifier"
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "notifier" {
  repository = aws_ecr_repository.notifier.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Repository for Temporal Worker
resource "aws_ecr_repository" "temporal_worker" {
  name                 = "${var.project_name}/${var.environment}/temporal-worker"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-temporal-worker"
    Service     = "temporal-worker"
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "temporal_worker" {
  repository = aws_ecr_repository.temporal_worker.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Repository for Temporal API
resource "aws_ecr_repository" "temporal_api" {
  name                 = "${var.project_name}/${var.environment}/temporal-api"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-temporal-api"
    Service     = "temporal-api"
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "temporal_api" {
  repository = aws_ecr_repository.temporal_api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Repository for Kafka Consumer
resource "aws_ecr_repository" "kafka_consumer" {
  name                 = "${var.project_name}/${var.environment}/kafka-consumer"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-kafka-consumer"
    Service     = "kafka-consumer"
    Environment = var.environment
  }
}

resource "aws_ecr_lifecycle_policy" "kafka_consumer" {
  repository = aws_ecr_repository.kafka_consumer.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Repository Pull Through Cache for public images
resource "aws_ecr_pull_through_cache_rule" "dockerhub" {
  count = var.enable_ecr_pull_through_cache ? 1 : 0

  ecr_repository_prefix = "dockerhub"
  upstream_registry_url = "registry-1.docker.io"
}

resource "aws_ecr_pull_through_cache_rule" "quay" {
  count = var.enable_ecr_pull_through_cache ? 1 : 0

  ecr_repository_prefix = "quay"
  upstream_registry_url = "quay.io"
}

# IAM Policy for ECR access from EKS
resource "aws_iam_policy" "ecr_read" {
  name        = "${var.project_name}-${var.environment}-ecr-read"
  description = "Allow EKS nodes to pull images from ECR"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_node_ecr_read" {
  role       = module.eks.eks_managed_node_groups["main"].iam_role_name
  policy_arn = aws_iam_policy.ecr_read.arn
}

# CloudWatch Metric Alarm for ECR scan findings
resource "aws_cloudwatch_metric_alarm" "ecr_critical_findings" {
  alarm_name          = "${var.project_name}-${var.environment}-ecr-critical-findings"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "ImageScanFindingsSeverityCritical"
  namespace           = "AWS/ECR"
  period              = "300"
  statistic           = "Maximum"
  threshold           = "0"
  alarm_description   = "Alert when critical vulnerabilities are found in ECR images"
  treat_missing_data  = "notBreaching"

  dimensions = {
    RepositoryName = aws_ecr_repository.api.name
  }
}
