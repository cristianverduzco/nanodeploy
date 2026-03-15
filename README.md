# NanoDeploy

> A self-hosted Kubernetes operator and control plane for provisioning and lifecycle-managing stateful infrastructure services via declarative Custom Resource Definitions (CRDs).

NanoDeploy lets you define managed services like PostgreSQL and Redis the same way you define any Kubernetes resource — with a simple YAML manifest. The operator handles provisioning, lifecycle management, and self-healing automatically.
```yaml
apiVersion: nanodeploy.io/v1alpha1
kind: ManagedService
metadata:
  name: my-postgres
  namespace: default
spec:
  type: postgresql
  version: "15"
  replicas: 1
  storageGB: 5
  databaseName: appdb
```

## Features

- **Kubernetes Operator** — built in Go using `controller-runtime`, watches `ManagedService` CRDs and reconciles desired state continuously
- **Self-Healing** — automatically detects and recovers from infrastructure drift; deleted Deployments and Services are restored within seconds
- **REST API** — Gin-based control plane with CRUD endpoints for managing services programmatically
- **React Dashboard** — real-time web UI displaying service phase, endpoint, and replica state with auto-refresh
- **Prometheus Metrics** — reconcile duration histograms, error counters, and service state gauges exposed at `:8080/metrics`
- **Helm Chart** — production-ready chart with RBAC manifests and a `ServiceMonitor` for Prometheus Operator auto-discovery
- **GitOps Ready** — fully declarative, version-controllable, installable via `helm install`

## Architecture
```
User applies ManagedService CRD
        ↓
Operator reconciliation loop fires
        ↓
Provisions Deployment + Service on Kubernetes
        ↓
Monitors and self-heals every 30 seconds
        ↓
Status exposed via REST API + React dashboard
```

## Stack

| Layer | Technology |
|---|---|
| Operator | Go, controller-runtime |
| API | Go, Gin |
| Dashboard | React, Tailwind CSS |
| Observability | Prometheus |
| Packaging | Helm, Docker (multi-stage) |
| Infrastructure | Kubernetes (kubeadm), Arch Linux |

## Supported Services

| Service | Status |
|---|---|
| PostgreSQL | ✅ Supported |
| Redis | 🚧 Coming soon |
| RabbitMQ | 🚧 Coming soon |

## Installation
```bash
# Apply the CRD
kubectl apply -f config/crd/managedservice.yaml

# Install via Helm
helm install nanodeploy ./charts/nanodeploy

# Deploy a managed service
kubectl apply -f config/deploy/sample-postgres.yaml
```

## Status

🚧 Under active development — core operator, REST API, dashboard, and Helm chart complete.