# Используем официальный образ Go
FROM golang:1.22.1-alpine

WORKDIR /app

# Копируем go.mod и go.sum, чтобы закэшировать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o router main.go

# Запускаем
CMD ["./router"]
