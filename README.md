# Immich Operator

A Kubernetes Operator for deploying and managing [Immich](https://immich.app/) (a high-performance, self-hosted photo and video management solution).

## Description

This operator automates the deployment and lifecycle management of Immich on Kubernetes. It provides an alternative method to the [official Immich Helm chart](https://github.com/immich-app/immich-charts/tree/main/charts/immich) for deploying and operating Immich, but as a native Kubernetes operator with the following benefits:

- **Declarative Configuration**: Define your Immich deployment as a Kubernetes Custom Resource (CR)
- **Automated Reconciliation**: The operator continuously ensures the actual state matches your desired state
- **Simplified Operations**: Single CR to manage all Immich components
- **Native Kubernetes Integration**: Works with kubectl, GitOps tools, and Kubernetes RBAC

### Components Managed

The operator manages the following Immich components:

- **Server**: The main Immich application server (handles web UI, API, and background jobs)
- **Machine Learning**: AI/ML service for smart search, face detection, and duplicate detection (built-in by default, can use external)
- **PostgreSQL**: Database with pgvecto.rs extension (built-in by default, can use external)
- **Valkey**: Redis-compatible cache for job queues (built-in by default, can use external)

## Getting Started

### Prerequisites

- Go version v1.25.0+
- Docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster
  - **Persistent Storage**: A StorageClass for creating PVCs (library, database, ML cache)
- If using an external PostgreSAL database, make sure it has the `pgvecto.rs` extension installed and enabled

> **Note**: PostgreSQL and Valkey are deployed by the operator by default. You can disable them and use external instances if preferred.

### Installation

**Deploy the operator image to the cluster:**

```sh
make deploy
```

### Quick Start

1. **Create a namespace for Immich:**

```sh
kubectl create namespace my-ns
```

2. **Deploy Immich (minimal):**

```yaml
apiVersion: media.rm3l.org/v1alpha1
kind: Immich
metadata:
  name: my-immich
  namespace: my-ns
# No spec required! Sensible defaults are provided.
```

That's it! The operator will automatically:

- Create a default library PVC (override with `spec.immich.persistence.library.size`)
- Deploy PostgreSQL with auto-generated credentials (stored in `<name>-postgres-credentials` secret)
- Deploy Valkey for job queues
- Deploy the Immich server and machine learning components

**With custom configuration:**

```yaml
apiVersion: media.rm3l.org/v1alpha1
kind: Immich
metadata:
  name: my-immich
  namespace: my-ns
spec:
  # Customize library storage size
  immich:
    persistence:
      library:
        size: 100Gi

  # Optional: Configure ingress
  server:
    ingress:
      enabled: true
      ingressClassName: nginx
      annotations:
        nginx.ingress.kubernetes.io/proxy-body-size: "0"
      hosts:
        - host: my-immich.example.com
          paths:
            - path: /
              pathType: Prefix
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
| `<component>.image` | Full image reference to override the default |
| `<component>.imagePullPolicy` | Pull policy override for this component |

#### Example: Overriding Images

```yaml
spec:
  server:
    image: my-registry.example.com/immich-server:v1.125.7
    imagePullPolicy: Always
  machineLearning:
    image: my-registry.example.com/immich-machine-learning:v1.125.7
  valkey:
    image: my-registry.example.com/valkey:9-alpine
```

#### Disconnected/Air-Gapped Environments

For disconnected environments (e.g., OpenShift), the operator supports `RELATED_IMAGE_*` environment variables as a fallback when images are not specified in the CR:

| Environment Variable | Component |
|---------------------|-----------|
| `RELATED_IMAGE_immich` | Immich Server |
| `RELATED_IMAGE_machineLearning` | Machine Learning |
| `RELATED_IMAGE_valkey` | Valkey (Redis) |
| `RELATED_IMAGE_postgres` | PostgreSQL |

Set these in the operator deployment:

```yaml
env:
  - name: RELATED_IMAGE_immich
    value: my-mirror.example.com/immich/server:v2.4.1
  - name: RELATED_IMAGE_machineLearning
    value: my-mirror.example.com/immich/ml:v2.4.1
  - name: RELATED_IMAGE_valkey
    value: my-mirror.example.com/valkey:9-alpine
  - name: RELATED_IMAGE_postgres
    value: my-mirror.example.com/immich/postgres:14-vectorchord0.4.3-pgvectors0.2.0
```

**Note:** If a user specifies an image in the CR, it takes precedence over the environment variable.

### PostgreSQL Configuration

The operator deploys PostgreSQL by default with auto-generated credentials. Set `postgres.enabled: false` to use an external database.

When using built-in PostgreSQL, credentials are automatically generated and stored in a secret named `<immich-name>-postgres-credentials`.

> **Data Safety:** PostgreSQL uses a StatefulSet with VolumeClaimTemplates, meaning the PVC (`data-<immich-name>-postgres-0`) will **persist** when the Immich CR is deleted. This protects your database from accidental deletion. When you recreate the CR with the same name, the existing PVC will be reused automatically. To delete the PVC, you must do so manually.

| Field | Description | Default |
|-------|-------------|---------|
| `postgres.enabled` | Deploy built-in PostgreSQL | `true` |
| `postgres.image` | Override default image | `RELATED_IMAGE_postgres` |
| `postgres.imagePullPolicy` | Pull policy for this component | (K8s default) |
| `postgres.resources` | Resource requirements | `{}` |
| `postgres.persistence.size` | Data PVC size | `10Gi` |
| `postgres.persistence.storageClass` | Storage class | (default) |
| `postgres.persistence.existingClaim` | Use existing PVC | - |
| `postgres.passwordSecretRef.name` | Secret name containing password | (auto-generated) |
| `postgres.passwordSecretRef.key` | Key in the secret | - |

**External PostgreSQL** (when `postgres.enabled: false`):

| Field | Description | Default |
|-------|-------------|---------|
| `postgres.host` | PostgreSQL hostname | Required |
| `postgres.port` | PostgreSQL port | `5432` |
| `postgres.database` | Database name | `immich` |
| `postgres.username` | Database username | `immich` |
| `postgres.urlSecretRef.name` | Secret containing full DATABASE_URL | - |

### Immich Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `immich.metrics.enabled` | Enable Prometheus metrics | `false` |
| `immich.persistence.library.existingClaim` | Use an existing PVC for photo storage | - |
| `immich.persistence.library.size` | Size of PVC to create (if existingClaim not set) | `10Gi` |
| `immich.persistence.library.storageClass` | Storage class for managed PVC | (default) |
| `immich.persistence.library.accessModes` | Access modes for managed PVC | `["ReadWriteOnce"]` |
| `immich.configuration` | Immich config file (YAML) | `{}` |
| `immich.configurationKind` | ConfigMap or Secret | `ConfigMap` |

### Server Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `server.enabled` | Enable server component | `true` |
| `server.image` | Override default image | `RELATED_IMAGE_immich` |
| `server.imagePullPolicy` | Pull policy for this component | (K8s default) |
| `server.replicas` | Number of replicas | `1` |
| `server.resources` | Resource requirements | `{}` |
| `server.ingress.enabled` | Enable ingress | `false` |
| `server.ingress.ingressClassName` | Ingress class | - |
| `server.ingress.hosts` | Ingress hosts | `[]` |
| `server.ingress.tls` | Ingress TLS configuration | `[]` |

### Machine Learning Configuration

The operator deploys the ML component by default. Set `machineLearning.enabled: false` to disable it or use an external ML service.

> **Note:** Machine Learning is optional. When disabled without an external URL, Immich runs without ML features (smart search, face detection, duplicate detection, etc.).

| Field | Description | Default |
|-------|-------------|---------|
| `machineLearning.enabled` | Deploy built-in ML component | `true` |
| `machineLearning.image` | Override default image | `RELATED_IMAGE_machineLearning` |
| `machineLearning.imagePullPolicy` | Pull policy for this component | (K8s default) |
| `machineLearning.replicas` | Number of replicas | `1` |
| `machineLearning.resources` | Resource requirements | `{}` |
| `machineLearning.persistence.enabled` | Enable cache persistence | `true` |
| `machineLearning.persistence.size` | Cache PVC size | `10Gi` |

**External ML Service** (when `machineLearning.enabled: false`):

| Field | Description | Default |
|-------|-------------|---------|
| `machineLearning.url` | URL of external ML service | - |

### Valkey (Redis) Configuration

The operator deploys Valkey by default. Set `valkey.enabled: false` to use an external Redis.

| Field | Description | Default |
|-------|-------------|---------|
| `valkey.enabled` | Deploy built-in Valkey | `true` |
| `valkey.image` | Override default image | `RELATED_IMAGE_valkey` |
| `valkey.imagePullPolicy` | Pull policy for this component | (K8s default) |
| `valkey.resources` | Resource requirements | `{}` |
| `valkey.persistence.enabled` | Enable data persistence | `false` |
| `valkey.persistence.size` | Data PVC size | `10Gi` |

**External Redis/Valkey** (when `valkey.enabled: false`):

| Field | Description | Default |
|-------|-------------|---------|
| `valkey.host` | Redis hostname | Required |
| `valkey.port` | Redis port | `6379` |
| `valkey.password` | Redis password (plain text) | - |
| `valkey.passwordSecretRef` | Secret containing password | - |
| `valkey.dbIndex` | Redis database index (0-15) | `0` |

### Immich Application Configuration

The operator automatically generates an Immich configuration file based on CR settings. This base configuration is merged with any user-provided settings in `immich.configuration`.

**Auto-generated settings:**

| Setting | Derived From | Description |
|---------|--------------|-------------|
| `machineLearning.enabled` | `spec.machineLearning.enabled` | Automatically set to `false` when ML is disabled |
| `machineLearning.url` | ML component state | Set to internal service URL or external URL |

**User configuration takes precedence** - any settings you provide in `immich.configuration` will override the auto-generated values.

Example with custom configuration:

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
      # User settings override auto-generated values
      clip:
        enabled: true
        modelName: "ViT-B-32__openai"
      facialRecognition:
        enabled: true
        modelName: "buffalo_l"
```

See the [Immich configuration documentation](https://immich.app/docs/install/config-file/) for all available options.

### Library Persistence

The photo library requires persistent storage. By default, the operator creates a default PVC for the library.

**Option 1: Operator-managed PVC** (default)

Let the operator create and manage the PVC (has some defaults if not specified):

```yaml
spec:
  immich:
    persistence:
      library:
        size: 100Gi
        storageClass: my-storage-class  # optional, uses default if not set
        accessModes: ["ReadWriteOnce"]  # optional, defaults to ReadWriteOnce
```

The operator creates a PVC named `<immich-name>-library`.

> **Data Safety:** The library PVC does **NOT** have an owner reference, meaning it will **persist** when the Immich CR is deleted. This protects your photos from accidental deletion. When you recreate the CR with the same name, the existing PVC will be reused automatically. To delete the PVC, you must do so manually.

**Option 2: Existing PVC**

Use a pre-existing PVC that you manage yourself:

```yaml
spec:
  immich:
    persistence:
      library:
        existingClaim: my-existing-pvc
```

This is useful when you want the PVC to persist beyond the lifecycle of the Immich CR, or when you have specific storage requirements.

### Multi-Node Cluster Considerations

By default, all PVCs use `ReadWriteOnce` access mode. Here's what this means for different storage types:

**Network-attached storage** (NFS, Ceph, Longhorn, etc.):

- Works across nodes since storage is accessible cluster-wide
- Each component has its own PVC, so `ReadWriteOnce` is sufficient for single-replica deployments

**Local storage** (local-path, hostPath, etc.):

- PVCs are bound to specific nodes
- All Immich pods must be scheduled on the same node
- Consider using node affinity or a different storage class

**High Availability (multiple server replicas)**:

If you want to run multiple replicas of the server component, the library PVC needs `ReadWriteMany`:

```yaml
spec:
  immich:
    persistence:
      library:
        size: 100Gi
        accessModes:
          - ReadWriteMany
  server:
    replicas: 3
```

> **Note:** `ReadWriteMany` requires a storage class that supports it (e.g., NFS, CephFS, Azure Files).

| Component | PVC | Notes |
|-----------|-----|-------|
| Server | Library PVC | Needs `ReadWriteMany` for multiple replicas |
| Machine Learning | ML Cache PVC | Separate per pod, `ReadWriteOnce` is fine |
| PostgreSQL | Postgres Data PVC | StatefulSet, `ReadWriteOnce` is fine |
| Valkey | Valkey Data PVC | StatefulSet, `ReadWriteOnce` is fine |

## Using External Services

### External PostgreSQL

To use an external PostgreSQL database instead of the built-in one:

```yaml
spec:
  postgres:
    enabled: false
    host: postgres.external.svc.cluster.local
    port: 5432
    database: immich
    username: immich
    passwordSecretRef:
      name: immich-postgres
      key: password
```

> **Note:** The external PostgreSQL must have the `pgvecto.rs` extension installed. You can use the [official Immich PostgreSQL image](https://github.com/immich-app/immich/pkgs/container/postgres) or install the extension on your existing database.

### External Redis/Valkey

To use an external Redis/Valkey instance instead of the built-in one:

```yaml
spec:
  valkey:
    enabled: false
    host: redis.external.svc.cluster.local
    port: 6379
    passwordSecretRef: # optional
      name: immich-redis
      key: password
    dbIndex: 0  # optional
```

### Disabling or Externalizing Machine Learning

Machine Learning is optional in Immich. You can disable it completely or use an external service.

**Disable ML completely** (no smart search, face detection, etc.):

```yaml
spec:
  machineLearning:
    enabled: false
```

**Use an external ML service:**

```yaml
spec:
  machineLearning:
    enabled: false
    url: http://external-ml-service:3003
```

External ML is useful for:

- Running ML inference on dedicated hardware (e.g., GPU nodes)
- Sharing a single ML service across multiple Immich instances
- Using a custom ML service implementation

## Status

The operator reports status through the `status` field:

```yaml
status:
  ready: true
  serverReady: true
  machineLearningReady: true
  valkeyReady: true
  postgresReady: true
  conditions:
    - type: Ready
      status: "True"
      reason: AllComponentsReady
      message: All Immich components are ready
```

View status with:

```sh
kubectl get immich
NAME     READY   AGE
immich   true    5m
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

Contributions are welcome! Please feel free to file an issue or submit a pull request.

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
