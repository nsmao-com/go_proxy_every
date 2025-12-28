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

# Download dependencies and build
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o proxy .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/proxy .

RUN mkdir -p /app/data

EXPOSE 8080

ENV TZ=Asia/Shanghai

CMD ["./proxy"]
