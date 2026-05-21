/**
 * Нагрузочный smoke-сценарий Habit Tracker API (k6).
 *
 * Перед прогоном отключите rate limit (иначе будут 429):
 *   В .env: RATE_LIMIT_ENABLED=false
 *   docker compose up -d --build
 *
 * Запуск:
 *   k6 run loadtest/smoke.js
 *   k6 run -e VUS=10 -e DURATION=1m loadtest/smoke.js
 */
import http from "k6/http";
import { check, sleep } from "k6";

const baseUrl = __ENV.BASE_URL || "http://localhost:8080";
const api = `${baseUrl}/api/v1`;

export const options = {
  scenarios: {
    mixed: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "15s", target: Number(__ENV.VUS) || 10 },
        { duration: __ENV.DURATION || "1m", target: Number(__ENV.VUS) || 10 },
        { duration: "10s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<500"],
    "checks{endpoint:health}": ["rate>0.95"],
    "checks{endpoint:me}": ["rate>0.95"],
    "checks{endpoint:dashboard}": ["rate>0.95"],
    "checks{endpoint:habits}": ["rate>0.95"],
  },
};

function jsonHeaders(token) {
  const h = { "Content-Type": "application/json" };
  if (token) {
    h.Authorization = `Bearer ${token}`;
  }
  return h;
}

function checkOK(res, name, tag) {
  return check(
    res,
    {
      [`${name} status 200`]: (r) => r.status === 200,
    },
    { endpoint: tag }
  );
}

export function setup() {
  const email = `load_${Date.now()}@loadtest.local`;
  const password = "secret12";
  const reg = http.post(
    `${api}/auth/register`,
    JSON.stringify({ email, password }),
    { headers: jsonHeaders(), tags: { endpoint: "register" } }
  );
  if (reg.status !== 201) {
    throw new Error(
      `setup register failed: ${reg.status} ${reg.body}. ` +
        "Проверьте API и отключите RATE_LIMIT_ENABLED для load test."
    );
  }
  const body = reg.json();
  return { token: body.access_token };
}

export default function (data) {
  let res = http.get(`${baseUrl}/health`, { tags: { endpoint: "health" } });
  checkOK(res, "health", "health");

  res = http.get(`${api}/me`, { headers: jsonHeaders(data.token), tags: { endpoint: "me" } });
  checkOK(res, "me", "me");

  res = http.get(`${api}/dashboard`, {
    headers: jsonHeaders(data.token),
    tags: { endpoint: "dashboard" },
  });
  checkOK(res, "dashboard", "dashboard");

  res = http.get(`${api}/habits?today=true`, {
    headers: jsonHeaders(data.token),
    tags: { endpoint: "habits" },
  });
  checkOK(res, "habits", "habits");

  // Пауза ~10 итераций/с на VU — не упираться в 120 req/min API limit
  sleep(Number(__ENV.SLEEP) || 1);
}
