.PHONY: help build run test clean docker-build docker-run docker-login docker-push docker-pull

# Harbor registry configuration
HARBOR_REGISTRY ?= harbor.dataknife.net
HARBOR_PROJECT ?= library
IMAGE_NAME ?= proxmox-ve-mcp
IMAGE_TAG ?= latest
FULL_IMAGE = $(HARBOR_REGISTRY)/$(HARBOR_PROJECT)/$(IMAGE_NAME):$(IMAGE_TAG)

help:
	@echo "Proxmox VE MCP Server - Available targets:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the server"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo "  docker-login  - Login to Harbor registry"
	@echo "  docker-push   - Push Docker image to Harbor"
	@echo "  docker-pull   - Pull Docker image from Harbor"
	@echo ""
	@echo "Harbor configuration:"
	@echo "  HARBOR_REGISTRY=$(HARBOR_REGISTRY)"
	@echo "  HARBOR_PROJECT=$(HARBOR_PROJECT)"
	@echo "  IMAGE_NAME=$(IMAGE_NAME)"
	@echo "  IMAGE_TAG=$(IMAGE_TAG)"
	@echo "  FULL_IMAGE=$(FULL_IMAGE)"

build:
	@echo "Building Proxmox VE MCP Server..."
	@mkdir -p bin
	go build -o bin/proxmox-ve-mcp ./cmd

run: build
	@echo "Running Proxmox VE MCP Server..."
	./bin/proxmox-ve-mcp

test:
	@echo "Running tests..."
	go test -v -cover ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

docker-build:
	@echo "Building Docker image: $(FULL_IMAGE)"
	docker build -t $(FULL_IMAGE) .
	@echo "Also tagging as $(IMAGE_NAME):$(IMAGE_TAG) for local use"
	docker tag $(FULL_IMAGE) $(IMAGE_NAME):$(IMAGE_TAG)

docker-login:
	@echo "Logging into Harbor registry..."
	@if [ -z "$(HARBOR_USER)" ] || [ -z "$(HARBOR_PASSWORD)" ]; then \
		echo "Error: HARBOR_USER and HARBOR_PASSWORD must be set"; \
		echo "Usage: make docker-login HARBOR_USER='user' HARBOR_PASSWORD='pass'"; \
		exit 1; \
	fi
	@echo '$(HARBOR_PASSWORD)' | docker login $(HARBOR_REGISTRY) -u '$(HARBOR_USER)' --password-stdin

docker-push: docker-build docker-login
	@echo "Pushing Docker image to Harbor: $(FULL_IMAGE)"
	docker push $(FULL_IMAGE)

docker-pull:
	@echo "Pulling Docker image from Harbor: $(FULL_IMAGE)"
	docker pull $(FULL_IMAGE)
	docker tag $(FULL_IMAGE) $(IMAGE_NAME):$(IMAGE_TAG)

docker-run:
	@echo "Running in Docker..."
	docker-compose up
