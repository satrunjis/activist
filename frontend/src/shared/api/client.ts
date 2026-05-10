import { normalizeApiError } from "./errorPolicy";
import type { ApiRequestError } from "./errorPolicy";

export type { ApiRequestError };

type RequestMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

type RequestOptions = Omit<RequestInit, "body" | "method"> & {
  method?: RequestMethod;
  body?: unknown;
  skipUnauthorizedRedirect?: boolean;
};

const mutatingMethods = new Set<RequestMethod>(["POST", "PUT", "PATCH", "DELETE"]);
const apiBaseUrl = "";
const AUTH_ROUTE_PATH = "/";

let isUnauthorizedRedirectInProgress = false;

let csrfToken: string | null = null;

export function setCsrfToken(token: string): void {
  csrfToken = token;
}

export function getCsrfToken(): string | null {
  return csrfToken;
}

export function clearCsrfToken(): void {
  csrfToken = null;
}

function toUrl(path: string): string {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  if (!apiBaseUrl) {
    return path;
  }
  const base = apiBaseUrl.endsWith("/") ? apiBaseUrl.slice(0, -1) : apiBaseUrl;
  const suffix = path.startsWith("/") ? path : `/${path}`;
  return `${base}${suffix}`;
}

function resolvePathname(value: string): string {
  try {
    return new URL(value, window.location.origin).pathname;
  } catch {
    return value;
  }
}

function handleUnauthorizedResponse(requestPath: string): void {
  clearCsrfToken();

  const pathname = resolvePathname(toUrl(requestPath));
  const isSessionProbeOnAuthRoute = pathname.endsWith("/api/v1/auth/session")
    && window.location.pathname === AUTH_ROUTE_PATH;

  if (isSessionProbeOnAuthRoute || isUnauthorizedRedirectInProgress) {
    return;
  }

  isUnauthorizedRedirectInProgress = true;
  if (window.location.pathname === AUTH_ROUTE_PATH) {
    window.location.reload();
    return;
  }
  window.location.href = AUTH_ROUTE_PATH;
}

function buildHeaders(method: RequestMethod, body: unknown, headers?: HeadersInit): Headers {
  const next = new Headers(headers);
  if (body !== undefined && !next.has("Content-Type")) {
    next.set("Content-Type", "application/json");
  }
  if (mutatingMethods.has(method) && csrfToken && !next.has("X-CSRF-Token")) {
    next.set("X-CSRF-Token", csrfToken);
  }
  return next;
}

export async function request<TResponse>(path: string, options: RequestOptions = {}): Promise<TResponse> {
  const method = options.method ?? "GET";
  const headers = buildHeaders(method, options.body, options.headers);
  const { skipUnauthorizedRedirect, ...fetchOptions } = options;

  let response: Response;
  try {
    response = await fetch(toUrl(path), {
      ...fetchOptions,
      method,
      credentials: "include",
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body)
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") throw error;
    // Network error / timeout: status 0
    const networkError = normalizeApiError("", 0, method);
    throw networkError;
  }

  if (!response.ok) {
    const text = await response.text();
    if (response.status === 401 && !skipUnauthorizedRedirect) {
      handleUnauthorizedResponse(path);
    }
    const apiError: ApiRequestError = normalizeApiError(text, response.status, method);
    throw apiError;
  }

  if (response.status === 204) {
    return undefined as TResponse;
  }

  return (await response.json()) as TResponse;
}
