FROM golang:1.23-alpine AS builder

# Install curl dan build tools
RUN apk add --no-cache curl make

# Download dan install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

FROM golang:1.23-alpine

# Copy migrate dari builder stage
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Pastikan direktori kerja
WORKDIR /app

# Install air dan swag
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN apk add --no-cache make

# Copy dependensi
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode proyek
COPY . .

# Jalankan migrate sebelum aplikasi dimulai
CMD make migrate-up && air -c .air.toml