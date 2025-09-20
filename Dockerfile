# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git build-base

# Solo módulos primero (cache de deps)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar binario estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./src/lambda/demo

# Final stage
FROM alpine:3.20

WORKDIR /root/
# Certificados y curl para healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# Copiar binario
COPY --from=builder /app/main .

EXPOSE 8080

ENV APP_PORT=8080

CMD ["./main"]
