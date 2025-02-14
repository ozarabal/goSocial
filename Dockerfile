# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install curl dan build tools
RUN apk add --no-cache curl make

# Download dan install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Set working directory
WORKDIR /app

# Copy module files dan download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode proyek
COPY . .

# Build aplikasi
RUN go build -o /app/bin/main ./cmd/api/main.go

# Stage 2: Runtime
FROM golang:1.23-alpine

# Copy migrate dari builder stage
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy aplikasi dari builder stage
COPY --from=builder /app/bin/main /app/bin/main

# Pastikan direktori kerja
WORKDIR /app

# Install alat bantu
RUN apk add --no-cache make

# Jalankan migrate sebelum aplikasi dimulai
ENTRYPOINT ["sh", "-c", "make migrate-up && /app/bin/main"]