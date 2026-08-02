# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
make manifests generate     # Regenerate CRD/RBAC YAML and DeepCopy code after changing api/v1alpha1 types
make test                   # Run unit + envtest suite (go test ./... except test/e2e), writes cover.out
make lint                   # golangci-lint run
make lint-fix                # golangci-lint run --fix
make validate                 # lint-operatorhub + lint (what CI's lint.yml runs)
make test-e2e                 # Spins up a Kind cluster, runs test/e2e, tears it down
make bundle build-installer   # Regenerate bundle/ (OLM) and dist/install.yaml from config/*
```

Run a single test:

```sh
go test ./internal/controller/... -run TestConfigSpecToMapPreservesCase   # plain go test
go test ./internal/controller/... -run TestControllers -args -ginkgo.focus "some spec text"  # ginkgo suite (suite_test.go/immich_controller_test.go)
```

`make test` requires envtest binaries; it invokes `setup-envtest` itself via `KUBEBUILDER_ASSETS`, so prefer it (or at least `make manifests generate` first) over a bare `go test` when CRD/RBAC-affecting code changed.

Local tool binaries (controller-gen, kustomize, operator-sdk, golangci-lint, setup-envtest) live in `bin/`, installed by their respective Makefile targets; `make <tool>` before first use if a command isn't found.

## Architecture

This is a Kubebuilder v4 operator with a single API kind, `Immich` (`media.rm3l.org/v1alpha1`), and a single reconciler (`internal/controller/immich_controller.go`). `Immich.Spec` (in `api/v1alpha1/immich_types.go`, ~1500 lines) is the source of truth for five sub-components, each reconciled independently and in this fixed order every `Reconcile` call:

1. **Library PVC** (`library.go`) — only if `ShouldCreateLibraryPVC()`; deliberately has **no owner reference**, so the PVC survives CR deletion (same for the Postgres data PVC, created via StatefulSet VolumeClaimTemplate).
2. **Config** (`config.go`) — merges CR-derived settings (e.g. ML enabled/URL) with user-supplied `spec.immich.configuration` into a ConfigMap or Secret (`spec.immich.configurationKind`).
3. **Postgres** (`postgres.go`) — built-in StatefulSet, or external DB via `spec.postgres.host`/`passwordSecretRef`/`urlSecretRef` when `enabled: false`.
4. **Valkey** (`valkey.go`) — built-in StatefulSet, or external Redis-compatible endpoint when `enabled: false`.
5. **Machine Learning** (`machine_learning.go`) — built-in Deployment, external URL, or fully disabled.
6. **Server** (`server.go`, the largest component file) — the Immich web/API Deployment, Service, optional Ingress and, when the OpenShift Route API is discovered at runtime (`IsRouteAPIAvailable`, cached via `DiscoveryClient`), an OpenShift Route.

Each component's presence/image is driven by accessor methods on `*mediav1alpha1.Immich` (`IsXEnabled()`, `GetXImage()`, etc.) rather than reading `Spec` fields directly in the controller — keep using those accessors when adding logic so defaulting stays centralized in `api/v1alpha1`.

Images resolve CR spec > `RELATED_IMAGE_*` env vars on the operator Deployment (constants in `api/v1alpha1`, e.g. `EnvRelatedImageImmich`) > none (`validateImages` in `validation.go` fails reconciliation with a `Degraded` condition if a required image is missing). This env-var fallback exists specifically for disconnected/air-gapped (e.g. OpenShift) installs.

Status aggregation (`status.go`) rolls up per-component readiness into `Immich.Status.{Ready,ServerReady,MachineLearningReady,ValkeyReady,PostgresReady}` plus standard `Ready`/`Progressing`/`Degraded` conditions; a finalizer (`media.rm3l.org/finalizer`) is added/removed per reconcile but currently does no extra cleanup — owned-resource garbage collection is left to Kubernetes owner references (`SetupWithManager`'s `Owns(...)` list).

`config/rbac/immich_{admin,editor,viewer}_role.yaml` define aggregated ClusterRoles for the `Immich` CRD (shipped in the OLM bundle) — update these alongside RBAC markers in `immich_controller.go` if resource verbs change.

### Bundle / OLM

`bundle/` and `dist/install.yaml` are generated, checked-in artifacts — never hand-edit them; regenerate via `make bundle build-installer` after changing `config/manifests/bases/immich-operator.clusterserviceversion.yaml`, `config/samples/*` (feeds `alm-examples`), or CRD/RBAC. See `CONTRIBUTING.md` for the full version-bump and OperatorHub submission process, including a known gotcha: the CSV's `containerImage` annotation is not auto-updated by `make bundle` and must be bumped by hand in the base file.

### CI

`.github/workflows/`: `test.yml` (unit/envtest), `test-e2e.yml` (Kind), `lint.yml` (`make validate`), `build-push.yml` (builds/signs/pushes multi-arch operator + bundle images to `ghcr.io/rm3l/immich-operator*` on push to `main`, reading `VERSION` from the `Makefile`).
