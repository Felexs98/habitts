/**
 * Auth load — только POST /auth/login.
 *
 * Сценарий A (perf): RATE_LIMIT_ENABLED=false
 *   k6 run loadtest/auth.js
 *
 * Сценарий B (защита): RATE_LIMIT_ENABLED=true в docker-compose
 *   k6 run -e EXPECT_RATE_LIMIT=true loadtest/auth.js
 */
import http from "k6/http";
import { check, sleep } from "k6";
import { api, jsonHeaders, setupUsersWithHabits, userForVU } from "./lib/common.js";

const vus = Number(__ENV.VUS) || 10;
const duration = __ENV.DURATION || "1m";
const expectRateLimit = __ENV.EXPECT_RATE_LIMIT === "true";

export const options = {
  scenarios: {
    auth: {
      executor: "constant-vus",
      vus: vus,
      duration: duration,
    },
  },
  thresholds: expectRateLimit
    ? {
        "checks{endpoint:login,result:rate_limited}": ["rate>0.30"],
        http_req_duration: ["p(95)<500"],
      }
    : {
        http_req_failed: ["rate<0.05"],
        http_req_duration: ["p(95)<500"],
        "checks{endpoint:login,result:ok}": ["rate>0.95"],
      },
};

export function setup() {
  return setupUsersWithHabits(vus, 0);
}

export default function (data) {
  const user = userForVU(data);
  if (!user) {
    return;
  }

  const res = http.post(
    `${api}/auth/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { headers: jsonHeaders(), tags: { endpoint: "login" } }
  );

  if (expectRateLimit) {
    const tag = res.status === 429 ? "rate_limited" : "ok";
    check(
      res,
      {
        "login rate limited (429)": (r) => r.status === 429,
      },
      { endpoint: "login", result: tag }
    );
  } else {
    check(
      res,
      {
        "login status 200": (r) => r.status === 200,
      },
      { endpoint: "login", result: "ok" }
    );
  }

  sleep(Number(__ENV.SLEEP) || 0.5);
}
