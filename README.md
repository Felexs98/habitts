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
  middleware/         — JWT, CORS, rate limit
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

## Аутентификация (access + refresh)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/register`, `/auth/login` | Возвращают `access_token`, `refresh_token`, `expires_in`, `user` |
| POST | `/auth/refresh` | Тело: `{"refresh_token":"..."}` — ротация refresh, новая пара токенов |
| GET | `/me` | Профиль (JWT access) |
| PATCH | `/me` | `email`, `display_name`, `password` + `current_password` при смене пароля |

По умолчанию: access **15m**, refresh **168h** (`JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`). Refresh хранится в БД по `jti` (отзыв при ротации).

## CI (GitHub Actions)

При push/PR в `main` или `master` запускается workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml):

1. **golangci-lint**
2. `go test -v -count=1 ./...` (unit)
3. `go build -o bin/api ./cmd/api`
4. `go test -tags=integration ./test/integration/...` (testcontainers + Postgres)

Версия Go берётся из `go.mod` (`actions/setup-go` + `go-version-file`). Для отчёта по диплому: вкладка **Actions** в репозитории GitHub → зелёный прогон после `git push`.

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

## Rate limiting

Защита от перегрузки и brute-force на `POST /auth/login` и `POST /auth/register`: middleware на базе **token bucket** (`golang.org/x/time/rate`), лимит **по IP** (`ClientIP`).

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `RATE_LIMIT_ENABLED` | `true` | Выкл. для локальной отладки: `false` |
| `RATE_LIMIT_AUTH_RPM` | `10` | Запросов/мин на IP для `/api/v1/auth/*` |
| `RATE_LIMIT_AUTH_BURST` | `3` | Кратковременный всплеск для auth |
| `RATE_LIMIT_API_RPM` | `120` | Запросов/мин на IP для всего `/api/v1` |
| `RATE_LIMIT_API_BURST` | `30` | Burst для API |

При превышении: **HTTP 429** `{"error":"too many requests"}`. На auth действуют **оба** лимита (API + ужесточённый auth).

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

## Миграции БД (golang-migrate)

SQL-файлы в [`migrations/`](migrations/). При старте API выполняется `migrate up` (путь: `MIGRATIONS_PATH`, по умолчанию `migrations`). `AutoMigrate` отключён.

Локально вручную (при установленном [migrate CLI](https://github.com/golang-migrate/migrate)):

```bash
migrate -path migrations -database "postgres://habit:habit_secret@localhost:5432/habit_tracker?sslmode=disable" up
```

## PostgreSQL: индексы и транзакции

**Составные индексы (GORM, `AutoMigrate`):**

| Таблица | Индекс | Запрос |
|---------|--------|--------|
| `habits` | `(user_id, created_at)` | Список привычек пользователя |
| `habits` | `(user_id, deleted_at)` | Активные / архив |
| `check_ins` | `(habit_id, date)` UNIQUE | Отметка и история по привычке |
| `check_ins` | `(date, …)` | «Уже отмечено сегодня» по пользователю (JOIN) |
| `user_achievements` | `(user_id, achievement_id)` UNIQUE | Выдача ачивок |

**Транзакция check-in:** `CheckInRepository.Apply` — в одном `db.Transaction`: insert `check_ins`, update `habits` (streak), update `users` (XP/level), insert `user_achievements` (`ON CONFLICT DO NOTHING`). Бизнес-правила остаются в usecase; запись в БД — атомарно.

После изменения тегов индексов: `docker compose up --build` (пересоздание схемы на dev) или отдельные SQL-миграции на prod.

## Архив и автоочистка

- `DELETE /habits/:id` — soft delete (архив)
- `POST /habits/:id/restore` — восстановление
- Фоновая **автоочистка**: привычки в архиве дольше `ARCHIVE_PURGE_DAYS` (по умолчанию 7) удаляются из БД навсегда (`ARCHIVE_PURGE_INTERVAL`, по умолчанию 24h)

## Дополнительные эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/logout` | `{"refresh_token":"..."}` — отзыв refresh |
| GET | `/habits/:id/check-ins?from=&to=` | История за период (`YYYY-MM-DD`) |

### Ачивки (каталог)

`first_check_in`, `check_ins_10`, `check_ins_50`, `level_3`, `level_5`, `streak_7`, `streak_30`, `habits_5`

## Следующие шаги для диплома

- [x] Геймификация (XP, уровни, achievements, dashboard)
- [x] Unit-тесты usecase-слоя
- [x] Составные индексы + транзакция check-in
- [x] SQL-миграции (golang-migrate)
- [x] CI (lint, unit test, build, integration)
- [x] Access + refresh JWT, PATCH /me
- [x] Rate limiting (token bucket, auth + API)
- [x] Интеграционные тесты (testcontainers)
- [x] Logout, check-in from/to, автоочистка архива, расширенные ачивки
