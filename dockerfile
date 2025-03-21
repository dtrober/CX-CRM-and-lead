# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp ./cmd/api

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Create a non-root user 
RUN addgroup -S appgroup && adduser -S appuser -G appgroup


# Copy the binary from the builder stage
COPY --from=builder /app/myapp .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy web assets
COPY --from=builder /app/web ./web

# Change ownership of web directory to appuser
RUN chown -R appuser:appgroup ./web

#Switch to non sudo user
USER appuser

# Expose port
EXPOSE 8080

# Define health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./myapp"]