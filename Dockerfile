FROM golang:1.23-alpine AS builder

# Install curl dan build tools
RUN apk add --no-cache curl make

# Download dan install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Set working directory
WORKDIR /app

# Copy go mod dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project
COPY . .

# Build binary
RUN go build -o bin/main ./cmd/api

# Final stage
FROM golang:1.23-alpine

# Set working directory
WORKDIR /app

# Copy migrate from builder
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy built binary from builder
COPY --from=builder /app/bin/main /app/bin/main

# Set executable permissions
RUN chmod +x /app/bin/main

# Expose the application port
EXPOSE 3000

# Run migrations and start the application
CMD make migrate-up && /app/bin/main
