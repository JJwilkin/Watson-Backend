# Build stage
FROM golang:1.23.10-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY background-worker/ ./background-worker/
COPY database/ ./database/
COPY plaid/ ./plaid/

# Build the worker binary
RUN GOOS=linux GOARCH=amd64 go build -o worker ./background-worker

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder stage
COPY --from=builder /app/worker .

EXPOSE 8081

CMD ["./worker"] 