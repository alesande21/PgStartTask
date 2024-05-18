# Используем официальный образ Go в качестве базового образа
FROM golang:1.21 as builder
LABEL authors="alesande"

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Сборка приложения
RUN go build -o myServer ./server

FROM alpine:latest as final

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache postgresql-client

COPY --from=builder /app/myServer .

# Настройка порта, который будет использоваться
EXPOSE 8080

#ENTRYPOINT ["./myServer"]
# Определение точки входа для контейнера
#RUN sh entrypoint.sh
#ENTRYPOINT ["top", "-b"]