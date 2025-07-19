# Build stage
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY api/ ./api/
COPY database/ ./database/
COPY plaid/ ./plaid/

# Build the API binary
RUN GOOS=linux GOARCH=amd64 go build -o main ./api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder stage
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]