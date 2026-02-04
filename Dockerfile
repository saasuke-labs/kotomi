# Build stage - includes all dependencies
FROM golang:1.25-alpine AS build

# Install build dependencies for CGO and SQLite
# Note: If apk fails due to network issues, this Dockerfile may need adjustment
RUN apk add --no-cache gcc musl-dev sqlite-dev || \
    (echo "Note: apk may fail in restricted environments. Build will use cached layers if available" && exit 1)

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the application with CGO enabled for SQLite
# Static linking to avoid runtime dependencies
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -linkmode external -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -tags 'osusergo netgo static_build' \
    -o /app/kotomi ./cmd/main.go

# Runtime stage - minimal image
FROM alpine:3.19

WORKDIR /app

# Create non-root user and directories
RUN addgroup -g 1000 kotomi 2>/dev/null || true && \
    adduser -D -u 1000 -G kotomi kotomi 2>/dev/null || true && \
    mkdir -p /app/data && \
    chown -R 1000:1000 /app/data 2>/dev/null || true

# Copy statically linked binary from build stage
COPY --from=build /app/kotomi .

# Copy static assets
COPY --chown=1000:1000 templates ./templates
COPY --chown=1000:1000 static ./static

# Switch to non-root user
USER kotomi

# Declare volume for database persistence
VOLUME ["/app/data"]

# Set default environment variables
ENV DB_PATH=/app/data/kotomi.db \
    PORT=8080

# Expose port
EXPOSE 8080

# Health check (may not work in all environments without wget/curl)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz 2>/dev/null || exit 1

# Run the application
CMD ["./kotomi"]