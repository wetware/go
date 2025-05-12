.PHONY: clean container-build container-push container-push-private

# Container registry and image settings
REGISTRY ?= ghcr.io
IMAGE_NAME ?= wetware/go
TAG ?= latest

# Private registry settings
PRIVATE_REGISTRY ?= localhost:5000
PRIVATE_USERNAME ?= 
PRIVATE_PASSWORD ?= 

all: generate install publish

clean:
	@if [ -f "ww" ]; then rm ww; fi
	@rm -f $(GOPATH)/bin/ww

generate:
	go generate ./...

publish:
	ipfs add -r .

install:
	go install github.com/wetware/go/cmd/ww

deploy: install && publish

# Container targets
container-build:
	docker build \
		--file build/Dockerfile \
		--tag $(REGISTRY)/$(IMAGE_NAME):$(TAG) \
		--platform linux/amd64,linux/arm64 \
		.

container-push: container-build
	docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Private registry targets
container-push-private: container-build
	@if [ -z "$(PRIVATE_USERNAME)" ] || [ -z "$(PRIVATE_PASSWORD)" ]; then \
		echo "Error: PRIVATE_USERNAME and PRIVATE_PASSWORD must be set"; \
		exit 1; \
	fi
	@echo "Logging in to private registry..."
	@echo "$(PRIVATE_PASSWORD)" | docker login $(PRIVATE_REGISTRY) -u $(PRIVATE_USERNAME) --password-stdin
	@echo "Tagging image for private registry..."
	docker tag $(REGISTRY)/$(IMAGE_NAME):$(TAG) $(PRIVATE_REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "Pushing to private registry..."
	docker push $(PRIVATE_REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "Testing pull from private registry..."
	docker pull $(PRIVATE_REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "Successfully pushed and pulled from private registry"
