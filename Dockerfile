# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git build-base

# Solo módulos primero (cache de deps)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar binario estático para la aplicación web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./src/lambda/binService

# Final stage
FROM alpine:3.20

WORKDIR /root/
# Certificados para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copiar binario
COPY --from=builder /app/main .

EXPOSE 8080

ENV APP_PORT=8080

CMD ["./main"]
