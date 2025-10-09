# House Helper Helm Chart

Helm chart for deploying the House Helper application to Kubernetes.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.10+
- AWS EKS cluster (for production)
- External Secrets Operator (optional, for AWS Secrets Manager integration)
- Prometheus Operator (optional, for monitoring)

## Installing the Chart

### Local Development

For local development with embedded PostgreSQL, Redis, and Kafka:

```bash
helm install house-helper ./house-helper \
  --namespace house-helper \
  --create-namespace \
  --set postgresql.enabled=true \
  --set redis.enabled=true \
  --set kafka.enabled=true \
  --set externalSecrets.enabled=false
```

### Production (AWS EKS)

For production deployment to AWS EKS with external services:

```bash
# Add required Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

# Install External Secrets Operator
helm install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets \
  --create-namespace

# Install House Helper
helm install house-helper ./house-helper \
  --namespace house-helper \
  --create-namespace \
  --values values-prod.yaml
```

## Configuration

The following table lists the configurable parameters of the House Helper chart and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.projectName` | Project name | `house-helper` |
| `global.environment` | Environment (dev/staging/prod) | `dev` |
| `global.domain` | Application domain | `househelper.app` |
| `global.imageRegistry` | Container image registry | `""` |
| `global.imagePullSecrets` | Image pull secrets | `[]` |

### API Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `api.enabled` | Enable API service | `true` |
| `api.replicaCount` | Number of replicas | `2` |
| `api.image.repository` | Image repository | `house-helper/dev/api` |
| `api.image.tag` | Image tag | `latest` |
| `api.service.port` | Service port | `8080` |
| `api.resources.requests.cpu` | CPU request | `250m` |
| `api.resources.requests.memory` | Memory request | `512Mi` |
| `api.autoscaling.enabled` | Enable HPA | `true` |
| `api.autoscaling.minReplicas` | Minimum replicas | `2` |
| `api.autoscaling.maxReplicas` | Maximum replicas | `10` |

### Notifier Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `notifier.enabled` | Enable Notifier service | `true` |
| `notifier.replicaCount` | Number of replicas | `2` |
| `notifier.image.repository` | Image repository | `house-helper/dev/notifier` |
| `notifier.service.port` | Service port | `8081` |
| `notifier.autoscaling.enabled` | Enable HPA | `true` |

### Temporal Services Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `temporalWorker.enabled` | Enable Temporal Worker | `true` |
| `temporalWorker.replicaCount` | Number of replicas | `2` |
| `temporalApi.enabled` | Enable Temporal API | `true` |
| `temporalApi.service.port` | Service port | `8082` |

### Kafka Consumer Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `kafkaConsumer.enabled` | Enable Kafka Consumer | `true` |
| `kafkaConsumer.replicaCount` | Number of replicas | `2` |
| `kafkaConsumer.env.CONSUMER_GROUP` | Consumer group name | `house-helper-consumers` |

### Ingress Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable Ingress | `true` |
| `ingress.className` | Ingress class name | `alb` |
| `ingress.hosts[0].host` | Hostname | `api.househelper.app` |

### External Secrets Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `externalSecrets.enabled` | Enable External Secrets | `true` |
| `externalSecrets.secretStore.name` | Secret store name | `aws-secrets-manager` |

### Database Parameters (Development)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Enable PostgreSQL | `false` |
| `postgresql.auth.username` | Database username | `house_helper` |
| `postgresql.auth.database` | Database name | `house_helper` |

### Redis Parameters (Development)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `redis.enabled` | Enable Redis | `false` |
| `redis.auth.enabled` | Enable authentication | `true` |

### Kafka Parameters (Development)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `kafka.enabled` | Enable Kafka | `false` |

### Monitoring Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitoring.enabled` | Enable monitoring | `true` |
| `monitoring.serviceMonitor.enabled` | Enable ServiceMonitor | `true` |
| `monitoring.prometheusRule.enabled` | Enable PrometheusRule | `true` |

## Deployment Examples

### Deploy to Development

```bash
helm upgrade --install house-helper ./house-helper \
  --namespace house-helper-dev \
  --create-namespace \
  --set environment=dev \
  --set global.domain=dev.househelper.app \
  --set api.image.tag=v1.0.0 \
  --set notifier.image.tag=v1.0.0 \
  --set temporalWorker.image.tag=v1.0.0 \
  --set temporalApi.image.tag=v1.0.0 \
  --set kafkaConsumer.image.tag=v1.0.0
```

### Deploy to Production

```bash
helm upgrade --install house-helper ./house-helper \
  --namespace house-helper-prod \
  --create-namespace \
  --values values-prod.yaml \
  --set api.image.tag=v1.0.0 \
  --set notifier.image.tag=v1.0.0 \
  --set temporalWorker.image.tag=v1.0.0 \
  --set temporalApi.image.tag=v1.0.0 \
  --set kafkaConsumer.image.tag=v1.0.0
```

### Dry Run

```bash
helm install house-helper ./house-helper \
  --namespace house-helper \
  --dry-run \
  --debug
```

## Upgrading

### Upgrade Release

```bash
helm upgrade house-helper ./house-helper \
  --namespace house-helper \
  --reuse-values \
  --set api.image.tag=v1.1.0
```

### Rollback

```bash
helm rollback house-helper 1 \
  --namespace house-helper
```

## Uninstalling

```bash
helm uninstall house-helper \
  --namespace house-helper
```

## Custom Values Files

### values-prod.yaml

Production configuration with increased resources and replicas:

```yaml
environment: prod
global:
  domain: api.househelper.app

api:
  replicaCount: 4
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 2Gi
  autoscaling:
    minReplicas: 4
    maxReplicas: 20

notifier:
  replicaCount: 4
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 2Gi

temporalWorker:
  replicaCount: 4
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi
    limits:
      cpu: 4000m
      memory: 4Gi

postgresql:
  enabled: false
redis:
  enabled: false
kafka:
  enabled: false

externalSecrets:
  enabled: true

ingress:
  annotations:
    alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-1:ACCOUNT:certificate/CERT_ID
```

### values-staging.yaml

Staging configuration with moderate resources:

```yaml
environment: staging
global:
  domain: staging.househelper.app

api:
  replicaCount: 2
  autoscaling:
    minReplicas: 2
    maxReplicas: 10

postgresql:
  enabled: false
redis:
  enabled: false
kafka:
  enabled: false
```

## Monitoring

The chart includes ServiceMonitor and PrometheusRule resources for integration with Prometheus Operator.

### View Metrics

```bash
kubectl port-forward -n house-helper svc/house-helper-api 9090:9090
curl http://localhost:9090/metrics
```

### View Logs

```bash
kubectl logs -n house-helper -l app.kubernetes.io/component=api --tail=100 -f
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n house-helper
kubectl describe pod -n house-helper <pod-name>
```

### Check Services

```bash
kubectl get svc -n house-helper
kubectl get ingress -n house-helper
```

### Check External Secrets

```bash
kubectl get externalsecrets -n house-helper
kubectl describe externalsecret -n house-helper app-secrets
```

### Check HPA

```bash
kubectl get hpa -n house-helper
kubectl describe hpa -n house-helper house-helper-api
```

### Debug Container

```bash
kubectl run -it --rm debug \
  --image=busybox \
  --restart=Never \
  --namespace=house-helper \
  -- sh
```

## License

Copyright © 2024 House Helper. All rights reserved.
