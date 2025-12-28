# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy all source code first (needed for local package imports)
COPY . .

# Show files for debugging
RUN ls -la && ls -la static/

# Download dependencies
RUN go mod download

# Build the binary with verbose output
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o proxy .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/proxy .

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Set environment
ENV TZ=Asia/Shanghai

# Run
CMD ["./proxy"]
