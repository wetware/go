# ┌─────────────── builder ───────────────┐
FROM golang:1.24-alpine AS builder

# Ensure our binary is statically linked
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /src

# Cache go.mod/go.sum and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your code and compile
COPY . .
# -trimpath strips file system paths, -s -w strip symbol table/debug info
RUN go build -o ww \
    -trimpath \
    # -ldflags="-s -w" \
    ./cmd/ww

# └───────────────────────────────────────┘


# ┌─────────────── final ────────────────┐
FROM alpine:latest

# (optional) if your binary uses TLS, pull a CA bundle:
# FROM alpine:latest AS certs
# RUN apk --no-cache add ca-certificates

# copy the compiled binary
COPY --from=builder /src/ww /ww

# (optional) copy certs if needed:
# COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# allow CommandContext to find "ww"
ENV PATH="/:${PATH}"  

ENTRYPOINT ["ww"]
CMD ["--help"]
# └───────────────────────────────────────┘
