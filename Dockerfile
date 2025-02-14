FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache curl make

# Install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Pastikan direktori kerja
WORKDIR /app

# Copy dependensi
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode proyek
COPY . .

# Build aplikasi
RUN go build -o bin/main ./cmd/api

# Tahap akhir
FROM golang:1.23-alpine

# Pastikan direktori kerja
WORKDIR /app

# Install make
RUN apk add --no-cache make

# Copy migrate dari builder stage
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy file yang dibutuhkan
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/Makefile /app/Makefile

# Berikan permission agar bisa dieksekusi
RUN chmod +x /app/bin/main

# Expose port aplikasi
EXPOSE 3000

# Jalankan migrate sebelum aplikasi dimulai
CMD make migrate && ./bin/main