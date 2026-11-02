# Build stage
FROM golang:1.22-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Generate go.sum and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy binary
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
