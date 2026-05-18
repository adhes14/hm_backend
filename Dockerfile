# ==============================================================================
# STAGE 1: Build stage
# ==============================================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies (git is needed for fetching some go packages)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the main server binary
# CGO_ENABLED=0 creates a static binary that can run on minimal alpine/scratch
# -ldflags="-w -s" strips debugging symbols to drastically reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server/main.go

# Install golang-migrate CLI binary with only postgres support (for container startup migrations)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ==============================================================================
# STAGE 2: Runner stage (lightweight and secure)
# ==============================================================================
FROM alpine:latest AS runner

# Install runtime dependencies (ca-certificates for HTTPS/SSL, tzdata for timezones)
RUN apk add --no-cache ca-certificates tzdata

# Create a non-privileged user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server /app/server

# Copy the migrations directory so we can run migrations on startup
COPY --from=builder /app/migrations /app/migrations

# Copy the compiled golang-migrate CLI binary
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Change ownership of the app directory to the non-privileged user
RUN chown -R appuser:appgroup /app

# Switch to the non-privileged user
USER appuser

# Expose the application port (Railway will automatically map this to the PORT env var)
EXPOSE 8080

# Environment variables with sensible defaults
ENV PORT=8080 \
    GO_ENV=production \
    RUN_MIGRATIONS=true

# Set entrypoint to run migrations and then start the server
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/server"]
