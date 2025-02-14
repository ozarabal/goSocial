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

# Install Swag untuk Swagger Docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger Docs
RUN make gen-docs

# Tahap akhir
FROM golang:1.23-alpine

# Pastikan direktori kerja
WORKDIR /app

# Install make
RUN apk add --no-cache make


# Copy migrate dari builder stage
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy file yang dibutuhkan untuk pre cmd
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/cmd/migrate/migrations /app/cmd/migrate/migrations
COPY --from=builder /app/makefile /app/makefile
COPY --from=builder /app/docs /app/docs

# Berikan permission agar bisa dieksekusi
RUN chmod +x /app/bin/main

# Expose port aplikasi
EXPOSE 3000

# Jalankan migrate sebelum aplikasi dimulai
CMD make migrate-up && ./bin/main