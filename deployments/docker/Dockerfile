# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/node ./cmd/node

# Final Stage
FROM alpine:3.18

WORKDIR /app

# Add non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binary from builder
COPY --from=builder /app/node .

# Default configuration (can be overridden by env vars)
ENV PORT=8000
ENV DATA_DIR=/app/data

# Expose port
EXPOSE 8000

# Create data directory with correct permissions
RUN mkdir -p /app/data

# Health check (basic TCP check using netcat/nc)
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD nc -z localhost 8000 || exit 1

ENTRYPOINT ["./node"]
