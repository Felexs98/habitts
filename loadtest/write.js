/**
 * Write-heavy — check-in, dashboard, stats (тяжелее read-only smoke).
 *
 *   k6 run loadtest/write.js
 *   k6 run -e VUS=15 -e DURATION=3m loadtest/write.js
 *
 * У каждого VU свой пользователь и 3 привычки; даты check-in уникальны (VU + итерация).
 * RATE_LIMIT_ENABLED=false
 */
import http from "k6/http";
import { sleep } from "k6";
import {
  api,
  jsonHeaders,
  checkStatus,
  setupUsersWithHabits,
  userForVU,
  uniqueCheckInDate,
} from "./lib/common.js";

const vus = Number(__ENV.VUS) || 15;
const duration = __ENV.DURATION || "3m";

export const options = {
  scenarios: {
    write: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "20s", target: vus },
        { duration: duration, target: vus },
        { duration: "20s", target: 0 },
      ],
      gracefulRampDown: "20s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<1000"],
    "checks{endpoint:check_in}": ["rate>0.95"],
    "checks{endpoint:dashboard}": ["rate>0.95"],
    "checks{endpoint:stats}": ["rate>0.95"],
  },
};

export function setup() {
  return setupUsersWithHabits(vus, 3);
}

export default function (data) {
  const user = userForVU(data);
  if (!user || !user.habitIds || user.habitIds.length === 0) {
    return;
  }

  const habitId = user.habitIds[__ITER % user.habitIds.length];
  const date = uniqueCheckInDate(__VU, __ITER);

  let res = http.post(
    `${api}/habits/${habitId}/check-ins`,
    JSON.stringify({ date }),
    { headers: jsonHeaders(user.token), tags: { endpoint: "check_in" } }
  );
  checkStatus(res, "check_in", "check_in", 201);

  res = http.get(`${api}/dashboard`, {
    headers: jsonHeaders(user.token),
    tags: { endpoint: "dashboard" },
  });
  checkStatus(res, "dashboard", "dashboard");

  res = http.get(`${api}/habits/${habitId}/stats`, {
    headers: jsonHeaders(user.token),
    tags: { endpoint: "stats" },
  });
  checkStatus(res, "stats", "stats");

  sleep(Number(__ENV.SLEEP) || 0.5);
}
