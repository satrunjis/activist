import http from "k6/http";
import { check, sleep } from "k6";
import { setupAuth } from "./lib/auth.js";

const baseUrl = __ENV.PERF_BASE_URL || "http://127.0.0.1:8080";
const vus = Number(__ENV.PERF_VUS || 4);
const iterations = Number(__ENV.PERF_ITERATIONS || 80);

http.setResponseCallback(http.expectedStatuses(200));

const endpoints = [
  "/api/v1/auth/session",
  "/api/v1/divisions/tree",
  "/api/v1/search/users?limit=5",
  "/api/v1/eventlog?limit=5",
];

export const options = {
  scenarios: {
    smoke: {
      executor: "shared-iterations",
      vus,
      iterations,
      maxDuration: "45s",
    },
  },
  summaryTrendStats: ["min", "avg", "med", "max", "p(50)", "p(95)", "p(99)"],
};

export function setup() {
  return setupAuth(baseUrl);
}

export default function (auth) {
  const headers = auth && auth.headers ? auth.headers : {};
  for (const endpoint of endpoints) {
    const response = http.get(`${baseUrl}${endpoint}`, {
      headers,
      tags: { endpoint },
      timeout: "5s",
    });

    check(response, {
      "status is 200": (r) => r.status === 200,
    });
  }

  sleep(0.2);
}
