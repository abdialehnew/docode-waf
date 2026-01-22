# Build stage
FROM golang:1.25.5-alpine3.21 AS builder

# Install security updates
RUN apk upgrade --no-cache

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the main application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o waf cmd/waf/main.go

# Build the reset-password CLI tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o reset-password cmd/reset-password/main.go

# Final stage - Using distroless for minimal CVE
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates and binaries from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/waf /waf
COPY --from=builder /app/reset-password /app/reset-password
COPY --from=builder /app/config.yaml /config.yaml
COPY --from=builder /app/GeoLite2-Country.mmdb /GeoLite2-Country.mmdb

# Expose ports
EXPOSE 8080 9090

# Run as non-root user
USER nonroot:nonroot

ENTRYPOINT ["/waf"]
