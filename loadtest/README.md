# Нагрузочное тестирование (k6)

Скрипты в этой папке. Общие хелперы: [`lib/common.js`](lib/common.js).

## Перед любым perf-прогоном

В `docker-compose.yml` (или `.env` + подстановка в compose):

```env
RATE_LIMIT_ENABLED=false
```

```bash
docker compose up -d --build
```

Для **`auth.js`** (сценарий защиты) — наоборот `RATE_LIMIT_ENABLED=true`.

## Установка k6 (Windows)

```powershell
winget install GrafanaLabs.k6
k6 version
```

## Скрипты

| Файл | Тип | Назначение |
|------|-----|------------|
| [`smoke.js`](smoke.js) | Smoke | Короткий прогон, 4 read-эндпоинта, 10 VU |
| [`load.js`](load.js) | Load | Плато 5–10 мин, взвешенный read-трафик |
| [`stress.js`](stress.js) | Stress | 10 → 25 → 50 → 100 VU, поиск предела |
| [`spike.js`](spike.js) | Spike | Всплеск 2 → 80 → 10 VU |
| [`write.js`](write.js) | Write-heavy | check-in + dashboard + stats |
| [`auth.js`](auth.js) | Auth | Только `POST /auth/login` |

## Запуск

```bash
# Smoke (быстрая проверка)
k6 run loadtest/smoke.js

# Load — основная нагрузка для таблицы в дипломе
k6 run loadtest/load.js
k6 run -e VUS=30 -e DURATION=5m -e SLEEP=0.5 loadtest/load.js

# Stress — график VU vs p95 (смотреть вывод по стадиям)
k6 run loadtest/stress.js
k6 run -e STAGE_DURATION=3m -e MAX_VU=100 loadtest/stress.js

# Spike
k6 run loadtest/spike.js
k6 run -e SPIKE_VU=80 loadtest/spike.js

# Write-heavy (сравнить p95 с smoke)
k6 run loadtest/write.js
k6 run -e VUS=15 -e DURATION=3m loadtest/write.js

# Auth — perf (rate limit выкл.)
k6 run loadtest/auth.js

# Auth — демонстрация 429 (rate limit вкл. + пересборка)
k6 run -e EXPECT_RATE_LIMIT=true -e VUS=10 -e DURATION=1m loadtest/auth.js
```

## Переменные окружения

| Переменная | По умолчанию | Скрипты |
|------------|--------------|---------|
| `BASE_URL` | `http://localhost:8080` | все |
| `VUS` | см. скрипт | load, write, auth |
| `DURATION` | см. скрипт | load, write, auth |
| `SLEEP` | 0.3–1 с | все с `sleep` |
| `RAMP_UP` | `30s` | load |
| `STAGE_DURATION` | `2m30s` | stress |
| `MAX_VU` | `100` | stress |
| `SPIKE_VU` | `80` | spike |
| `PLATEAU_VU` | `10` | spike |
| `EXPECT_RATE_LIMIT` | — | auth (`true` для сценария 429) |

## Load (`load.js`) — веса трафика

| Доля | Эндпоинт |
|------|----------|
| 40% | `GET /habits?today=true` |
| 25% | `GET /dashboard` |
| 20% | `GET /me` |
| 10% | `GET /achievements` |
| 5% | `GET /health` |

В `setup`: по одному пользователю и **3 привычки** на каждый VU.

Пороги: ошибки &lt; 5%, p50/p95/p99, checks &gt; 95% по эндпоинтам.

## Почему раньше падал smoke с rate limit

При `RATE_LIMIT_ENABLED=true`:

| Лимит | Значение | Эффект |
|-------|----------|--------|
| Auth | 10 req/min, burst 3 | 429 на `/auth/*` |
| API | 120 req/min, burst 30 | 429 на `/api/v1/*` |

`/health` без лимита — остаётся зелёным. p95 ≈ 1 ms — типично для 429.

## Интерпретация для диплома

**Таблица прогонов** (пример колонок):

| Сценарий | VU | Длительность | Rate limit | RPS | p50 | p95 | p99 | Ошибки |
|----------|-----|--------------|------------|-----|-----|-----|-----|--------|
| Smoke | 10 | 1m25s | off | … | … | … | … | 0% |
| Load | 30 | 5m | off | … | … | … | … | 0% |
| Stress | 10→100 | ~12m | off | … | … | растёт на высоких VU | … | … |
| Spike | 2→80→10 | ~5m | off | … | … | до/после пика | … | … |
| Write | 15 | 3m | off | … | … | выше smoke | … | 0% |
| Auth | 10 | 1m | off / on | … | … | … | … | 0% / много 429 |

**Выводы для текста:**

- **Load:** «При нормальной нагрузке (30 VU, 5 мин) API удерживает p95 &lt; 500 ms, ошибок нет».
- **Stress:** «До N VU p95 в пределах SLO; при N+1 растёт latency / появляются 5xx».
- **Spike:** «После всплеска до 80 VU p95 восстанавливается за …».
- **Write vs smoke:** «Write-heavy p95 в X раз выше read-only».
- **Auth + rate limit:** «При 10 req/min доля 429 &gt; 30% — защита от brute-force».

Экспорт JSON для графиков:

```bash
k6 run --out json=loadtest/results/load.json loadtest/load.js
```
