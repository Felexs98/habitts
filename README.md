# Habit Tracker API (Backend MVP)

Серверная часть веб-сервиса трекера привычек. Дипломный MVP на Go с многослойной архитектурой.

## Стек

- **Go** + **Gin** (REST API)
- **PostgreSQL** + **GORM** (`AutoMigrate`, Foreign Keys)
- **JWT** (stateless auth) + **bcrypt**
- **Swagger** (Swaggo)
- **Docker Compose** (`app` + `postgres`)

## Структура проекта

```
cmd/api/              — точка входа
internal/
  config/             — конфигурация из env
  domain/             — сущности и интерфейсы репозиториев
  repository/         — GORM + PostgreSQL
  usecase/            — бизнес-логика (auth, habits, streak)
  delivery/http/      — HTTP handlers (Gin)
  middleware/         — JWT middleware
pkg/
  hash/               — bcrypt
  jwtutil/            — JWT
docs/                 — Swagger (генерируется swag)
```

## Postman

Коллекция и environment для тестирования API: папка [`postman/`](postman/) — см. [postman/README.md](postman/README.md).

## Быстрый старт (Docker)

```bash
cp .env.example .env
# при необходимости измените JWT_SECRET в .env и docker-compose

docker compose up --build
```

API: http://localhost:8080  
Health: http://localhost:8080/health  
Swagger UI: http://localhost:8080/swagger/index.html

## Локальный запуск (без Docker для app)

1. Поднять только Postgres:

```bash
docker compose up postgres -d
```

2. Скопировать env и запустить API:

```bash
cp .env.example .env
# в .env: DB_HOST=localhost

go mod tidy
go run ./cmd/api
```

## API (кратко)

| Метод | Путь | Auth |
|-------|------|------|
| POST | `/api/v1/auth/register` | — |
| POST | `/api/v1/auth/login` | — |
| GET/POST | `/api/v1/habits` | Bearer JWT |
| GET/PUT/DELETE | `/api/v1/habits/:id` | Bearer JWT |
| POST | `/api/v1/habits/:id/check-ins` | Bearer JWT |
| GET | `/api/v1/habits/:id/stats` | Bearer JWT |

### Пример

```bash
# Регистрация
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret12"}'

# Создать привычку (подставьте token из ответа login/register)
curl -X POST http://localhost:8080/api/v1/habits \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Читать 30 мин","description":"Каждый день"}'
```

## Swagger

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger
# или: swag init -g cmd/api/main.go -o docs
```

## Следующие шаги для диплома

- [ ] Расширить геймификацию (очки, уровни, achievements)
- [ ] Unit-тесты usecase-слоя
- [ ] Миграции (опционально, вместо только AutoMigrate)
- [ ] CI (lint, test, build)
