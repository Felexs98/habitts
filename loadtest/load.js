/**
 * Load test — устойчивая «рабочая» нагрузка, смешанный read-трафик по весам.
 *
 *   k6 run loadtest/load.js
 *   k6 run -e VUS=30 -e DURATION=5m -e SLEEP=0.5 loadtest/load.js
 *
 * Перед прогоном: RATE_LIMIT_ENABLED=false, docker compose up -d --build
 */
import { sleep } from "k6";
import { setupUsersWithHabits, userForVU, runWeightedRead } from "./lib/common.js";

const vus = Number(__ENV.VUS) || 30;
const duration = __ENV.DURATION || "5m";
const rampUp = __ENV.RAMP_UP || "30s";

export const options = {
  scenarios: {
    load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: rampUp, target: vus },
        { duration: duration, target: vus },
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "30s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(50)<200", "p(95)<500", "p(99)<1000"],
    "checks{endpoint:habits}": ["rate>0.95"],
    "checks{endpoint:dashboard}": ["rate>0.95"],
    "checks{endpoint:me}": ["rate>0.95"],
    "checks{endpoint:achievements}": ["rate>0.95"],
    "checks{endpoint:health}": ["rate>0.95"],
  },
};

export function setup() {
  return setupUsersWithHabits(vus, 3);
}

export default function (data) {
  const user = userForVU(data);
  if (!user) {
    return;
  }
  runWeightedRead(user.token);
  sleep(Number(__ENV.SLEEP) || 0.5);
}
