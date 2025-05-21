FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o social-network-backend .

# Create a minimal production image
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates sqlite

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/social-network-backend .

# Copy migrations
COPY --from=builder /app/pkg/db/migrations /app/pkg/db/migrations

# Create directories for uploads and database
RUN mkdir -p /app/uploads/avatars /app/uploads/posts /app/data

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/app/data/social_network.db
ENV MIGRATIONS_PATH=/app/pkg/db/migrations/sqlite

# Expose port
EXPOSE 8080

# Run the application
CMD ["./social-network-backend", "-port", "8080", "-db", "/app/data/social_network.db", "-migrations", "/app/pkg/db/migrations/sqlite"]
