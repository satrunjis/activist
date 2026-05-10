import http from "k6/http";
import { fail } from "k6";

const DEFAULT_CREDENTIALS = "admin:admin,test0:test0";

function parseCredentials() {
  const explicitLogin = (__ENV.PERF_LOGIN || "").trim();
  const explicitPassword = (__ENV.PERF_PASSWORD || "").trim();
  if (explicitLogin && explicitPassword) {
    return [{ login: explicitLogin, password: explicitPassword }];
  }

  const raw = (__ENV.PERF_AUTH_CREDENTIALS || DEFAULT_CREDENTIALS).trim();
  if (!raw) {
    return [];
  }

  const parsed = [];
  for (const chunk of raw.split(",")) {
    const candidate = chunk.trim();
    if (!candidate) {
      continue;
    }
    const separator = candidate.indexOf(":");
    if (separator <= 0 || separator >= candidate.length - 1) {
      continue;
    }
    const login = candidate.slice(0, separator).trim();
    const password = candidate.slice(separator + 1).trim();
    if (login && password) {
      parsed.push({ login, password });
    }
  }
  return parsed;
}

function buildCookieHeader(responseCookies) {
  const parts = [];
  for (const [name, values] of Object.entries(responseCookies || {})) {
    if (!Array.isArray(values) || values.length === 0) {
      continue;
    }
    const value = values[0] && values[0].value;
    if (typeof value === "string" && value.length > 0) {
      parts.push(`${name}=${value}`);
    }
  }
  return parts.join("; ");
}

export function setupAuth(baseUrl) {
  const skipAuth = (__ENV.PERF_SKIP_AUTH || "").trim().toLowerCase() === "true";
  if (skipAuth) {
    return { headers: {}, login: "anonymous" };
  }

  const credentials = parseCredentials();
  if (credentials.length === 0) {
    fail("Perf auth failed: no credentials provided. Set PERF_LOGIN/PERF_PASSWORD or PERF_AUTH_CREDENTIALS.");
  }

  for (const candidate of credentials) {
    const loginResponse = http.post(
      `${baseUrl}/api/v1/auth/login`,
      JSON.stringify({ login: candidate.login, password: candidate.password }),
      {
        headers: { "Content-Type": "application/json" },
        timeout: "10s",
        tags: { perf_auth: "login" },
        responseCallback: http.expectedStatuses(200, 401, 403, 429),
      }
    );

    if (loginResponse.status !== 200) {
      continue;
    }

    const cookieHeader = buildCookieHeader(loginResponse.cookies);
    if (!cookieHeader) {
      continue;
    }

    const sessionResponse = http.get(`${baseUrl}/api/v1/auth/session`, {
      headers: { Cookie: cookieHeader },
      timeout: "10s",
      tags: { perf_auth: "session" },
      responseCallback: http.expectedStatuses(200, 401, 403),
    });

    if (sessionResponse.status === 200) {
      return {
        headers: { Cookie: cookieHeader },
        login: candidate.login,
      };
    }
  }

  fail(
    "Perf auth failed: could not authenticate with provided credentials. " +
      "Set valid PERF_LOGIN/PERF_PASSWORD for backend seed data."
  );
}
