# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Kubernetes controller that monitors a ConfigMap named `thing` in the `thing` namespace and watches for changes. It uses the Kubernetes client-go library to implement a simple informer-based controller pattern.

**Target Platform:** OpenShift (compatible with vanilla Kubernetes)

## Development Setup

**Prerequisites:**
- Go 1.24 or later (managed via go.mod toolchain directive)
- Docker or Podman for building container images
- kubectl configured to access your OpenShift/Kubernetes cluster
- Access to an image registry (default: quay.io/cdoan1)

**Dependencies:**
- k8s.io/client-go - Kubernetes client library
- k8s.io/api - Kubernetes API types
- k8s.io/apimachinery - Kubernetes API machinery
- k8s.io/klog/v2 - Structured logging

## Build and Development Commands

All commands are defined in the Makefile. Run `make help` to see all available targets.

**Local Development:**
```bash
make tidy          # Download and tidy dependencies
make fmt           # Format Go code
make vet           # Run Go vet linter
make build         # Build the controller binary to bin/controller
make test          # Run tests
make run           # Run controller locally (requires kubeconfig)
```

**Container Image:**
```bash
make docker-build           # Build image with Docker
make docker-push            # Push image with Docker
make podman-build          # Build image with Podman (preferred for OpenShift)
make podman-push           # Push image with Podman
```

**Deployment:**
```bash
make deploy         # Deploy all manifests to cluster
make undeploy       # Remove all resources from cluster
make redeploy       # Undeploy and deploy (full refresh)
make status         # Show controller and ConfigMap status
make logs           # Tail controller logs
make update-configmap  # Update ConfigMap to trigger controller event
```

**Full Workflow:**
```bash
make all            # Build, docker-build, docker-push, deploy
make all-podman     # Build, podman-build, podman-push, deploy
```

## Architecture

**Controller Pattern:**
This controller uses the informer pattern from client-go:

1. **SharedInformer**: Watches the "thing" ConfigMap in the "thing" namespace using a field selector for efficiency
2. **Event Handlers**: Three handlers for Add, Update, and Delete events
3. **Cache Sync**: Waits for cache synchronization before processing events
4. **Graceful Shutdown**: Handles SIGINT/SIGTERM for clean shutdown

**Code Structure:**
- `main.go` - Single-file controller implementation containing:
  - `Controller` struct - Manages the informer and clientset
  - `NewController()` - Initializes the controller with event handlers
  - `Run()` - Starts the informer and manages lifecycle
  - `handleAdd()`, `handleUpdate()`, `handleDelete()` - Event handlers that log ConfigMap changes
  - `buildConfig()` - Handles both in-cluster and kubeconfig-based authentication

**Key Design Decisions:**
- Uses a field selector to watch only the specific ConfigMap (more efficient than watching all ConfigMaps)
- 30-second resync period for periodic reconciliation
- Read-only operations (watch, list, get) - does not modify resources
- Structured logging with klog for consistency with Kubernetes ecosystem

## Kubernetes Integration

**Namespace:** `thing`
The controller operates entirely within the "thing" namespace.

**RBAC Requirements:**
- **ServiceAccount:** `thing-controller` in `thing` namespace
- **Role:** `thing-controller-role` with permissions:
  - `configmaps`: get, list, watch
- **RoleBinding:** Links the Role to the ServiceAccount

Note: Uses a Role (not ClusterRole) since the controller only watches a single namespace.

**Deployment Configuration:**
- **Image:** quay.io/cdoan1/thing-controller:latest
- **Replicas:** 1 (single instance is sufficient for this use case)
- **Resources:**
  - Requests: 50m CPU, 64Mi memory
  - Limits: 100m CPU, 128Mi memory
- **Security:** OpenShift-compatible security context (non-root, no privilege escalation, dropped capabilities)

**ConfigMap Schema:**
The controller watches the `thing` ConfigMap but does not enforce any specific schema. It logs all data keys and values when changes occur.

**Command-line Flags:**
- `--namespace=thing` - Target namespace (default: thing)
- `--configmap=thing` - ConfigMap to watch (default: thing)
- `-v=2` - Log verbosity level

## Container Image

**Multi-stage Build:**
1. Builder stage: Uses Red Hat UBI9 Go toolset (1.24)
2. Final stage: Uses Red Hat UBI9 micro (minimal base image)

**Security:**
- Runs as non-root user (UID 65532)
- CGO disabled for static binary
- Minimal attack surface using ubi-micro

**Registry:**
Default: quay.io/cdoan1/thing-controller:latest
Configure via Makefile variables: IMAGE_REGISTRY, IMAGE_REPO, IMAGE_NAME, IMAGE_TAG

## Deployment Files

All Kubernetes manifests are in the `deploy/` directory:
- `namespace.yaml` - Creates the "thing" namespace
- `serviceaccount.yaml` - ServiceAccount for the controller
- `rbac.yaml` - Role and RoleBinding for ConfigMap access
- `configmap.yaml` - Sample ConfigMap that the controller watches
- `deployment.yaml` - Controller deployment

Apply order matters: namespace → serviceaccount → rbac → configmap → deployment

## License

Apache License 2.0
