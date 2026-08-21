# Stage 1: Build static binary with CGO disabled
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS/SSL connections
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary targeting Hexagonal entry point cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Stage 2: Final Minimal Scratch Image (Ultra Light RAM ~5MB - 10MB)
FROM scratch

WORKDIR /app

# Copy SSL certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /app/main /app/main

EXPOSE 8080

ENTRYPOINT ["/app/main"]
