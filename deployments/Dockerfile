# Build stage
FROM golang:1.21-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker ./cmd/worker

# Final stage
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /root/

# Copy binaries from builder stage
COPY --from=builder /app/bin/server .
COPY --from=builder /app/bin/worker .

# Copy config files
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Default command (can be overridden)
CMD ["./server"]
