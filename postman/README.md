# Postman — Habit Tracker API

## Swagger

После изменения API: `swag init -g cmd/api/main.go -o docs --parseInternal` (или `make swagger`), затем `docker compose up --build -d`.

UI: http://localhost:8080/swagger/index.html → **Authorize** → `Bearer <access_token>`.

## Импорт

1. Запустите API: `docker compose up -d`
2. Postman → **Import** → выберите оба файла:
   - `Habit-Tracker-API.postman_collection.json`
   - `Habit-Tracker-Local.postman_environment.json`
3. В правом верхнем углу выберите environment **Habit Tracker - Local**

> Если запросов Restore/Refresh нет — удалите старую коллекцию и импортируйте файлы заново.

## Порядок запросов

1. **Health** → Health Check  
2. **Auth** → Register *(или Login)* — сохраняются `token` (access) и `refreshToken`  
3. **Auth** → Get Me  
4. **Auth** → Update Me (PATCH) — `display_name`  
5. **Auth** → Refresh Token — проверка ротации refresh  
6. **Auth** → Dashboard  
7. **Habits** → Create Habit → Check-in → Stats …  

### Архив (soft delete)

- **Archive Habit** → **List Habits (archived)** → **Restore Habit**

После **Register** / **Login** / **Refresh** в environment: `token` = access JWT, `refreshToken` = refresh JWT.  
После **Create Habit** — `habitId`.

## Переменные environment

| Переменная      | Назначение                                      |
|-----------------|-------------------------------------------------|
| `baseUrl`       | `http://localhost:8080`                         |
| `email`         | email для register/login                        |
| `password`      | минимум 6 символов                              |
| `token`         | access JWT (Bearer), заполняется автоматически  |
| `refreshToken`  | refresh JWT, заполняется автоматически          |
| `displayName`   | для PATCH /me                                   |
| `habitId`       | после Create Habit                              |
| `checkInDate`   | `YYYY-MM-DD` для check-in по дате               |

## Auth: формат ответа

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 900,
  "user": { "id": 1, "email": "...", "xp": 0, "level": 1 }
}
```

Старый формат `"token": "..."` больше не используется.
