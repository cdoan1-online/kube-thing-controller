# Kube Thing Controller

A Kubernetes controller that monitors a ConfigMap named `thing` in the `thing` namespace and watches for changes.

## Overview

This controller uses the Kubernetes client-go library to watch for changes to a specific ConfigMap and logs events when the ConfigMap is added, updated, or deleted.

## Prerequisites

- Go 1.24 or later
- Docker or Podman
- kubectl configured to access your OpenShift/Kubernetes cluster
- Access to an image registry (default: quay.io)

## Quick Start

### Build and Run Locally

```bash
# Download dependencies
make tidy

# Build the binary
make build

# Run locally (requires kubeconfig)
make run
```

### Build and Deploy to OpenShift

```bash
# Build container image
make podman-build

# Push to registry (update IMAGE_REPO in Makefile first)
make podman-push

# Deploy to cluster
make deploy
```

### Verify Deployment

```bash
# Check controller status
make status

# View controller logs
make logs

# Test by updating the ConfigMap
make update-configmap
```

## Configuration

The controller can be configured using command-line flags:

- `--namespace`: Namespace to watch (default: "thing")
- `--configmap`: ConfigMap name to watch (default: "thing")
- `--kubeconfig`: Path to kubeconfig file (optional, uses in-cluster config if not provided)
- `-v`: Log verbosity level

## Development

```bash
# Format code
make fmt

# Run linter
make vet

# Run tests
make test
```

## Cleanup

```bash
# Remove controller from cluster
make undeploy
```

## License

Apache License 2.0
