# Используем официальный образ Go в качестве базового образа
FROM golang:1.21 as builder
LABEL authors="alesande"

RUN mkdir -p /app
# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myServer ./server

RUN cd server && ls
# RUN go build -o myServer ./server

FROM alpine:latest as final

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache postgresql-client

RUN mkdir -p /app

COPY --from=builder /app/myServer /app/myServer

RUN cd app && ls -l

# Делаем файл исполняемым
RUN chmod +x /app/myServer

WORKDIR /app

# Настройка порта, который будет использоваться
EXPOSE 8080

#ENTRYPOINT ["./myServer"]
# Определение точки входа для контейнера
#RUN sh entrypoint.sh
#ENTRYPOINT ["top", "-b"]
