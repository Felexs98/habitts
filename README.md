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

## Тесты

```bash
go test ./... -v
# или
make test
```

Покрыты `internal/domain` (уровень от XP) и `internal/usecase` (auth, check-in, streak) с моками репозиториев.

## Периодичность привычек

| `frequency` | Поведение |
|-------------|-----------|
| `daily` | Нужно выполнять каждый день |
| `weekly` | Нужно выполнять в выбранный день недели (`weekday`: 0=вс … 6=сб) |

При создании `weekly` без `weekday` подставляется текущий день недели.

**Фильтр «на сегодня»:** `GET /api/v1/habits?today=true` — daily + weekly на сегодня, **исключая** уже отмеченные check-in за текущие сутки.

## Архив (soft delete)

`DELETE` помечает `deleted_at` (GORM soft delete). Архив: `GET /habits?archived=true`. Восстановление: `POST /habits/:id/restore`. К архивным нельзя сделать check-in.

## Геймификация

- При регистрации: `xp = 0`, `level = 1`
- Каждый check-in: **+10 XP**
- Уровень: `level = xp / 100 + 1` (каждые 100 XP — новый уровень)
- `GET /api/v1/me` — профиль с прогрессом
- Ответ check-in включает `xp_awarded`, `total_xp`, `level`, `current_streak`, `max_streak`, `new_achievements`
- У привычки в БД: `current_streak`, `max_streak` (обновляются при check-in)
- `GET /api/v1/achievements` — каталог достижений и статус `earned`
- Ачивки: `first_check_in`, `check_ins_10`, `level_3`, `streak_7` (выдаются при check-in)

## CORS (веб-фронт)

Браузерные клиенты (React, Vue на `localhost:3000` / `:5173`) могут вызывать API. Настройка через `CORS_ORIGINS` в `.env` (см. `.env.example`).

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
| GET | `/api/v1/me` | Bearer JWT (xp, level) |
| GET | `/api/v1/dashboard` | Сводка: XP, level, today, streaks |
| GET | `/api/v1/habits/:id/check-ins` | История логов привычки |
| GET/POST | `/api/v1/habits` | Bearer JWT |
| GET | `/api/v1/habits?today=true` | Привычки «на сегодня» (не отмеченные) |
| GET | `/api/v1/habits?archived=true` | Архив (soft delete) |
| DELETE | `/api/v1/habits/:id` | В архив (не удаляет из БД) |
| POST | `/api/v1/habits/:id/restore` | Восстановить из архива |
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
# или: swag init -g cmd/api/main.go -o docs --parseInternal
docker compose up --build -d   # пересборка, чтобы UI подхватил docs
```

Если в UI «No operations defined» — перегенерируйте docs и пересоберите контейнер `app`.

## Следующие шаги для диплома

- [ ] Расширить геймификацию (очки, уровни, achievements)
- [ ] Unit-тесты usecase-слоя
- [ ] Миграции (опционально, вместо только AutoMigrate)
- [ ] CI (lint, test, build)
