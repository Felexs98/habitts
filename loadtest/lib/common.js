import http from "k6/http";
import { check } from "k6";

export const baseUrl = __ENV.BASE_URL || "http://localhost:8080";
export const api = `${baseUrl}/api/v1`;
export const defaultPassword = "secret12";

export function jsonHeaders(token) {
  const h = { "Content-Type": "application/json" };
  if (token) {
    h.Authorization = `Bearer ${token}`;
  }
  return h;
}

export function checkStatus(res, name, tag, expected = 200) {
  return check(
    res,
    {
      [`${name} status ${expected}`]: (r) => r.status === expected,
    },
    { endpoint: tag }
  );
}

/** Регистрация пользователя для нагрузочных тестов. */
export function registerUser(suffix) {
  const email = `load_${suffix}_${Date.now()}@loadtest.local`;
  const reg = http.post(
    `${api}/auth/register`,
    JSON.stringify({ email, password: defaultPassword }),
    { headers: jsonHeaders(), tags: { endpoint: "register" } }
  );
  if (reg.status !== 201) {
    throw new Error(
      `register failed (${suffix}): ${reg.status} ${reg.body}. ` +
        "Проверьте API; для perf-тестов: RATE_LIMIT_ENABLED=false и docker compose up -d --build."
    );
  }
  const body = reg.json();
  return { email, password: defaultPassword, token: body.access_token };
}

/** Создаёт count ежедневных привычек, возвращает массив id. */
export function createHabits(token, count, prefix = "habit") {
  const ids = [];
  for (let i = 0; i < count; i++) {
    const res = http.post(
      `${api}/habits`,
      JSON.stringify({
        title: `${prefix}_${i}`,
        description: "load test",
        frequency: "daily",
      }),
      { headers: jsonHeaders(token), tags: { endpoint: "habits_create" } }
    );
    if (res.status !== 201) {
      throw new Error(`create habit failed: ${res.status} ${res.body}`);
    }
    ids.push(res.json().id);
  }
  return ids;
}

/** Пул пользователей с привычками: по одному на каждый VU (1..maxVu). */
export function setupUsersWithHabits(maxVu, habitsPerUser = 3) {
  const users = [];
  const ts = Date.now();
  for (let vu = 1; vu <= maxVu; vu++) {
    const u = registerUser(`vu${vu}_${ts}`);
    u.habitIds = createHabits(u.token, habitsPerUser, `vu${vu}`);
    users.push(u);
  }
  return { users };
}

export function userForVU(data) {
  if (!data.users || data.users.length === 0) {
    return null;
  }
  const idx = (__VU - 1) % data.users.length;
  return data.users[idx];
}

/** Уникальная дата YYYY-MM-DD для пары (VU, итерация). */
export function uniqueCheckInDate(vu, iter) {
  const offset = vu * 50000 + iter;
  const d = new Date(Date.UTC(2020, 0, 1));
  d.setUTCDate(d.getUTCDate() + offset);
  return d.toISOString().slice(0, 10);
}

/**
 * Случайный read-эндпоинт по весам (%):
 * habits 40, dashboard 25, me 20, achievements 10, health 5
 */
export function runWeightedRead(token) {
  const r = Math.random() * 100;
  let acc = 0;

  acc += 40;
  if (r < acc) {
    const res = http.get(`${api}/habits?today=true`, {
      headers: jsonHeaders(token),
      tags: { endpoint: "habits" },
    });
    return checkStatus(res, "habits", "habits");
  }

  acc += 25;
  if (r < acc) {
    const res = http.get(`${api}/dashboard`, {
      headers: jsonHeaders(token),
      tags: { endpoint: "dashboard" },
    });
    return checkStatus(res, "dashboard", "dashboard");
  }

  acc += 20;
  if (r < acc) {
    const res = http.get(`${api}/me`, {
      headers: jsonHeaders(token),
      tags: { endpoint: "me" },
    });
    return checkStatus(res, "me", "me");
  }

  acc += 10;
  if (r < acc) {
    const res = http.get(`${api}/achievements`, {
      headers: jsonHeaders(token),
      tags: { endpoint: "achievements" },
    });
    return checkStatus(res, "achievements", "achievements");
  }

  const res = http.get(`${baseUrl}/health`, { tags: { endpoint: "health" } });
  return checkStatus(res, "health", "health");
}

/** Тот же read-набор, что smoke (последовательно). */
export function runReadSequence(token) {
  let res = http.get(`${baseUrl}/health`, { tags: { endpoint: "health" } });
  checkStatus(res, "health", "health");

  res = http.get(`${api}/me`, { headers: jsonHeaders(token), tags: { endpoint: "me" } });
  checkStatus(res, "me", "me");

  res = http.get(`${api}/dashboard`, {
    headers: jsonHeaders(token),
    tags: { endpoint: "dashboard" },
  });
  checkStatus(res, "dashboard", "dashboard");

  res = http.get(`${api}/habits?today=true`, {
    headers: jsonHeaders(token),
    tags: { endpoint: "habits" },
  });
  checkStatus(res, "habits", "habits");
}
