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

FROM alpine:latest as final

RUN apk --no-cache add ca-certificates
FROM cr.yandex/crp6prgfic4t20er8gcr/pgstart:1c372b3c022ed72fba041ef5595fe240b0856e77
RUN apk add --no-cache postgresql-client

COPY --from=builder /app/myapp .

# Настройка порта, который будет использоваться
EXPOSE 8080

# Определение точки входа для контейнера
#RUN sh entrypoint.sh
#ENTRYPOINT ["top", "-b"]