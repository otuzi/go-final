# Этап 1: сборка статического бинарника.
FROM golang:1.24 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo . \
    && CGO_ENABLED=0 GOOS=linux go build -o /seed ./cmd/seed

# Этап 2: минимальный образ для запуска.
FROM ubuntu:latest

WORKDIR /app
COPY --from=build /todo /app/todo
COPY --from=build /seed /app/seed
COPY web /app/web

# Переменные окружения, которые понимает сервер:
#   TODO_PORT         — порт веб-сервера;
#   TODO_DBFILE       — путь к файлу базы данных SQLite;
#   TODO_POSTGRES_DSN — строка подключения к PostgreSQL; если задана,
#                       SQLite не используется (так работает docker compose);
#   TODO_PASSWORD     — пароль для входа; задаётся при запуске через
#                       docker run -e TODO_PASSWORD=..., по умолчанию выключен.
ENV TODO_PORT=7540 \
    TODO_DBFILE=/data/scheduler.db

# Директория для базы данных: монтируется с хоста через -v.
VOLUME ["/data"]
EXPOSE 7540

CMD ["/app/todo"]
