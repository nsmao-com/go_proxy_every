# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files first
COPY go.mod go.sum ./

# Copy all source code
COPY auth/ ./auth/
COPY config/ ./config/
COPY handlers/ ./handlers/
COPY proxy/ ./proxy/
COPY static/ ./static/
COPY main.go ./

# Download dependencies
RUN go mod download

# Build the binary (use 'server' to avoid conflict with proxy/ directory)
RUN CGO_ENABLED=0 GOOS=linux go build -o server . && \
    ls -la /app/server && \
    chmod +x /app/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server /app/server

RUN chmod +x /app/server && ls -la /app/server

RUN mkdir -p /app/data

EXPOSE 8080

ENV TZ=Asia/Shanghai

CMD ["/app/server"]
