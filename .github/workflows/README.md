# GitHub Actions CI/CD Documentation

This directory contains GitHub Actions workflows for continuous integration and deployment of the House Helper application.

## Workflows Overview

### 1. Flutter Mobile App CI/CD (`flutter-ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`
- Manual dispatch

**Jobs:**
- **Analyze**: Code formatting, static analysis, custom lints
- **Test**: Unit and widget tests with coverage reporting
- **Build Android**: APK and App Bundle generation with signing
- **Build iOS**: IPA generation (requires code signing setup)
- **Deploy Android**: Upload to Google Play Internal Testing

**Required Secrets:**
- `ANDROID_KEYSTORE_BASE64`: Base64-encoded Android keystore
- `ANDROID_KEYSTORE_PASSWORD`: Keystore password
- `ANDROID_KEY_ALIAS`: Key alias
- `ANDROID_KEY_PASSWORD`: Key password
- `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON`: Google Play service account JSON

### 2. Go Services CI/CD (`go-services-ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests
- Manual dispatch

**Jobs:**
- **Test**: Run tests with race detection and coverage for all services
- **Lint**: golangci-lint, staticcheck, go vet
- **Build**: Build and push Docker images to Amazon ECR
- **Deploy Dev**: Deploy to development environment
- **Deploy Prod**: Deploy to production environment with verification

**Required Secrets:**
- `AWS_ROLE_ARN`: AWS IAM role ARN for OIDC authentication

**Services:**
- API
- Notifier
- Temporal Worker
- Temporal API
- Kafka Consumer

### 3. Infrastructure CI/CD (`infrastructure-ci.yml`)

**Triggers:**
- Push to `main` branch (Terraform changes)
- Pull requests (Terraform changes)
- Manual dispatch with action selection

**Jobs:**
- **Validate**: Format check, init, validate, tfsec security scan
- **Plan**: Generate Terraform plan and comment on PR
- **Apply**: Apply changes to production infrastructure
- **Destroy**: Destroy infrastructure (manual trigger only)

**Required Secrets:**
- `AWS_ROLE_ARN`: AWS IAM role ARN for OIDC authentication

### 4. Security Scanning (`security-scanning.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests
- Daily schedule (2 AM UTC)
- Manual dispatch

**Jobs:**
- **Dependency Review**: Check for vulnerable dependencies in PRs
- **Snyk Security**: Vulnerability scanning with Snyk
- **CodeQL Analysis**: Static analysis for Go and JavaScript
- **Secret Scanning**: Detect exposed secrets with TruffleHog
- **Terraform Security**: tfsec security scanning
- **Docker Security**: Trivy and Grype image scanning
- **OSV Scanner**: Open Source Vulnerabilities scanning
- **License Compliance**: Check open source license compliance
- **SBOM Generation**: Generate Software Bill of Materials

**Required Secrets:**
- `SNYK_TOKEN`: Snyk API token

### 5. Database Migrations (`database-migrations.yml`)

**Triggers:**
- Push to `main` or `develop` branches (migration changes)
- Pull requests (migration changes)
- Manual dispatch with environment selection

**Jobs:**
- **Validate Migrations**: Test migrations with PostgreSQL container
- **Run Migrations Dev**: Apply to development database
- **Run Migrations Prod**: Apply to production with backup

**Required Secrets:**
- `AWS_ROLE_ARN`: AWS IAM role ARN for OIDC authentication

## Setup Instructions

### 1. AWS OIDC Configuration

Set up OIDC provider in AWS for GitHub Actions:

```bash
# Create OIDC provider (already done in Terraform)
# Terraform: infra/terraform/iam.tf creates GitHub Actions OIDC provider

# Get the Role ARN
terraform output github_actions_role_arn
```

Add the role ARN to GitHub secrets:
- Repository Settings > Secrets and variables > Actions
- Add secret: `AWS_ROLE_ARN`

### 2. Android Build Secrets

Generate keystore:

```bash
keytool -genkey -v -keystore house-helper.jks \
  -keyalg RSA -keysize 2048 -validity 10000 \
  -alias house-helper

# Convert to base64
base64 house-helper.jks > keystore.base64
```

Add to GitHub secrets:
- `ANDROID_KEYSTORE_BASE64`: Contents of keystore.base64
- `ANDROID_KEYSTORE_PASSWORD`: Keystore password
- `ANDROID_KEY_ALIAS`: Key alias (house-helper)
- `ANDROID_KEY_PASSWORD`: Key password

### 3. Google Play Deployment

Create service account in Google Play Console:
1. Google Play Console > Setup > API access
2. Create new service account
3. Grant "Release to production, exclude devices, and use Play App Signing" permission
4. Download JSON key

Add to GitHub secrets:
- `GOOGLE_PLAY_SERVICE_ACCOUNT_JSON`: Contents of JSON key file

### 4. Snyk Security Scanning

Sign up for Snyk and get API token:
1. Sign up at https://snyk.io/
2. Account Settings > API Token
3. Copy token

Add to GitHub secrets:
- `SNYK_TOKEN`: Snyk API token

### 5. Environment Protection Rules

Configure environment protection in GitHub:
1. Repository Settings > Environments
2. Create environments:
   - `development`
   - `production`
   - `production-infrastructure`
   - `production-infrastructure-destroy`

3. For production environments:
   - Enable "Required reviewers"
   - Add reviewers
   - Enable "Wait timer" (optional)

## Workflow Patterns

### Branch Strategy

- `develop`: Development branch, deploys to dev environment
- `main`: Production branch, deploys to production
- Feature branches: Run tests and security scans only

### Deployment Flow

```
Feature Branch → PR → develop → Deploy to Dev → main → Deploy to Production
                 ↓                                ↓
              Tests + Scans                   Tests + Scans + Approval
```

### Versioning

Docker images are tagged with:
- Branch name
- Git SHA
- Semver (if tagged)
- `latest` for default branch

## Monitoring Workflows

### View Workflow Runs

```bash
# Using GitHub CLI
gh run list --limit 20

# View specific run
gh run view <run-id>

# View logs
gh run view <run-id> --log
```

### Workflow Status Badges

Add to README.md:

```markdown
![Flutter CI](https://github.com/YakirShlomo/House-Helper/workflows/Flutter%20Mobile%20App%20CI%2FCD/badge.svg)
![Go Services CI](https://github.com/YakirShlomo/House-Helper/workflows/Go%20Services%20CI%2FCD/badge.svg)
![Security](https://github.com/YakirShlomo/House-Helper/workflows/Security%20Scanning/badge.svg)
```

## Troubleshooting

### Failed AWS Authentication

Check OIDC configuration:
```bash
aws iam get-open-id-connect-provider \
  --open-id-connect-provider-arn <provider-arn>
```

### Failed Docker Push

Verify ECR permissions:
```bash
aws ecr describe-repositories
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account>.dkr.ecr.us-east-1.amazonaws.com
```

### Failed Kubernetes Deployment

Check EKS access:
```bash
aws eks update-kubeconfig --name house-helper-prod
kubectl get nodes
kubectl get pods -n house-helper-prod
```

### Failed Database Migration

Test migration locally:
```bash
migrate -path services/api/migrations \
  -database "postgres://user:pass@host:5432/db?sslmode=require" \
  up
```

## Best Practices

1. **Secrets Management**
   - Never commit secrets to repository
   - Use GitHub Secrets for sensitive data
   - Rotate credentials regularly

2. **Security Scanning**
   - Review security scan results
   - Fix critical vulnerabilities before merging
   - Keep dependencies up to date

3. **Testing**
   - Maintain high test coverage (>80%)
   - Run tests locally before pushing
   - Fix failing tests immediately

4. **Deployment**
   - Always test in development first
   - Review deployment plans
   - Monitor applications after deployment

5. **Infrastructure**
   - Review Terraform plans carefully
   - Use separate workspaces for environments
   - Backup critical data before changes

## Manual Workflow Dispatch

Run workflows manually:

```bash
# Using GitHub CLI
gh workflow run flutter-ci.yml

gh workflow run infrastructure-ci.yml \
  -f action=plan

gh workflow run database-migrations.yml \
  -f environment=development
```

## Cost Optimization

- Use caching for dependencies
- Limit concurrent workflows
- Use workflow conditions to skip unnecessary jobs
- Clean up old artifacts regularly

## Support

For issues with workflows:
1. Check workflow logs in GitHub Actions tab
2. Review security scan results
3. Verify secrets configuration
4. Test locally when possible

## License

Copyright © 2024 House Helper. All rights reserved.
