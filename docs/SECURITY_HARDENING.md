# Security Hardening Guide

## Table of Contents

1. [Application Security](#application-security)
2. [Infrastructure Security](#infrastructure-security)
3. [Container Security](#container-security)
4. [Network Security](#network-security)
5. [Database Security](#database-security)
6. [API Security](#api-security)
7. [Mobile App Security](#mobile-app-security)
8. [Secrets Management](#secrets-management)
9. [Monitoring and Logging](#monitoring-and-logging)
10. [Incident Response](#incident-response)

## Application Security

### Input Validation

**Go Services:**

```go
// Validate all user input
import (
    "github.com/go-playground/validator/v10"
)

type CreateTaskRequest struct {
    Title       string    `json:"title" validate:"required,min=1,max=200"`
    Description string    `json:"description" validate:"max=2000"`
    DueDate     time.Time `json:"due_date" validate:"required,gtefield=Now"`
    AssignedTo  string    `json:"assigned_to" validate:"required,uuid4"`
    Points      int       `json:"points" validate:"required,min=1,max=1000"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    var req CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate
    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Sanitize HTML in text fields
    req.Title = html.EscapeString(req.Title)
    req.Description = html.EscapeString(req.Description)
    
    // Process request...
}
```

**Flutter App:**

```dart
// Form validation
class TaskFormValidator {
  static String? validateTitle(String? value) {
    if (value == null || value.isEmpty) {
      return 'Title is required';
    }
    if (value.length > 200) {
      return 'Title must be 200 characters or less';
    }
    return null;
  }
  
  static String? validatePoints(String? value) {
    if (value == null || value.isEmpty) {
      return 'Points are required';
    }
    final points = int.tryParse(value);
    if (points == null || points < 1 || points > 1000) {
      return 'Points must be between 1 and 1000';
    }
    return null;
  }
}

// Use in form
TextFormField(
  controller: _titleController,
  validator: TaskFormValidator.validateTitle,
  decoration: InputDecoration(labelText: 'Title'),
)
```

### SQL Injection Prevention

```go
// ✅ Use parameterized queries
func (r *TaskRepository) GetByID(id string) (*Task, error) {
    var task Task
    err := r.db.Where("id = ?", id).First(&task).Error
    return &task, err
}

// ✅ Use ORM with safe methods
func (r *TaskRepository) Search(query string) ([]*Task, error) {
    var tasks []*Task
    err := r.db.Where("title ILIKE ? OR description ILIKE ?", 
        "%"+query+"%", "%"+query+"%").Find(&tasks).Error
    return tasks, err
}

// ❌ NEVER concatenate user input into SQL
// BAD: r.db.Raw("SELECT * FROM tasks WHERE title = '" + userInput + "'")
```

### XSS Prevention

```go
// Sanitize HTML output
import "html"

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
    task, err := h.service.GetTask(taskID)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    // Escape HTML in response
    response := map[string]interface{}{
        "id":          task.ID,
        "title":       html.EscapeString(task.Title),
        "description": html.EscapeString(task.Description),
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### CSRF Protection

```go
// Use CSRF tokens for state-changing operations
import "github.com/gorilla/csrf"

func main() {
    r := mux.NewRouter()
    
    // CSRF protection middleware
    csrfMiddleware := csrf.Protect(
        []byte("32-byte-long-secret-key-here"),
        csrf.Secure(true), // Only send over HTTPS
        csrf.HttpOnly(true),
        csrf.SameSite(csrf.SameSiteStrictMode),
    )
    
    http.ListenAndServe(":8080", csrfMiddleware(r))
}
```

### Rate Limiting

```go
// Rate limiting middleware
import (
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    limiter, exists := rl.limiters[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[key] = limiter
    }
    
    return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Use IP address as key
        ip := getIPAddress(r)
        limiter := rl.getLimiter(ip)
        
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Usage
rateLimiter := NewRateLimiter(10, 20) // 10 requests per second, burst of 20
r.Use(rateLimiter.Middleware)
```

### Authentication & Authorization

```go
// JWT authentication middleware
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Parse and validate token
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return m.jwtSecret, nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add claims to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "role", claims.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Role-based authorization
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole := r.Context().Value("role").(string)
            if userRole != role && userRole != "admin" {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Infrastructure Security

### VPC Configuration

```hcl
# infra/terraform/vpc.tf
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "house-helper-vpc"
  }
}

# Private subnets for workloads
resource "aws_subnet" "private" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 1}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  tags = {
    Name = "house-helper-private-${count.index + 1}"
    "kubernetes.io/role/internal-elb" = "1"
  }
}

# Public subnets for load balancers only
resource "aws_subnet" "public" {
  count                   = 3
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.${count.index + 101}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  
  tags = {
    Name = "house-helper-public-${count.index + 1}"
    "kubernetes.io/role/elb" = "1"
  }
}
```

### Security Groups

```hcl
# API security group - allow traffic only from ALB
resource "aws_security_group" "api" {
  name        = "house-helper-api"
  description = "Security group for API pods"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
    description     = "Allow traffic from ALB"
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
  
  tags = {
    Name = "house-helper-api"
  }
}

# RDS security group - allow traffic only from EKS nodes
resource "aws_security_group" "rds" {
  name        = "house-helper-rds"
  description = "Security group for RDS instance"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_eks_cluster.main.vpc_config[0].cluster_security_group_id]
    description     = "Allow PostgreSQL from EKS"
  }
  
  tags = {
    Name = "house-helper-rds"
  }
}
```

### IAM Policies (Least Privilege)

```hcl
# EKS node IAM role
resource "aws_iam_role_policy" "node_policy" {
  name = "house-helper-node-policy"
  role = aws_iam_role.node.id
  
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
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ]
        Resource = "${aws_s3_bucket.assets.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = "arn:aws:secretsmanager:${var.region}:${data.aws_caller_identity.current.account_id}:secret:house-helper/*"
      }
    ]
  })
}
```

### WAF Rules

```hcl
# AWS WAF for API protection
resource "aws_wafv2_web_acl" "api" {
  name  = "house-helper-api-waf"
  scope = "REGIONAL"
  
  default_action {
    allow {}
  }
  
  # Rate limiting
  rule {
    name     = "rate-limit"
    priority = 1
    
    action {
      block {}
    }
    
    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }
    
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
  }
  
  # AWS managed rules - Core rule set
  rule {
    name     = "aws-managed-core"
    priority = 2
    
    override_action {
      none {}
    }
    
    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }
    
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedCoreRule"
      sampled_requests_enabled   = true
    }
  }
  
  # SQL injection protection
  rule {
    name     = "sql-injection"
    priority = 3
    
    action {
      block {}
    }
    
    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesSQLiRuleSet"
        vendor_name = "AWS"
      }
    }
    
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "SQLInjectionRule"
      sampled_requests_enabled   = true
    }
  }
  
  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "HouseHelperWAF"
    sampled_requests_enabled   = true
  }
}
```

## Container Security

### Non-Root Containers

```dockerfile
# services/api/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Use distroless base image
FROM gcr.io/distroless/static-debian11

# Run as non-root user
USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /api /api

ENTRYPOINT ["/api"]
```

### Kubernetes Security Context

```yaml
# infra/helm/house-helper/templates/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "house-helper.fullname" . }}-api
spec:
  template:
    spec:
      # Pod-level security context
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        runAsGroup: 65532
        fsGroup: 65532
        seccompProfile:
          type: RuntimeDefault
      
      containers:
      - name: api
        image: {{ .Values.api.image.repository }}:{{ .Values.api.image.tag }}
        
        # Container-level security context
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 65532
          capabilities:
            drop:
            - ALL
        
        # Resource limits
        resources:
          requests:
            cpu: 250m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        
        # Read-only root filesystem requires tmp directory
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      
      volumes:
      - name: tmp
        emptyDir: {}
```

### Image Scanning

```yaml
# .github/workflows/go-services-ci.yml (already created)
# Trivy security scanning is included
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ${{ env.ECR_REGISTRY }}/${{ matrix.service }}:${{ github.sha }}
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'

- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-results.sarif'
```

## Network Security

### Network Policies

```yaml
# infra/helm/house-helper/templates/networkpolicy.yaml (already created)
# Restrict traffic between pods
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
  # Allow traffic from ingress controller
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow database
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: postgresql
    ports:
    - protocol: TCP
      port: 5432
  # Allow Redis
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: redis
    ports:
    - protocol: TCP
      port: 6379
  # Allow Kafka
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: kafka
    ports:
    - protocol: TCP
      port: 9092
```

### mTLS with Service Mesh (Optional)

```yaml
# Using Istio for mTLS
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: house-helper-prod
spec:
  mtls:
    mode: STRICT

---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: api-mtls
  namespace: house-helper-prod
spec:
  host: house-helper-api.house-helper-prod.svc.cluster.local
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
```

## Database Security

### SSL/TLS Connections

```go
// Force SSL connections to RDS
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func ConnectDB(host, port, user, password, dbname string) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
        host, port, user, password, dbname,
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Verify SSL connection
    var sslStatus string
    err = db.QueryRow("SELECT current_setting('ssl')").Scan(&sslStatus)
    if err != nil || sslStatus != "on" {
        return nil, fmt.Errorf("SSL not enabled on database connection")
    }
    
    return db, nil
}
```

### Encryption at Rest

```hcl
# RDS encryption
resource "aws_db_instance" "main" {
  identifier = "house-helper-prod"
  
  # Enable encryption
  storage_encrypted = true
  kms_key_id        = aws_kms_key.rds.arn
  
  # Enable backup encryption
  backup_retention_period = 30
  
  # Enable Performance Insights with encryption
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn
}

# KMS key for RDS
resource "aws_kms_key" "rds" {
  description             = "KMS key for RDS encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  
  tags = {
    Name = "house-helper-rds-key"
  }
}
```

### Database Hardening

```sql
-- Revoke public permissions
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- Create application user with minimal privileges
CREATE USER house_helper_app WITH PASSWORD 'secure_password_here';

-- Grant only necessary permissions
GRANT CONNECT ON DATABASE househelper TO house_helper_app;
GRANT USAGE ON SCHEMA public TO house_helper_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO house_helper_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO house_helper_app;

-- Create read-only user for analytics
CREATE USER house_helper_readonly WITH PASSWORD 'secure_password_here';
GRANT CONNECT ON DATABASE househelper TO house_helper_readonly;
GRANT USAGE ON SCHEMA public TO house_helper_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO house_helper_readonly;

-- Enable row-level security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see their own data
CREATE POLICY user_isolation_policy ON users
    FOR ALL
    TO house_helper_app
    USING (id = current_setting('app.current_user_id')::uuid);

-- Enable audit logging
ALTER SYSTEM SET log_connections = 'on';
ALTER SYSTEM SET log_disconnections = 'on';
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();
```

## API Security

### API Authentication

```go
// API key authentication for service-to-service
func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            http.Error(w, "Missing API key", http.StatusUnauthorized)
            return
        }
        
        // Validate API key (check against database or cache)
        valid, err := m.validateAPIKey(apiKey)
        if err != nil || !valid {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Request Signing

```go
// HMAC request signing for API security
func (m *SignatureMiddleware) Verify(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract signature from header
        signature := r.Header.Get("X-Signature")
        timestamp := r.Header.Get("X-Timestamp")
        
        // Check timestamp to prevent replay attacks
        ts, err := time.Parse(time.RFC3339, timestamp)
        if err != nil || time.Since(ts) > 5*time.Minute {
            http.Error(w, "Invalid or expired timestamp", http.StatusUnauthorized)
            return
        }
        
        // Read and buffer body
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Cannot read body", http.StatusBadRequest)
            return
        }
        r.Body = io.NopCloser(bytes.NewBuffer(body))
        
        // Compute expected signature
        message := fmt.Sprintf("%s%s%s", r.Method, r.URL.Path, timestamp) + string(body)
        mac := hmac.New(sha256.New, []byte(m.secret))
        mac.Write([]byte(message))
        expected := hex.EncodeToString(mac.Sum(nil))
        
        // Compare signatures
        if !hmac.Equal([]byte(signature), []byte(expected)) {
            http.Error(w, "Invalid signature", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### CORS Configuration

```go
// Secure CORS configuration
import "github.com/rs/cors"

func setupCORS() *cors.Cors {
    return cors.New(cors.Options{
        AllowedOrigins: []string{
            "https://app.house-helper.com",
            "https://admin.house-helper.com",
        },
        AllowedMethods: []string{
            http.MethodGet,
            http.MethodPost,
            http.MethodPut,
            http.MethodPatch,
            http.MethodDelete,
        },
        AllowedHeaders: []string{
            "Authorization",
            "Content-Type",
            "X-CSRF-Token",
        },
        ExposedHeaders: []string{
            "X-Total-Count",
            "X-Page-Number",
        },
        AllowCredentials: true,
        MaxAge:           300,
    })
}
```

## Mobile App Security

### Certificate Pinning

```dart
// lib/services/http_service.dart
import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'dart:io';

class HttpService {
  late Dio _dio;
  
  HttpService() {
    _dio = Dio();
    _setupCertificatePinning();
  }
  
  void _setupCertificatePinning() {
    (_dio.httpClientAdapter as DefaultHttpClientAdapter).onHttpClientCreate = 
      (HttpClient client) {
        client.badCertificateCallback = 
          (X509Certificate cert, String host, int port) {
            // Pin specific certificates
            final expectedThumbprint = 
              'YOUR_CERTIFICATE_SHA256_THUMBPRINT_HERE';
            final actualThumbprint = 
              cert.sha256.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
            
            return actualThumbprint == expectedThumbprint;
          };
        return client;
      };
  }
}
```

### Secure Storage

```dart
// lib/services/secure_storage_service.dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorageService {
  final _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(
      encryptedSharedPreferences: true,
    ),
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock,
    ),
  );
  
  Future<void> saveToken(String token) async {
    await _storage.write(key: 'auth_token', value: token);
  }
  
  Future<String?> getToken() async {
    return await _storage.read(key: 'auth_token');
  }
  
  Future<void> deleteToken() async {
    await _storage.delete(key: 'auth_token');
  }
}
```

### Root/Jailbreak Detection

```dart
// lib/services/security_service.dart
import 'package:flutter_jailbreak_detection/flutter_jailbreak_detection.dart';
import 'package:safe_device/safe_device.dart';

class SecurityService {
  Future<bool> isDeviceSecure() async {
    // Check for jailbreak/root
    final isJailbroken = await FlutterJailbreakDetection.jailbroken;
    if (isJailbroken) {
      return false;
    }
    
    // Check for emulator
    final isRealDevice = await SafeDevice.isRealDevice;
    if (!isRealDevice) {
      return false;
    }
    
    // Check for developer mode
    final isDevelopmentMode = await SafeDevice.isDevelopmentModeEnable;
    if (isDevelopmentMode) {
      return false;
    }
    
    return true;
  }
}
```

## Secrets Management

### AWS Secrets Manager

```go
// Retrieve secrets from AWS Secrets Manager
import (
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetSecret(secretName string) (string, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        return "", err
    }
    
    client := secretsmanager.NewFromConfig(cfg)
    
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }
    
    result, err := client.GetSecretValue(context.TODO(), input)
    if err != nil {
        return "", err
    }
    
    return *result.SecretString, nil
}

// Usage
dbPassword, err := GetSecret("house-helper/prod/db-password")
if err != nil {
    log.Fatal(err)
}
```

### External Secrets Operator

```yaml
# infra/helm/house-helper/templates/external-secrets.yaml (already created)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: {{ include "house-helper.fullname" . }}-secrets
spec:
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: {{ include "house-helper.fullname" . }}-secrets
    creationPolicy: Owner
  data:
  - secretKey: db-password
    remoteRef:
      key: house-helper/{{ .Values.environment }}/db
      property: password
  - secretKey: jwt-secret
    remoteRef:
      key: house-helper/{{ .Values.environment }}/jwt
      property: secret
```

### Secret Rotation

```bash
# Rotate database password
aws secretsmanager rotate-secret \
  --secret-id house-helper/prod/db-password \
  --rotation-lambda-arn arn:aws:lambda:us-east-1:ACCOUNT:function:rotate-secret \
  --rotation-rules AutomaticallyAfterDays=30
```

## Monitoring and Logging

### Audit Logging

```go
// Audit log middleware
func AuditLog(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Capture response
        recorder := &responseRecorder{
            ResponseWriter: w,
            statusCode:     http.StatusOK,
        }
        
        next.ServeHTTP(recorder, r)
        
        // Log audit event
        userID := getUserID(r.Context())
        log.WithFields(log.Fields{
            "user_id":     userID,
            "method":      r.Method,
            "path":        r.URL.Path,
            "status_code": recorder.statusCode,
            "duration_ms": time.Since(start).Milliseconds(),
            "ip_address":  getIPAddress(r),
            "user_agent":  r.UserAgent(),
        }).Info("API request")
    })
}
```

### Security Alerts

```yaml
# Prometheus alert rules (already in Helm chart)
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
      description: "More than 10 failed login attempts per second in the last 5 minutes"
  
  - alert: UnauthorizedAccessAttempt
    expr: rate(unauthorized_access_attempts[5m]) > 5
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Unauthorized access attempts detected"
      description: "More than 5 unauthorized access attempts per second"
  
  - alert: AnomalousDataAccess
    expr: rate(data_access_total[5m]) > avg_over_time(data_access_total[1h]) * 3
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "Anomalous data access pattern detected"
      description: "Data access rate is 3x higher than normal"
```

## Incident Response

### Incident Response Playbook

**1. Detection and Analysis:**

```bash
# Check recent security alerts
kubectl get events -n house-helper-prod --sort-by='.lastTimestamp'

# Review CloudWatch Logs for suspicious activity
aws logs filter-log-events \
  --log-group-name /aws/eks/house-helper-prod \
  --filter-pattern "ERROR" \
  --start-time $(date -d '1 hour ago' +%s)000

# Check failed authentication attempts
kubectl logs -n house-helper-prod -l app.kubernetes.io/name=house-helper-api \
  | grep "unauthorized"
```

**2. Containment:**

```bash
# Isolate affected pods
kubectl cordon <node-name>
kubectl drain <node-name> --ignore-daemonsets

# Block malicious IP at WAF level
aws wafv2 update-ip-set \
  --name blocked-ips \
  --scope REGIONAL \
  --id <ip-set-id> \
  --addresses <malicious-ip>/32

# Revoke compromised credentials
aws iam delete-access-key --access-key-id <key-id>
aws secretsmanager rotate-secret --secret-id <secret-id>
```

**3. Eradication and Recovery:**

```bash
# Update to patched image
kubectl set image deployment/house-helper-api \
  api=<ecr-repo>/api:patched-version

# Verify rollout
kubectl rollout status deployment/house-helper-api

# Restore from backup if needed
kubectl exec -n house-helper-prod postgresql-0 -- \
  pg_restore -d househelper /backup/backup-file.dump
```

**4. Post-Incident:**

- Document timeline and actions
- Conduct post-mortem meeting
- Update security procedures
- Implement preventive measures

---

## Security Checklist

Use this checklist for security reviews:

### Application
- [ ] All user input is validated
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)
- [ ] CSRF tokens on state-changing operations
- [ ] Authentication implemented (JWT, OAuth)
- [ ] Authorization checks (RBAC)
- [ ] Rate limiting on all endpoints
- [ ] Session timeout after inactivity
- [ ] Secure password hashing (bcrypt)
- [ ] Audit logging for sensitive operations

### Infrastructure
- [ ] VPC with private subnets
- [ ] Security groups with minimal access
- [ ] WAF rules configured
- [ ] TLS 1.3 on all public endpoints
- [ ] Certificates managed and rotated
- [ ] Secrets in Secrets Manager
- [ ] IAM roles with least privilege
- [ ] Encryption at rest enabled
- [ ] Encryption in transit enforced
- [ ] Multi-AZ deployment for HA

### Container
- [ ] Non-root containers
- [ ] Read-only root filesystem
- [ ] Resource limits defined
- [ ] Security context constraints
- [ ] Network policies configured
- [ ] Image scanning before deployment
- [ ] Base images regularly updated
- [ ] No secrets in images
- [ ] Minimal base images (distroless)

### Monitoring
- [ ] CloudWatch logs enabled
- [ ] Security alerts configured
- [ ] Audit logs retained
- [ ] Metrics dashboards created
- [ ] On-call rotation established
- [ ] Incident response plan documented
- [ ] Regular security reviews scheduled

## License

Copyright © 2024 House Helper. All rights reserved.
