# Subscription Service

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Требования

- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 15+

## Запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/sssseraphim/effective_mobile
cd effective_mobile
```
2. Скопируйте файл с переменными окружения:

```bash
cp .env.example .env
```

3. Запустите сервис с помощью Docker Compose:

```bash
docker-compose up -d
```
Сервис будет доступен по адресу: http://localhost:8080

Swagger документация по адресу: http://localhost:8080/swagger/index.html
