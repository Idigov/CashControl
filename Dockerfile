# Этап сборки
FROM golang:1.24-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Рабочая директория
WORKDIR /build

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей (кэшируется если go.mod/go.sum не изменились)
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cashcontrol ./cmd/cashcontrol/main.go

# Этап выполнения
FROM alpine:latest

# Установка CA сертификатов и tzdata для работы с HTTPS и временными зонами
RUN apk --no-cache add ca-certificates tzdata

# Создание пользователя для безопасности
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копирование бинарника из этапа сборки
COPY --from=builder /build/cashcontrol .

# Изменение владельца файлов
RUN chown -R appuser:appuser /app

# Переключение на непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Команда запуска
CMD ["./cashcontrol"]

