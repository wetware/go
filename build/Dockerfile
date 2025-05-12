# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ww-go .

# Final stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/ww-go /ww-go

# Copy any additional required files (like certificates)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set the entrypoint
ENTRYPOINT ["/ww-go"] 