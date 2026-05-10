import http from "k6/http";
import { check, sleep } from "k6";
import { setupAuth } from "./lib/auth.js";

const baseUrl = __ENV.PERF_BASE_URL || "http://127.0.0.1:8080";
const vus = Number(__ENV.PERF_VUS || 4);
const iterations = Number(__ENV.PERF_ITERATIONS || 50);

http.setResponseCallback(http.expectedStatuses(200));

export const options = {
  scenarios: {
    read_tree: {
      executor: "shared-iterations",
      exec: "readTree",
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

export function readTree(auth) {
  const headers = auth && auth.headers ? auth.headers : {};
  const response = http.get(`${baseUrl}/api/v1/divisions/tree`, {
    headers,
    tags: { endpoint: "/api/v1/divisions/tree" },
    timeout: "5s",
  });

  check(response, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(0.2);
}
