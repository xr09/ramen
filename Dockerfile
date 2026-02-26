# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ramen ./cmd/ramen

# Final stage
FROM alpine:latest

# Add necessary certificates for potential future features and ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /bin/ramen /usr/local/bin/

# Set the entrypoint to the binary
ENTRYPOINT ["ramen"]

# Default command arguments (can be overridden)
CMD ["-h"]
