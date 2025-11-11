# Build
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o auth-service ./cmd/server

# Install migrate CLI tool
RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz -C /usr/local/bin

# Run
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/auth-service .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY internal/db/migrations ./migrations

EXPOSE 8000
CMD ["sh", "-c", "migrate -path /app/migrations -database $DB_SOURCE up && ./auth-service"]