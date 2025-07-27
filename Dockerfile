# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for Go modules
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o teleflix ./cmd/main.go

# Final stage
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/teleflix .

# Copy default config
COPY --from=builder /app/config.yaml .

# Create manifests directory
RUN mkdir -p /manifests

# Set default command
ENTRYPOINT ["./teleflix"]
CMD ["--output", "/manifests"]