/**
 * Stress test — поиск предела: 10 → 25 → 50 → 100 VU.
 * Пороги только по latency; на высоких VU допустимы ошибки.
 *
 *   k6 run loadtest/stress.js
 *   k6 run -e STAGE_DURATION=3m loadtest/stress.js
 *
 * RATE_LIMIT_ENABLED=false
 */
import { sleep } from "k6";
import { setupUsersWithHabits, userForVU, runWeightedRead } from "./lib/common.js";

const stageDuration = __ENV.STAGE_DURATION || "2m30s";
const maxVu = Number(__ENV.MAX_VU) || 100;

export const options = {
  scenarios: {
    stress: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 10 },
        { duration: stageDuration, target: 10 },
        { duration: "30s", target: 25 },
        { duration: stageDuration, target: 25 },
        { duration: "30s", target: 50 },
        { duration: stageDuration, target: 50 },
        { duration: "30s", target: maxVu },
        { duration: stageDuration, target: maxVu },
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "30s",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<500"],
  },
};

export function setup() {
  return setupUsersWithHabits(maxVu, 2);
}

export default function (data) {
  const user = userForVU(data);
  if (!user) {
    return;
  }
  runWeightedRead(user.token);
  sleep(Number(__ENV.SLEEP) || 0.3);
}
