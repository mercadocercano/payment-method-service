# ==============================================
# Payment Method Service - Optimized Multi-stage Dockerfile
# ==============================================

# ==============================================
# Stage 1: Dependencies and cache optimization
# ==============================================
FROM golang:1.25-alpine AS deps
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

# Copy dependency files and download modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ==============================================
# Stage 2: Build stage
# ==============================================
FROM deps AS builder

# Copy source code
COPY . .

# Build optimized binary with security hardening
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o payment-method-service ./src/main.go

# ==============================================
# Stage 3: Development stage
# ==============================================
FROM golang:1.25-alpine AS development

# Security: Create non-root user first
RUN addgroup -g 1001 -S appgroup && \
    adduser -S -D -h /app -s /bin/sh -G appgroup -u 1001 appuser

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    postgresql-client \
    git \
    && cp /usr/share/zoneinfo/UTC /etc/localtime \
    && echo "UTC" > /etc/timezone \
    && apk del tzdata

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Create necessary directories and set permissions
RUN mkdir -p tmp logs /go/pkg/mod && \
    chmod -R 777 /go/pkg && \
    chown -R appuser:appgroup /app tmp logs

# Copy source code with correct ownership
COPY --chown=appuser:appgroup . .

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

EXPOSE 8080

CMD sh -c 'if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi && air -c .air.toml'

# ==============================================
# Stage 4: Migrate stage (Alpine + psql para Job K8s)
# Sin dependencia de builder - solo copia migrations/seeds (build rápido)
# ==============================================
FROM alpine:3.18 AS migrate

RUN apk add --no-cache postgresql-client

WORKDIR /app
COPY migrations ./migrations
COPY seeds ./seeds

# ==============================================
# Stage 5: Production stage (Distroless)
# ==============================================
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Metadata
LABEL org.opencontainers.image.title="Payment Method Service" \
      org.opencontainers.image.description="Multi-tenant Payment Method Management (Read-Only MVP)" \
      org.opencontainers.image.source="https://github.com/mercadocercano/payment-method-service" \
      org.opencontainers.image.vendor="Mercado Cercano Team" \
      org.opencontainers.image.licenses="MIT"

WORKDIR /app

# Copy binary and essential files only
COPY --from=builder --chown=nonroot:nonroot /app/payment-method-service ./
COPY --from=builder --chown=nonroot:nonroot /app/migrations ./migrations/

# Use distroless nonroot user (uid=65532)
USER nonroot

EXPOSE 8080

ENTRYPOINT ["./payment-method-service"]

# ==============================================
# Default stage: Development
# ==============================================
FROM development
