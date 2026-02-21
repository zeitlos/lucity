.PHONY: build proto dev dev-gateway dev-builder dev-packager dev-deployer dev-webhook dev-dashboard dev-docs dev-logs dev-stop generate-graphql lint test-integration test-integration-short minikube infra infra-down infra-forward infra-forward-stop argocd-password infra-tokens argocd-token softserve-token

# Build all Go services
build:
	go build ./services/gateway/cmd/gateway/...
	go build ./services/builder/cmd/builder/...
	go build ./services/packager/cmd/packager/...
	go build ./services/deployer/cmd/deployer/...
	go build ./services/webhook/cmd/webhook/...

# Generate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
proto:
	cd pkg/builder && go generate ./...
	cd pkg/deployer && go generate ./...
	cd pkg/packager && go generate ./...
	cd pkg/webhook && go generate ./...

# Generate GraphQL resolvers (requires gqlgen)
generate-graphql:
	cd services/gateway && go generate ./graphql/resolver.go

# Start all services with hot reload (air + vite)
dev:
	@bash scripts/dev.sh

# View all service logs
dev-logs:
	@tail -f tmp/logs/*.log

# View specific service logs (e.g., make dev-logs-gateway)
dev-logs-%:
	@tail -f tmp/logs/$*.log

# Stop all dev services
dev-stop:
	@for port in 8080 9001 9002 9003 9004 5173; do \
		lsof -ti :$$port -sTCP:LISTEN | xargs kill 2>/dev/null || true; \
	done
	@echo "All services stopped."

# Run individual services (without air)
dev-gateway:
	cd services/gateway && set -a && . .env 2>/dev/null && set +a && go run ./cmd/gateway/...

dev-builder:
	cd services/builder && set -a && . .env 2>/dev/null && set +a && go run ./cmd/builder/...

dev-packager:
	cd services/packager && set -a && . .env 2>/dev/null && set +a && go run ./cmd/packager/...

dev-deployer:
	cd services/deployer && set -a && . .env 2>/dev/null && set +a && go run ./cmd/deployer/...

dev-webhook:
	cd services/webhook && set -a && . .env 2>/dev/null && set +a && go run ./cmd/webhook/...

dev-dashboard:
	cd services/dashboard && npm run dev

dev-docs:
	cd docs && npm run dev

# Lint
lint:
	cd services/dashboard && npm run lint

# Integration tests (requires services running via make dev)
test-integration:
	cd tests && go test -v -count=1 ./...

test-integration-short:
	cd tests && go test -v -count=1 -short ./...

# Create minikube cluster for local development
# --insecure-registry covers the entire service CIDR so Docker trusts Zot over HTTP.
# See: https://minikube.sigs.k8s.io/docs/handbook/registry/#enabling-insecure-registries
minikube:
	minikube start --insecure-registry="10.96.0.0/12"

# Deploy infrastructure (Zot + Soft-serve + ArgoCD + Envoy Gateway) to a cluster
# Also installs Gateway API CRDs required for HTTPRoutes
# Usage: make infra CLUSTER=flxp
CLUSTER ?= minikube
infra:
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml
	helm dependency update charts/lucity-infra
	helm upgrade --install lucity-infra charts/lucity-infra \
		-n lucity-system --create-namespace \
		-f deployments/$(CLUSTER)/values.yaml

infra-down:
	helm uninstall lucity-infra -n lucity-system

# Port-forward infrastructure services for local development
infra-forward: infra-forward-stop
	@echo "Port-forwarding Zot (5000), Soft-serve (23231, 23232), ArgoCD (8443), and Gateway (8880)..."
	@kubectl port-forward svc/lucity-infra-zot 5000:5000 -n lucity-system &
	@kubectl port-forward svc/lucity-infra-soft-serve 23231:23231 23232:23232 -n lucity-system &
	@kubectl port-forward svc/lucity-infra-argocd-server 8443:80 -n lucity-system &
	@kubectl port-forward svc/envoy-lucity-system-lucity-gateway 8880:80 -n envoy-gateway-system 2>/dev/null &
	@echo "Ready. Use 'make infra-forward-stop' to stop."

infra-forward-stop:
	@for port in 5000 23231 23232 8443 8880; do \
		lsof -ti :$$port | xargs kill 2>/dev/null || true; \
	done

# Print the ArgoCD admin password (auto-generated, stored in K8s secret)
argocd-password:
	@kubectl get secret argocd-initial-admin-secret -n lucity-system -o jsonpath='{.data.password}' | base64 -d && echo

# Generate API tokens for infrastructure services
# Requires: infra-forward running
infra-tokens: argocd-token softserve-token

# Generate an ArgoCD API token for the lucity service account
# Requires: infra-forward running (ArgoCD on localhost:8443)
argocd-token:
	@ADMIN_PASS=$$(kubectl get secret argocd-initial-admin-secret -n lucity-system -o jsonpath='{.data.password}' | base64 -d) && \
	SESSION=$$(curl -sk -H "Content-Type: application/json" http://localhost:8443/api/v1/session -d "{\"username\":\"admin\",\"password\":\"$$ADMIN_PASS\"}" | jq -r '.token') && \
	TOKEN=$$(curl -sk -H "Content-Type: application/json" -H "Authorization: Bearer $$SESSION" -X POST http://localhost:8443/api/v1/account/lucity/token | jq -r '.token') && \
	echo "ARGOCD_TOKEN=$$TOKEN"

# Generate a Soft-serve access token for the packager
# Requires: infra-forward running (Soft-serve SSH on localhost:23231)
softserve-token:
	@ssh-keygen -R "[localhost]:23231" 2>/dev/null || true
	@ssh -i ~/.ssh/lucity-admin-minikube -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new -p 23231 localhost token create 'packager'

# Sync workspace
sync:
	go work sync
