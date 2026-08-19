FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

FROM alpine:latest

# Instalar bash y certbot dentro del contenedor para ejecuciones directas
RUN apk add --no-cache bash certbot certbot-nginx nginx curl

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8089

CMD ["./server"]
