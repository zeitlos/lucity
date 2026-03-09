.PHONY: build proto dev dev-gateway dev-builder dev-packager dev-deployer dev-cashier dev-webhook dev-dashboard dev-docs dev-logs dev-stop generate-graphql lint test-integration test-integration-short test-watch minikube dns infra infra-down infra-forward infra-forward-stop argocd-password infra-tokens argocd-token softserve-token db-forward deploy-prod deploy-prod-infra

# Build all Go services
build:
	go build ./services/gateway/cmd/gateway/...
	go build ./services/builder/cmd/builder/...
	go build ./services/packager/cmd/packager/...
	go build ./services/deployer/cmd/deployer/...
	go build ./services/webhook/cmd/webhook/...
	go build ./services/cashier/cmd/cashier/...

# Generate protobuf code (requires protoc, protoc-gen-go, protoc-gen-go-grpc)
proto:
	cd pkg && go generate ./...

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
	@for port in 8080 9001 9002 9003 9004 9005 9006 5173; do \
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

dev-cashier:
	cd services/cashier && set -a && . .env 2>/dev/null && set +a && go run ./cmd/cashier/...

dev-dashboard:
	cd services/dashboard && npm run dev

dev-docs:
	cd docs && npm run dev

# Lint
lint:
	cd services/dashboard && npm run lint

# Integration tests (requires services running via make dev)
test-integration:
	@mkdir -p tmp/logs
	cd tests && go test -v -count=1 -run TestIntegration ./... 2>&1 | tee ../tmp/logs/tests.log

test-integration-short:
	@mkdir -p tmp/logs
	cd tests && go test -v -count=1 -short -run TestIntegration ./... 2>&1 | tee ../tmp/logs/tests.log

# Watch mode: re-run tests on file changes (requires watchexec)
test-watch:
	@bash scripts/test-watch.sh

# Create minikube cluster for local development
# --insecure-registry covers the entire service CIDR so Docker trusts Zot over HTTP.
# See: https://minikube.sigs.k8s.io/docs/handbook/registry/#enabling-insecure-registries
minikube:
	minikube start --insecure-registry="10.96.0.0/12"

# Set up local DNS so *.lucity.local resolves to 127.0.0.1 (requires Homebrew)
# Run once — survives reboots. Uses dnsmasq on port 5380 (unprivileged) with
# macOS /etc/resolver pointing to it. No root needed for dnsmasq itself.
dns:
	@if ! command -v dnsmasq >/dev/null 2>&1; then \
		echo "Installing dnsmasq..."; \
		brew install dnsmasq; \
	fi
	@if ! grep -q 'port=5380' /opt/homebrew/etc/dnsmasq.conf 2>/dev/null; then \
		echo "port=5380" >> /opt/homebrew/etc/dnsmasq.conf; \
		echo "Set dnsmasq to listen on port 5380."; \
	fi
	@if ! grep -q 'address=/lucity.local/' /opt/homebrew/etc/dnsmasq.conf 2>/dev/null; then \
		echo "address=/lucity.local/127.0.0.1" >> /opt/homebrew/etc/dnsmasq.conf; \
		echo "Added *.lucity.local → 127.0.0.1 to dnsmasq config."; \
	else \
		echo "dnsmasq already configured for *.lucity.local."; \
	fi
	@sudo mkdir -p /etc/resolver
	@printf "nameserver 127.0.0.1\nport 5380\n" | sudo tee /etc/resolver/lucity.local > /dev/null
	@brew services restart dnsmasq
	@echo "Done. All *.lucity.local domains now resolve to 127.0.0.1."

# Deploy infrastructure (Zot + Soft-serve + ArgoCD + Envoy Gateway + CNPG) to a cluster
# Envoy Gateway is installed separately — it needs its own namespace for cert management.
# Usage: make infra CLUSTER=flxp
CLUSTER ?= minikube
infra:
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml
	helm upgrade --install envoy-gateway oci://docker.io/envoyproxy/gateway-helm \
		--version v1.2.6 -n envoy-gateway-system --create-namespace --skip-crds
	@kubectl apply -f - <<< '{"apiVersion":"gateway.networking.k8s.io/v1","kind":"GatewayClass","metadata":{"name":"eg"},"spec":{"controllerName":"gateway.envoyproxy.io/gatewayclass-controller"}}'
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
	@(GATEWAY_SVC=$$(kubectl get svc -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=lucity-gateway -o name 2>/dev/null) && \
		[ -n "$$GATEWAY_SVC" ] && kubectl port-forward $$GATEWAY_SVC 8880:80 -n envoy-gateway-system &) 2>/dev/null || true
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

# Port-forward a project's database for local development (interactive picker)
db-forward:
	@bash scripts/db-forward.sh

# Production deployment from OCI registry charts
# Usage: make deploy-prod VERSION=0.0.0-85-gabcdef0 HELM_ARGS="..."
# Secrets are stored in deployments/lucity-prod/secrets.yaml (gitignored).
# First deploy: create secrets.yaml from secrets.yaml.example, then run deploy-prod.
PROD_CONTEXT ?= lucity-prod
VERSION ?=

deploy-prod-infra:
	@test -f deployments/lucity-prod/secrets.yaml || { echo "Error: deployments/lucity-prod/secrets.yaml not found. Copy secrets.yaml.example and fill in values."; exit 1; }
	helm upgrade --install lucity-infra \
		oci://ghcr.io/zeitlos/lucity/charts/lucity-infra \
		$(if $(VERSION),--version $(VERSION)) \
		--kube-context $(PROD_CONTEXT) \
		-n lucity-system --create-namespace \
		-f deployments/lucity-prod/infra-values.yaml \
		-f deployments/lucity-prod/secrets.yaml \
		$(HELM_ARGS)

deploy-prod:
	@test -f deployments/lucity-prod/secrets.yaml || { echo "Error: deployments/lucity-prod/secrets.yaml not found. Copy secrets.yaml.example and fill in values."; exit 1; }
	helm upgrade --install lucity \
		oci://ghcr.io/zeitlos/lucity/charts/lucity \
		$(if $(VERSION),--version $(VERSION)) \
		--kube-context $(PROD_CONTEXT) \
		-n lucity-system --create-namespace \
		-f deployments/lucity-prod/values.yaml \
		-f deployments/lucity-prod/secrets.yaml \
		$(HELM_ARGS)

# Sync workspace
sync:
	go work sync
