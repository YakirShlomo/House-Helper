# Security Policy

## Supported Versions

We take security seriously and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in House Helper, please report it responsibly. **Do not open a public GitHub issue.**

### Reporting Process

1. **Email**: Send details to `security@house-helper.com` (replace with actual email)
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)
3. **Response Time**: We aim to respond within 48 hours
4. **Resolution**: Critical issues will be patched within 7 days

### Disclosure Policy

- We will confirm receipt of your report within 48 hours
- We will provide an estimated timeline for a fix
- We will notify you when the vulnerability is fixed
- We will credit you in our security advisories (unless you prefer anonymity)

## Security Best Practices

### For Developers

#### 1. Secrets Management

**Never commit secrets to the repository:**

```bash
# Check for secrets before committing
git diff --staged | grep -i "password\|secret\|key\|token"

# Use git-secrets to prevent commits with secrets
git secrets --install
git secrets --register-aws
```

**Use environment variables:**

```go
// ✅ Good
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ Bad
dbPassword := "hardcoded_password_123"
```

**Use AWS Secrets Manager or Kubernetes Secrets:**

```yaml
# External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: house-helper-secrets
spec:
  secretStoreRef:
    name: aws-secrets-manager
  target:
    name: house-helper-secrets
  data:
    - secretKey: db-password
      remoteRef:
        key: house-helper/prod/db
        property: password
```

#### 2. Dependency Management

**Keep dependencies up to date:**

```bash
# Go dependencies
go get -u ./...
go mod tidy

# Flutter dependencies
flutter pub upgrade
```

**Scan for vulnerabilities:**

```bash
# Go vulnerabilities
govulncheck ./...

# Snyk scanning
snyk test --all-projects

# OWASP Dependency Check
dependency-check --project house-helper --scan .
```

#### 3. Code Security

**Input Validation:**

```go
// Validate and sanitize user input
func validateEmail(email string) error {
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    return nil
}

// Use prepared statements for SQL
stmt, err := db.Prepare("SELECT * FROM users WHERE email = $1")
rows, err := stmt.Query(email)
```

**Authentication:**

```go
// Use bcrypt for password hashing
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verify passwords
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(providedPassword))
```

**Authorization:**

```go
// Implement role-based access control (RBAC)
func requireRole(role string) middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r.Context())
            if !user.HasRole(role) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### 4. Secure Communication

**TLS Configuration:**

```go
// Use TLS 1.3 with secure cipher suites
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
    },
}

server := &http.Server{
    Addr:      ":8443",
    TLSConfig: tlsConfig,
}
```

**Certificate Management:**

```bash
# Use AWS Certificate Manager (ACM) for TLS certificates
# Certificates are automatically renewed by ACM

# For local development, use mkcert
mkcert -install
mkcert localhost 127.0.0.1
```

### For Operations

#### 1. Infrastructure Security

**Network Security:**

- VPC isolation with private subnets
- Security groups with least privilege access
- Network ACLs for additional layer
- WAF for application protection

**Database Security:**

```bash
# Enable encryption at rest
aws rds modify-db-instance \
  --db-instance-identifier house-helper-prod \
  --storage-encrypted \
  --apply-immediately

# Enable automated backups with encryption
aws rds modify-db-instance \
  --db-instance-identifier house-helper-prod \
  --backup-retention-period 7 \
  --preferred-backup-window "03:00-04:00"

# Enable SSL/TLS connections
psql "host=mydb.123456789.us-east-1.rds.amazonaws.com port=5432 dbname=househelper user=admin sslmode=require"
```

**Kubernetes Security:**

```yaml
# Pod Security Standards
apiVersion: v1
kind: Namespace
metadata:
  name: house-helper-prod
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted

# Network Policies
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-network-policy
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: house-helper-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: ingress-nginx
  egress:
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: postgresql
    ports:
    - protocol: TCP
      port: 5432
```

#### 2. Access Control

**IAM Best Practices:**

```bash
# Use IAM roles, not access keys
# GitHub Actions uses OIDC for temporary credentials

# Principle of least privilege
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:PutImage"
      ],
      "Resource": "arn:aws:ecr:us-east-1:ACCOUNT_ID:repository/house-helper-*"
    }
  ]
}
```

**Kubernetes RBAC:**

```yaml
# Service Account with limited permissions
apiVersion: v1
kind: ServiceAccount
metadata:
  name: house-helper-api
  namespace: house-helper-prod

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: house-helper-api
  namespace: house-helper-prod
rules:
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: house-helper-api
  namespace: house-helper-prod
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: house-helper-api
subjects:
- kind: ServiceAccount
  name: house-helper-api
  namespace: house-helper-prod
```

#### 3. Monitoring & Logging

**Enable CloudWatch Logs:**

```bash
# Container logs to CloudWatch
aws logs create-log-group --log-group-name /aws/eks/house-helper-prod

# RDS logs
aws rds modify-db-instance \
  --db-instance-identifier house-helper-prod \
  --cloudwatch-logs-export-configuration \
    '{"EnableLogTypes":["error","general","slowquery"]}'
```

**Security Monitoring:**

```yaml
# Prometheus alerts for security events
groups:
- name: security
  rules:
  - alert: HighFailedLoginRate
    expr: rate(failed_login_attempts[5m]) > 10
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High failed login rate detected"
      
  - alert: UnauthorizedAccessAttempt
    expr: rate(unauthorized_access_attempts[5m]) > 5
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Unauthorized access attempts detected"
```

#### 4. Backup & Recovery

**Database Backups:**

```bash
# Automated daily backups
aws rds modify-db-instance \
  --db-instance-identifier house-helper-prod \
  --backup-retention-period 30 \
  --preferred-backup-window "03:00-04:00"

# Manual snapshot
aws rds create-db-snapshot \
  --db-snapshot-identifier house-helper-prod-manual-$(date +%Y%m%d) \
  --db-instance-identifier house-helper-prod

# Point-in-time recovery
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier house-helper-prod \
  --target-db-instance-identifier house-helper-prod-restored \
  --restore-time 2024-01-15T14:00:00Z
```

**Disaster Recovery:**

```bash
# Cross-region replication for S3
aws s3api put-bucket-replication \
  --bucket house-helper-prod-assets \
  --replication-configuration file://replication-config.json

# Multi-region EKS clusters
# Primary: us-east-1
# Secondary: us-west-2
```

## Encryption

### Data at Rest

- **RDS**: AES-256 encryption enabled via AWS KMS
- **S3**: Server-side encryption with AWS managed keys (SSE-S3)
- **EBS**: Encrypted volumes for EKS worker nodes
- **ElastiCache**: At-rest encryption enabled

### Data in Transit

- **API**: TLS 1.3 with strong cipher suites
- **Database**: SSL/TLS enforced for all connections
- **Internal Services**: mTLS between microservices
- **Mobile App**: Certificate pinning implemented

### Key Management

```bash
# AWS KMS for encryption keys
aws kms create-key \
  --description "House Helper production encryption key" \
  --key-policy file://key-policy.json

# Rotate keys annually
aws kms enable-key-rotation --key-id <key-id>

# Audit key usage
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceType,AttributeValue=AWS::KMS::Key
```

## GDPR Compliance

### Data Protection

1. **Data Minimization**: Collect only necessary user data
2. **Purpose Limitation**: Use data only for stated purposes
3. **Storage Limitation**: Retain data only as long as needed
4. **Accuracy**: Ensure data is accurate and up-to-date
5. **Security**: Implement technical and organizational measures

### User Rights

Implement endpoints for:

- **Right to Access**: Users can export their data
- **Right to Rectification**: Users can update their data
- **Right to Erasure**: Users can delete their accounts
- **Right to Data Portability**: Users can download data in JSON format
- **Right to Object**: Users can opt-out of data processing

```go
// Example: User data export
func (s *UserService) ExportUserData(userID string) (*UserDataExport, error) {
    user, err := s.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    tasks, err := s.GetUserTasks(userID)
    if err != nil {
        return nil, err
    }
    
    return &UserDataExport{
        User:      user,
        Tasks:     tasks,
        ExportedAt: time.Now(),
    }, nil
}

// Example: Account deletion
func (s *UserService) DeleteUser(userID string) error {
    // Anonymize user data
    err := s.AnonymizeUser(userID)
    if err != nil {
        return err
    }
    
    // Delete associated data
    err = s.DeleteUserTasks(userID)
    if err != nil {
        return err
    }
    
    // Soft delete user record
    return s.SoftDeleteUser(userID)
}
```

### Data Processing Records

Maintain records of:
- What data is collected
- Why it's collected
- How it's processed
- Who has access
- How long it's retained

### Privacy by Design

- Default to minimal data collection
- Implement data retention policies
- Use pseudonymization where possible
- Regular privacy impact assessments

## Security Hardening Checklist

### Application Level

- [ ] Input validation on all user inputs
- [ ] Output encoding to prevent XSS
- [ ] Parameterized queries to prevent SQL injection
- [ ] CSRF tokens on all state-changing operations
- [ ] Rate limiting on all API endpoints
- [ ] Password complexity requirements enforced
- [ ] Multi-factor authentication available
- [ ] Session timeout after inactivity
- [ ] Secure password reset flow
- [ ] Audit logging for sensitive operations

### Infrastructure Level

- [ ] VPC with private subnets
- [ ] Security groups with minimal access
- [ ] WAF rules for common attacks
- [ ] DDoS protection via CloudFront
- [ ] Encrypted storage volumes
- [ ] Encrypted database instances
- [ ] TLS 1.3 on all public endpoints
- [ ] Certificate rotation automated
- [ ] Secrets stored in Secrets Manager
- [ ] IAM roles follow least privilege

### Container Level

- [ ] Non-root containers
- [ ] Read-only root filesystems
- [ ] Resource limits defined
- [ ] Security context constraints
- [ ] Pod security policies enforced
- [ ] Network policies configured
- [ ] Image scanning before deployment
- [ ] Base images regularly updated
- [ ] No secrets in container images
- [ ] Service mesh for mTLS

### Operational Level

- [ ] Automated security scanning in CI/CD
- [ ] Dependency vulnerability scanning
- [ ] SAST and DAST in pipeline
- [ ] Container image scanning
- [ ] Infrastructure security scanning (tfsec)
- [ ] Regular penetration testing
- [ ] Security training for team
- [ ] Incident response plan documented
- [ ] Disaster recovery plan tested
- [ ] Security metrics tracked

## Incident Response

### Detection

Monitor for:
- Unusual access patterns
- Failed authentication attempts
- Unauthorized access attempts
- Resource exhaustion attacks
- Data exfiltration attempts

### Response Procedure

1. **Identify** the incident
   - Review logs and alerts
   - Determine scope and impact

2. **Contain** the incident
   - Isolate affected systems
   - Revoke compromised credentials
   - Block malicious IPs

3. **Eradicate** the threat
   - Remove malware or unauthorized access
   - Patch vulnerabilities
   - Update security rules

4. **Recover** systems
   - Restore from clean backups
   - Verify system integrity
   - Monitor for recurrence

5. **Learn** from the incident
   - Document timeline and actions
   - Conduct post-mortem
   - Update procedures

### Emergency Contacts

- **Security Team**: security@house-helper.com
- **On-Call Engineer**: Use PagerDuty rotation
- **AWS Support**: Enterprise support contract
- **Legal/Compliance**: legal@house-helper.com

### Incident Severity Levels

- **P0 (Critical)**: Data breach, system compromise
  - Response time: Immediate
  - Escalation: Security team + Management

- **P1 (High)**: Service disruption, attempted breach
  - Response time: < 1 hour
  - Escalation: Security team

- **P2 (Medium)**: Vulnerability identified, no active exploit
  - Response time: < 4 hours
  - Escalation: Development team

- **P3 (Low)**: Minor security issue, no immediate risk
  - Response time: < 24 hours
  - Escalation: Standard process

## Security Contacts

- **Security Team**: security@house-helper.com
- **Privacy Officer**: privacy@house-helper.com
- **Data Protection Officer**: dpo@house-helper.com

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [AWS Security Best Practices](https://aws.amazon.com/security/best-practices/)
- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)

## License

Copyright © 2024 House Helper. All rights reserved.
