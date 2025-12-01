# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs (optional, if swag is available)
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o openvpn-mng ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/openvpn-mng .

# Copy web templates and static files
COPY --from=builder /app/web ./web

# Copy Swagger docs if they exist
COPY --from=builder /app/docs ./docs

# Create config directory
RUN mkdir -p /app/config

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./openvpn-mng"]
