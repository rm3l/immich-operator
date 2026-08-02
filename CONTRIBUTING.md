# Contributing

Thanks for your interest in contributing to the Immich Operator!

## Getting Started

- Go version v1.25.0+
- Docker (or another `CONTAINER_TOOL`) and access to a Kubernetes cluster (e.g. [kind](https://kind.sigs.k8s.io/)) for local testing.
- Fork the repository and create a feature branch off `main`.

```sh
make install      # Install CRDs into your cluster
make run          # Run the controller locally, against your current kubeconfig context
```

## Before Opening a PR

```sh
make manifests generate   # Regenerate CRDs/RBAC and DeepCopy code after changing api/ types
make test                 # Run unit/envtest suite
make validate              # Run golangci-lint and OperatorHub bundle checks
```

CI (`.github/workflows/lint.yml`, `test.yml`, `test-e2e.yml`) runs the same checks on every PR against `main`.

Keep commits focused and prefer squashing before merge. See [`README.md`](README.md) for architecture and usage details.

## Release Process

Releases are cut from `main` and published both as container images and as an OperatorHub bundle submission.

1. **Bump the version.** Update `VERSION` in `Makefile`, then regenerate the bundle and installer manifests:

   ```sh
   make bundle build-installer
   make test
   ```

   Also bump `containerImage` and `version` by hand in `config/manifests/bases/immich-operator.clusterserviceversion.yaml` — `make bundle` does **not** update these automatically. Add/update the `spec.replaces` field there too (e.g. `immich-operator.v0.1.0`) so OLM can upgrade cleanly from the previous version.

2. **Open a PR** with the version bump and let CI (`build-push.yml`) validate it, then merge to `main`. On merge, CI builds, signs (cosign, keyless OIDC), and pushes multi-arch images to `ghcr.io/rm3l/immich-operator:<version>` and `ghcr.io/rm3l/immich-operator-bundle:<version>`.

3. **Tag the release** (`git tag <version>`, `git push --tags`) and create a matching GitHub release.

4. **Submit the bundle to [k8s-operatorhub/community-operators](https://github.com/k8s-operatorhub/community-operators)**: in a fork of that repo, sync `main` with upstream, branch off it, copy `bundle/{manifests,metadata,tests}` and `bundle.Dockerfile` from this repo into `operators/immich-operator/<version>/`, commit with a sign-off (`git commit -s`), and open a PR against `k8s-operatorhub/community-operators` (`main`). See [PR #7327](https://github.com/k8s-operatorhub/community-operators/pull/7327) (initial `0.1.0` submission) and [PR #8859](https://github.com/k8s-operatorhub/community-operators/pull/8859) (`0.2.0` update) for reference.
