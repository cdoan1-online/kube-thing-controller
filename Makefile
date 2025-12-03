# Image registry and repository
IMAGE_REGISTRY ?= quay.io
IMAGE_REPO ?= cdoan1
IMAGE_NAME ?= thing-controller
IMAGE_TAG ?= latest
IMAGE ?= $(IMAGE_REGISTRY)/$(IMAGE_REPO)/$(IMAGE_NAME):$(IMAGE_TAG)

# Namespace and ConfigMap name
NAMESPACE ?= thing
CONFIGMAP_NAME ?= thing

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: build
build: fmt vet ## Build controller binary
	go build -o bin/controller main.go

.PHONY: run
run: fmt vet ## Run controller locally (requires kubeconfig)
	go run main.go -v=2

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -v

##@ Container Image

.PHONY: docker-build
docker-build: ## Build container image
	docker build -t $(IMAGE) .

.PHONY: docker-push
docker-push: ## Push container image
	docker push $(IMAGE)

.PHONY: podman-build
podman-build: ## Build container image using podman
	podman build -t $(IMAGE) .

.PHONY: podman-push
podman-push: ## Push container image using podman
	podman push $(IMAGE)

##@ Deployment

.PHONY: deploy
deploy: ## Deploy controller to OpenShift/Kubernetes cluster
	kubectl apply -f deploy/namespace.yaml
	kubectl apply -f deploy/serviceaccount.yaml
	kubectl apply -f deploy/rbac.yaml
	kubectl apply -f deploy/configmap.yaml
	kubectl apply -f deploy/deployment.yaml

.PHONY: undeploy
undeploy: ## Remove controller from cluster
	kubectl delete -f deploy/deployment.yaml --ignore-not-found=true
	kubectl delete -f deploy/configmap.yaml --ignore-not-found=true
	kubectl delete -f deploy/rbac.yaml --ignore-not-found=true
	kubectl delete -f deploy/serviceaccount.yaml --ignore-not-found=true
	kubectl delete -f deploy/namespace.yaml --ignore-not-found=true

.PHONY: redeploy
redeploy: undeploy deploy ## Undeploy and deploy controller

##@ Testing & Debugging

.PHONY: logs
logs: ## Show controller logs
	kubectl logs -n $(NAMESPACE) -l app=thing-controller -f

.PHONY: status
status: ## Show controller status
	kubectl get all -n $(NAMESPACE)
	kubectl get configmap -n $(NAMESPACE)

.PHONY: update-configmap
update-configmap: ## Update the ConfigMap to test the controller
	kubectl patch configmap $(CONFIGMAP_NAME) -n $(NAMESPACE) --type merge -p '{"data":{"timestamp":"'$$(date +%s)'"}}'

##@ Full Workflow

.PHONY: all
all: build docker-build docker-push deploy ## Build, push image, and deploy

.PHONY: all-podman
all-podman: build podman-build podman-push deploy ## Build, push image using podman, and deploy
