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
RUN make buildCustom

RUN apk add --no-cache postgresql-client

# Настройка порта, который будет использоваться
EXPOSE 8080

FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Определение точки входа для контейнера
#RUN sh entrypoint.sh
#ENTRYPOINT ["top", "-b"]