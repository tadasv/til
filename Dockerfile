# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

RUN CGO_ENABLED=0 go run build-db.go record.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o til-server til-server.go record.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/til-server .
COPY --from=builder /app/tils.db .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./til-server"]
