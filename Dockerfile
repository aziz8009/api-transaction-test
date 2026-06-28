# Dockerfile definition for Backend application service.

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build our binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy OpenAPI spec (optional)
# COPY api.yaml .

# Expose port (sesuai dengan PORT di environment)
EXPOSE 8080

# Run the application
CMD ["./main"]