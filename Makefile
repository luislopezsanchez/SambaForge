.PHONY: all dev build build-api build-web test lint clean deploy

SAMBAFORGE_VERSION ?= 0.0.1-dev
SAMBAFORGE_PORT ?= 8444

all: build

# Development: run API + Web with hot reload
dev:
	@echo "Starting SambaForge dev environment..."
	@cd apps/api && go run . &
	@cd apps/web && npm run dev
	@wait

# Build everything
build: build-web build-api
	@echo "Build complete: bin/sambaforge"

# Build Go backend (static binary, no CGO)
build-api:
	@echo "Building SambaForge API..."
	@cd apps/api && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(SAMBAFORGE_VERSION)" -o ../../bin/sambaforge .

# Build React frontend (output to apps/api/web for embedding)
build-web:
	@echo "Building SambaForge Web..."
	@cd apps/web && npm ci && npm run build
	@rm -rf apps/api/web/dist
	@cp -r apps/web/dist apps/api/web/dist

# Test
test:
	@cd apps/api && go test -v ./...
	@cd apps/web && npm test -- --run

# Lint
lint:
	@cd apps/api && go vet ./...
	@cd apps/web && npm run lint

# Clean
clean:
	@rm -rf bin/ apps/web/dist apps/api/web/dist
	@cd apps/web && rm -rf node_modules

# Deploy to VM (requires VM_IP env var)
deploy: build
	@if [ -z "$(VM_IP)" ]; then echo "VM_IP not set"; exit 1; fi
	@echo "Deploying to $(VM_IP)..."
	@scp -O bin/sambaforge root@$(VM_IP):/usr/local/bin/sambaforge
	@scp -O deploy/sambaforge.service root@$(VM_IP):/etc/systemd/system/sambaforge.service
	@ssh root@$(VM_IP) "systemctl daemon-reload && systemctl restart sambaforge && systemctl status sambaforge"