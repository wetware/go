# Build stage
FROM golang:1.24-alpine AS builder

# Install git
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy the entire project from root context
COPY . /build/

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ww-go ./cmd/ww

# Final stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/ww-go /ww-go

# Copy any additional required files (like certificates)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the entrypoint
ENTRYPOINT ["/ww-go"] 