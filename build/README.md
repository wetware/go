# Container Build

This directory contains the Dockerfile and related files for building the Wetware container image.

## Registries

The container image is pushed to two registries:

1. GitHub Container Registry (ghcr.io)
   - Public registry
   - Format: `ghcr.io/wetware/go`
   - Tags: latest, commit SHA, branch name, tag name

2. Wetware Private Registry
   - Private registry at container-registry.wetware.run
   - Format: `container-registry.wetware.run/wetware/go`
   - Tags: latest, commit SHA

## Building Locally

To build the container locally:

```bash
docker build -t wetware/go -f build/Dockerfile .
```

## Registry Access

- GitHub Container Registry (ghcr.io): Public registry, requires GitHub authentication for push
- Wetware Registry (container-registry.wetware.run): Private registry, requires Wetware registry credentials

# Container Build Process

This directory contains the container build configuration for the Wetware Go project.

## Build Strategy

The container build uses a multi-stage approach to create a minimal production image:

1. **Builder Stage** (`golang:1.22.1-alpine`)
   - Uses Alpine Linux for a small build environment
   - Installs git and build dependencies
   - Compiles a statically linked binary with CGO disabled
   - Outputs a single binary with no external dependencies

2. **Production Stage** (`scratch`)
   - Starts from an empty image
   - Contains only:
     - The compiled binary (`/ww-go`)
     - SSL certificates (for HTTPS connections)
   - Minimal attack surface and image size

## Building Locally

The build process is managed through Make targets:

```bash
# Build with defaults (ghcr.io/wetware/go:latest)
make container-build

# Build with custom registry
make container-build REGISTRY=myregistry.com

# Build with custom tag
make container-build TAG=v1.0.0

# Build and push
make container-push

# Build and push with custom values
make container-push REGISTRY=myregistry.com IMAGE_NAME=myorg/myapp TAG=v1.0.0
```

## Private Registry

The project supports pushing to a private registry running on port 5000:

```bash
# Build and push to private registry
make container-push-private \
    PRIVATE_REGISTRY=your-server:5000 \
    PRIVATE_USERNAME=your-username \
    PRIVATE_PASSWORD=your-password

# Or set environment variables
export PRIVATE_REGISTRY=your-server:5000
export PRIVATE_USERNAME=your-username
export PRIVATE_PASSWORD=your-password
make container-push-private
```

The private registry target will:
1. Log in to the registry
2. Tag the image for the private registry
3. Push the image
4. Test pulling the image back to verify the setup

## CI/CD Integration

The container build is integrated into the CI/CD pipeline:

1. Builds are triggered on:
   - Push to master
   - Pull requests
   - Tags

2. Images are automatically tagged with:
   - `latest` (on master branch)
   - Git SHA (short format)
   - Branch name
   - Tag name (if a tag is pushed)

3. Multi-platform support:
   - linux/amd64
   - linux/arm64

## Security Considerations

- Uses `scratch` base image for minimal attack surface
- Statically linked binary with no external dependencies
- Includes only necessary SSL certificates
- Built in CI with reproducible results 