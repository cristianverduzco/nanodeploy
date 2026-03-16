# NanoDeploy

> A self-hosted Kubernetes operator and control plane for provisioning and lifecycle-managing stateful infrastructure services via declarative Custom Resource Definitions (CRDs).

NanoDeploy lets you define managed services like PostgreSQL and Redis the same way you define any Kubernetes resource — with a simple YAML manifest. The operator handles provisioning, lifecycle management, and self-healing automatically. Think of it as a self-hosted AWS RDS or ElastiCache, running entirely on your own Kubernetes cluster.

Built from scratch to demonstrate deep ownership of the Kubernetes operator pattern, control plane design, and production observability.

---

## How It Works
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

Apply that manifest and NanoDeploy does the rest:
```
User applies ManagedService CRD
        ↓
Operator reconciliation loop fires
        ↓
Provisions Deployment + Service on Kubernetes
        ↓
Status updated: phase=Ready, endpoint=my-postgres.default.svc.cluster.local
        ↓
Self-heals every 30 seconds — drift is corrected automatically
        ↓
State exposed via REST API + React dashboard
```

---

## Features

- **Kubernetes Operator** — built in Go using `controller-runtime`, watches `ManagedService` CRDs and reconciles desired state continuously
- **Self-Healing** — automatically detects and recovers from infrastructure drift; deleted Deployments and Services are restored within seconds
- **REST API** — Gin-based control plane with CRUD endpoints for managing services programmatically
- **React Dashboard** — real-time web UI displaying service phase, endpoint, and replica state with auto-refresh
- **Prometheus Metrics** — reconcile duration histograms, error counters, and service state gauges exposed at `:8080/metrics`
- **Helm Chart** — production-ready chart with RBAC manifests and a `ServiceMonitor` for Prometheus Operator auto-discovery
- **GitOps Ready** — fully declarative, version-controllable, installable via `helm install`
- **Distroless Container** — minimal attack surface, runs as non-root, built with multi-stage Dockerfile

---

## Architecture
```
┌──────────────────────────────────────────────────────────────┐
│                        NanoDeploy                            │
│                                                              │
│  ┌─────────────┐   ┌──────────────┐   ┌──────────────────┐ │
│  │ ManagedSvc  │   │  Controller  │   │   REST API       │ │
│  │    CRD      │──▶│  Reconciler  │   │   (Gin)          │ │
│  └─────────────┘   └──────────────┘   └──────────────────┘ │
│                           │                    │             │
│                           ▼                    ▼             │
│                    ┌──────────────┐   ┌──────────────────┐ │
│                    │  Kubernetes  │   │  React Dashboard  │ │
│                    │     API      │   │  + Prometheus     │ │
│                    └──────────────┘   └──────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

---

## Demo

**Deploy a PostgreSQL instance:**
```bash
kubectl apply -f config/deploy/sample-postgres.yaml
kubectl get managedservice -w
```
```
NAME          TYPE         VERSION   PHASE         ENDPOINT                                AGE
my-postgres   postgresql   15        Provisioning  —                                       2s
my-postgres   postgresql   15        Ready         my-postgres.default.svc.cluster.local   8s
```

**Self-healing in action:**
```bash
kubectl delete deployment my-postgres
# Operator detects drift within 30 seconds and restores it automatically
kubectl get deployment my-postgres
# NAME          READY   UP-TO-DATE   AVAILABLE   AGE
# my-postgres   1/1     1            1           12s
```

---

## Prometheus Metrics

Exposed at `:8080/metrics`:

| Metric | Type | Description |
|---|---|---|
| `nanodeploy_managed_services_total` | Gauge | Total ManagedServices by type and phase |
| `nanodeploy_reconcile_duration_seconds` | Histogram | Duration of each reconcile loop |
| `nanodeploy_reconcile_errors_total` | Counter | Total reconcile errors by service type |

---

## Installation

### Prerequisites

- Kubernetes cluster (kubeadm, EKS, GKE, etc.)
- `kubectl` configured
- `helm` installed

### Deploy via Helm
```bash
# Clone the repo
git clone https://github.com/cristianverduzco/nanodeploy
cd nanodeploy

# Apply the CRD
kubectl apply -f config/crd/managedservice.yaml

# Install the operator
helm install nanodeploy ./charts/nanodeploy

# Deploy your first managed service
kubectl apply -f config/deploy/sample-postgres.yaml

# Watch it provision
kubectl get managedservice -w
```

### Verify the operator is running
```bash
kubectl get pods -n nanodeploy-system
kubectl logs -n nanodeploy-system deployment/nanodeploy-operator
```

---

## CLI / API

The REST API is exposed at `:9090`:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/services` | List all ManagedServices |
| `GET` | `/api/v1/services/:namespace/:name` | Get a specific ManagedService |
| `POST` | `/api/v1/services` | Create a new ManagedService |
| `DELETE` | `/api/v1/services/:namespace/:name` | Delete a ManagedService |

---

## Supported Services

| Service | Status |
|---|---|
| PostgreSQL | ✅ Supported |
| Redis | 🚧 Coming soon |
| RabbitMQ | 🚧 Coming soon |

---

## Stack

| Layer | Technology |
|---|---|
| Operator | Go, controller-runtime |
| API | Go, Gin |
| Dashboard | React, Tailwind CSS |
| Observability | Prometheus |
| Packaging | Helm, distroless Docker (multi-stage) |
| Infrastructure | Kubernetes (kubeadm), Arch Linux |

---

## Roadmap

- [x] ManagedService CRD with spec/status reconciliation
- [x] Self-healing operator with 30-second requeue
- [x] REST API control plane
- [x] React + Tailwind dashboard
- [x] Prometheus metrics instrumentation
- [x] Helm chart with RBAC and ServiceMonitor
- [x] Deployed on self-hosted kubeadm cluster
- [ ] PersistentVolumeClaim provisioning for data durability
- [ ] Secret management for database credentials
- [ ] Redis provisioner implementation
- [ ] Webhook validation for ManagedService specs
- [ ] Multi-namespace isolation per tenant

---

## Status

✅ Core operator, REST API, dashboard, Prometheus metrics, and Helm chart complete — running on a self-hosted kubeadm cluster (Arch Linux, Kubernetes v1.35).