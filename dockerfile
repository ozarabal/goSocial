# Gunakan multi-stage
FROM golang:1.23-alpine AS builder

# Install curl dan build tools
RUN apk add --no-cache curl

# Download dan install migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

FROM golang:1.23-alpine

COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

WORKDIR /app

RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN apk add --no-cache make

# Copy dependensi
COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]