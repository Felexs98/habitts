# Postman — Habit Tracker API

## Импорт

1. Запустите API: `docker compose up -d`
2. Postman → **Import** → выберите оба файла:
   - `Habit-Tracker-API.postman_collection.json`
   - `Habit-Tracker-Local.postman_environment.json`
3. В правом верхнем углу выберите environment **Habit Tracker - Local**

## Порядок запросов

1. **Health** → Health Check  
2. **Auth** → Register *(или Login, если пользователь уже есть)*  
3. **Habits** → Create Habit → Check-in (today) → Habit Stats  
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
