# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies including libvips for image processing
RUN apk add --no-cache git ca-certificates tzdata vips-dev gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the worker application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o worker image_worker/internal/cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies for image processing
RUN apk --no-cache add ca-certificates vips

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/worker .

# Copy .env file if exists
COPY --from=builder /app/.env* ./

# Run the worker
CMD ["./worker"]