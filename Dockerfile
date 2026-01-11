FROM golang:1.24-alpine AS builder

WORKDIR /src

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /src/bot .

# Создаём директорию для временных файлов
RUN mkdir -p /tmp/telegram-bot

# Экспортируем порт (если нужен, для future webhook support)
EXPOSE 8080

# Запускаем бота
CMD ["./bot"]

