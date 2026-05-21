/**
 * Spike test — резкий всплеск: 2 → 80 → 10 VU.
 *
 *   k6 run loadtest/spike.js
 *
 * RATE_LIMIT_ENABLED=false
 */
import { sleep } from "k6";
import { setupUsersWithHabits, userForVU, runReadSequence } from "./lib/common.js";

const spikeVu = Number(__ENV.SPIKE_VU) || 80;
const plateauVu = Number(__ENV.PLATEAU_VU) || 10;

export const options = {
  scenarios: {
    spike: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 2 },
        { duration: "30s", target: spikeVu },
        { duration: "2m", target: spikeVu },
        { duration: "30s", target: plateauVu },
        { duration: "1m", target: plateauVu },
        { duration: "20s", target: 0 },
      ],
      gracefulRampDown: "20s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.10"],
    http_req_duration: ["p(95)<1000"],
    "checks{endpoint:health}": ["rate>0.95"],
    "checks{endpoint:me}": ["rate>0.90"],
    "checks{endpoint:dashboard}": ["rate>0.90"],
    "checks{endpoint:habits}": ["rate>0.90"],
  },
};

export function setup() {
  return setupUsersWithHabits(spikeVu, 2);
}

export default function (data) {
  const user = userForVU(data);
  if (!user) {
    return;
  }
  runReadSequence(user.token);
  sleep(Number(__ENV.SLEEP) || 0.5);
}
