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
2. **Auth** → Register *(или Login, если пользователь уже есть)*  
3. **Auth** → Get Me — `xp: 0`, `level: 1`  
4. **Habits** → Create Habit → Check-in (today) — в ответе `xp_awarded`, `total_xp`, `level`  
5. **Auth** → Get Me снова — XP вырос  
6. **Habits** → Habit Stats  
4. **Check-in Duplicate** — ожидается `409`  
5. При необходимости → Delete Habit  

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
