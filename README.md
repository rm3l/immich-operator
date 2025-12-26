# Immich Operator

A Kubernetes Operator for deploying and managing [Immich](https://immich.app/) - a high-performance, self-hosted photo and video management solution.

## Description

This operator automates the deployment and lifecycle management of Immich on Kubernetes. It works similarly to the [official Immich Helm chart](https://github.com/immich-app/immich-charts/tree/main/charts/immich), but as a native Kubernetes operator with the following benefits:

- **Declarative Configuration**: Define your Immich deployment as a Kubernetes Custom Resource
- **Automated Reconciliation**: The operator continuously ensures the actual state matches your desired state
- **Simplified Operations**: Single CR to manage all Immich components
- **Native Kubernetes Integration**: Works with kubectl, GitOps tools, and Kubernetes RBAC

### Components Managed

The operator manages the following Immich components:

- **Server**: The main Immich application server (handles web UI, API, and background jobs)
- **Machine Learning**: AI/ML service for smart search, face detection, and duplicate detection
- **Valkey**: Redis-compatible cache for job queues (optional, can use external Redis)

### Prerequisites

External dependencies you need to provide:

- **PostgreSQL Database**: With the `pgvecto.rs` extension installed
- **Persistent Storage**: A PVC for storing photos and videos

## Getting Started

### Prerequisites

- Go version v1.24.0+
- Docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster
- A PostgreSQL database with `pgvecto.rs` extension

### Installation

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the operator to the cluster:**

```sh
make deploy IMG=<some-registry>/immich-operator:tag
```

### Quick Start

1. **Create a namespace for Immich:**

```sh
kubectl create namespace immich
```

2. **Create a secret for the PostgreSQL password:**

```sh
kubectl create secret generic immich-postgres \
  --namespace immich \
  --from-literal=password=your-postgres-password
```

3. **Create a PVC for the photo library:**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: immich-library
  namespace: immich
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
```

4. **Deploy Immich:**

```yaml
apiVersion: media.rm3l.org/v1alpha1
kind: Immich
metadata:
  name: immich
  namespace: immich
spec:
  postgres:
    host: postgres.database.svc.cluster.local
    database: immich
    username: immich
    passwordSecretRef:
      name: immich-postgres
      key: password

  immich:
    persistence:
      library:
        existingClaim: immich-library

  server:
    enabled: true
    image:
      image: ghcr.io/immich-app/immich-server:v1.125.7
    ingress:
      enabled: true
      ingressClassName: nginx
      annotations:
        nginx.ingress.kubernetes.io/proxy-body-size: "0"
      hosts:
        - host: immich.example.com
          paths:
            - path: /
              pathType: Prefix

  machineLearning:
    enabled: true
    image:
      image: ghcr.io/immich-app/immich-machine-learning:v1.125.7
    persistence:
      enabled: true
      size: 10Gi

  valkey:
    enabled: true
    image:
      image: docker.io/valkey/valkey:8-alpine
```

## Configuration Reference

### Image Configuration

Images are resolved in the following order:

1. **CR spec** (highest priority) - if specified in the Custom Resource
2. **Environment variables** (default) - `RELATED_IMAGE_*` set on the operator deployment

The operator comes with default images configured via `RELATED_IMAGE_*` environment variables, so you don't need to specify images in the CR unless you want to override them.

#### Global Image Settings

| Field | Description | Default |
|-------|-------------|---------|
| `imagePullSecrets` | Image pull secrets for private registries | `[]` |

#### Per-Component Image Settings

Each component (server, machineLearning, valkey) supports optional image configuration:

| Field | Description |
|-------|-------------|
| `<component>.image.image` | Full image reference to override the default |
| `<component>.image.pullPolicy` | Pull policy override for this component |

#### Example: Overriding Images

```yaml
spec:
  server:
    image:
      image: my-registry.example.com/immich-server:v1.125.7
  machineLearning:
    image:
      image: ghcr.io/immich-app/immich-machine-learning:v1.125.7
  valkey:
    image:
      image: docker.io/valkey/valkey:8-alpine
```

#### Disconnected/Air-Gapped Environments

For disconnected environments (e.g., OpenShift), the operator supports `RELATED_IMAGE_*` environment variables as a fallback when images are not specified in the CR:

| Environment Variable | Component |
|---------------------|-----------|
| `RELATED_IMAGE_immich` | Immich Server |
| `RELATED_IMAGE_machineLearning` | Machine Learning |
| `RELATED_IMAGE_valkey` | Valkey (Redis) |

Set these in the operator deployment:

```yaml
env:
  - name: RELATED_IMAGE_immich
    value: my-mirror.example.com/immich/server:v1.125.7
  - name: RELATED_IMAGE_machineLearning
    value: my-mirror.example.com/immich/ml:v1.125.7
  - name: RELATED_IMAGE_valkey
    value: my-mirror.example.com/valkey:8-alpine
```

**Note:** If a user specifies an image in the CR, it takes precedence over the environment variable.

### PostgreSQL Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `postgres.host` | PostgreSQL hostname | Required |
| `postgres.port` | PostgreSQL port | `5432` |
| `postgres.database` | Database name | `immich` |
| `postgres.username` | Database username | `immich` |
| `postgres.password` | Database password (plain text) | - |
| `postgres.passwordSecretRef.name` | Secret name containing password | - |
| `postgres.passwordSecretRef.key` | Key in the secret | - |
| `postgres.urlSecretRef.name` | Secret containing full DATABASE_URL | - |

### Immich Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `immich.metrics.enabled` | Enable Prometheus metrics | `false` |
| `immich.persistence.library.existingClaim` | PVC name for photo storage | Required |
| `immich.configuration` | Immich config file (YAML) | `{}` |
| `immich.configurationKind` | ConfigMap or Secret | `ConfigMap` |

### Server Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `server.enabled` | Enable server component | `true` |
| `server.image.image` | Override default image | `RELATED_IMAGE_immich` |
| `server.replicas` | Number of replicas | `1` |
| `server.resources` | Resource requirements | `{}` |
| `server.ingress.enabled` | Enable ingress | `false` |
| `server.ingress.ingressClassName` | Ingress class | - |
| `server.ingress.hosts` | Ingress hosts | `[]` |
| `server.ingress.tls` | Ingress TLS configuration | `[]` |

### Machine Learning Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `machineLearning.enabled` | Enable ML component | `true` |
| `machineLearning.image.image` | Override default image | `RELATED_IMAGE_machineLearning` |
| `machineLearning.replicas` | Number of replicas | `1` |
| `machineLearning.resources` | Resource requirements | `{}` |
| `machineLearning.persistence.enabled` | Enable cache persistence | `true` |
| `machineLearning.persistence.size` | Cache PVC size | `10Gi` |

### Valkey (Redis) Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `valkey.enabled` | Enable built-in Valkey | `true` |
| `valkey.image.image` | Override default image | `RELATED_IMAGE_valkey` |
| `valkey.resources` | Resource requirements | `{}` |
| `valkey.persistence.enabled` | Enable data persistence | `false` |
| `valkey.persistence.size` | Data PVC size | `1Gi` |

### Immich Application Configuration

The `immich.configuration` field accepts Immich's native configuration format. Example:

```yaml
immich:
  configuration:
    trash:
      enabled: true
      days: 30
    storageTemplate:
      enabled: true
      template: "{{y}}/{{y}}-{{MM}}-{{dd}}/{{filename}}"
    ffmpeg:
      crf: 23
      preset: "medium"
      targetVideoCodec: "h264"
    machineLearning:
      enabled: true
      clip:
        enabled: true
        modelName: "ViT-B-32__openai"
      facialRecognition:
        enabled: true
        modelName: "buffalo_l"
```

See the [Immich configuration documentation](https://immich.app/docs/install/config-file/) for all available options.

## Using External Redis

To use an external Redis/Valkey instance instead of the built-in one:

```yaml
spec:
  valkey:
    enabled: false
  server:
    env:
      - name: REDIS_HOSTNAME
        value: redis.external.svc.cluster.local
      - name: REDIS_PORT
        value: "6379"
```

## Status

The operator reports status through the `status` field:

```yaml
status:
  ready: true
  serverReady: true
  machineLearningReady: true
  valkeyReady: true
  version: v1.125.7
  conditions:
    - type: Ready
      status: "True"
      reason: AllComponentsReady
      message: All Immich components are ready
```

View status with:

```sh
kubectl get immich
NAME     READY   VERSION    AGE
immich   true    v1.125.7   5m
```

## Uninstall

**Delete Immich instances:**

```sh
kubectl delete immich --all -A
```

**Delete the CRDs:**

```sh
make uninstall
```

**Undeploy the operator:**

```sh
make undeploy
```

## Development

### Running Locally

```sh
make install      # Install CRDs
make run          # Run controller locally
```

### Running Tests

```sh
make test
```

### Building

```sh
make build                                    # Build binary
make docker-build IMG=myregistry/immich-operator:tag  # Build container
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
