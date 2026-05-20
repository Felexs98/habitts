# Postman — Habit Tracker API

## Swagger

После изменения API: `swag init -g cmd/api/main.go -o docs --parseInternal` (или `make swagger`), затем `docker compose up --build -d`.

UI: http://localhost:8080/swagger/index.html → **Authorize** → `Bearer <token>`.

## Импорт

1. Запустите API: `docker compose up -d`
2. Postman → **Import** → выберите оба файла:
   - `Habit-Tracker-API.postman_collection.json`
   - `Habit-Tracker-Local.postman_environment.json`
3. В правом верхнем углу выберите environment **Habit Tracker - Local**

## Порядок запросов

1. **Health** → Health Check  
2. **Auth** → Register *(или Login)*  
3. **Auth** → Get Me  
4. **Auth** → Dashboard — сводка XP, today, streaks  
5. **Habits** → Create Habit → List Check-ins (history) → Check-in (today) → Habit Stats  
5. **Check-in Duplicate** — ожидается `409`  

### Архив (soft delete)

6. **Archive Habit (soft delete)** — `DELETE`, 204  
7. **List Habits** — активных нет  
8. **List Habits (archived)** — `?archived=true`, видна привычка с `deleted_at`  
9. **Restore Habit** — `POST .../habits/{{habitId}}/restore`, снова в активных  

После **Register** / **Login** в environment подставляется `token`.  
После **Create Habit** — `habitId`.

## Переменные environment

| Переменная    | Назначение                          |
|---------------|-------------------------------------|
| `baseUrl`     | `http://localhost:8080`           |
| `email`       | email для register/login            |
| `password`    | минимум 6 символов                  |
| `token`       | заполняется автоматически           |
| `habitId`     | заполняется после Create Habit      |
| `checkInDate` | `YYYY-MM-DD` для check-in по дате   |
