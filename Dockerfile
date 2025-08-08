# Multi-stage build for dockrune Go binary

# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -o dockrune ./cmd/dockrune

# Stage 2: Runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    docker-cli \
    docker-cli-compose \
    nodejs \
    npm \
    python3 \
    py3-pip \
    bash \
    curl \
    sqlite \
    tzdata

# Install pm2 globally for process management
RUN npm install -g pm2

# Create non-root user
RUN addgroup -g 1000 dockrune && \
    adduser -D -u 1000 -G dockrune dockrune && \
    addgroup dockrune docker

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/dockrune /usr/local/bin/dockrune
RUN chmod +x /usr/local/bin/dockrune

# Copy static files if they exist
COPY --chown=dockrune:dockrune static/ ./static/ 2>/dev/null || true

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/repos && \
    chown -R dockrune:dockrune /app

# Switch to non-root user
USER dockrune

# Environment variables
ENV DATABASE_PATH=/app/data/dockrune.db \
    REPOS_DIR=/app/repos \
    LOGS_DIR=/app/logs \
    WEBHOOK_PORT=8000 \
    ADMIN_PORT=8001

# Expose ports
EXPOSE 8000 8001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Volume mounts
VOLUME ["/app/data", "/app/logs", "/app/repos"]

# Entry point
ENTRYPOINT ["dockrune"]
CMD ["serve"]