/**
 * Centralized API error parsing, retry policy, and field error utilities.
 *
 * Design decisions (09-CONTEXT.md):
 *   D-08: Use API `error.message` when readable; fallback to generic Russian message.
 *   D-09: Retry is allowed only for GET + 5xx or network-like (no status / timeout) failures.
 *   D-10: Validation errors (400/422) must be representable inline at field level.
 */

/** Structured error thrown by the shared API client after normalization. */
export type ApiRequestError = {
  /** Human-readable message: API message if readable, else fallback. */
  message: string;
  /** Machine-readable error code from the API envelope, if present. */
  code?: string;
  /** HTTP status code (0 for network/timeout). */
  status: number;
  /** HTTP method that produced the error. */
  method: string;
  /** Whether this error qualifies for a retry action (GET + 5xx/network only). */
  retryable: boolean;
  /** Field-level errors for inline form rendering (populated for 400/422 with field). */
  fieldErrors?: Record<string, string>;
};

const FALLBACK_MESSAGE = "Что-то пошло не так";

type ApiErrorEnvelopeShape = {
  error?: {
    code?: string;
    message?: string;
    field?: string;
  };
};

function parseEnvelope(raw: string): ApiErrorEnvelopeShape | null {
  if (!raw.trim()) return null;
  try {
    return JSON.parse(raw) as ApiErrorEnvelopeShape;
  } catch {
    return null;
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 0 || status >= 500;
}

/**
 * Normalize a raw API error response body into a structured `ApiRequestError`.
 *
 * @param raw     - Raw response text from the API.
 * @param status  - HTTP status code (use 0 for network / timeout failures).
 * @param method  - HTTP method of the failed request.
 */
export function normalizeApiError(raw: string, status: number, method: string): ApiRequestError {
  const parsed = parseEnvelope(raw);
  const errorBody = parsed?.error;

  const message = errorBody?.message?.trim() || FALLBACK_MESSAGE;
  const code = errorBody?.code?.trim();
  const field = errorBody?.field?.trim();
  const retryable = method === "GET" && isRetryableStatus(status);

  let fieldErrors: Record<string, string> | undefined;
  if (field && (status === 400 || status === 422)) {
    fieldErrors = { [field]: message };
  }

  return {
    message,
    code,
    status,
    method,
    retryable,
    fieldErrors
  };
}

/**
 * Determine whether a retry UI affordance should be shown for the given error.
 *
 * Hard gate (D-09): only GET requests with 5xx or network-like conditions qualify.
 *
 * @param error  - The normalized ApiRequestError from `normalizeApiError`.
 * @param method - HTTP method of the failed request (double-checked against error.method).
 */
export function shouldRetryRequest(error: ApiRequestError, method: string): boolean {
  if (method !== "GET") return false;
  return isRetryableStatus(error.status);
}

/**
 * Extract a flat field-error map from an ApiRequestError for inline form rendering.
 * Returns an empty object when no field errors are present.
 *
 * @param error - The normalized ApiRequestError.
 */
export function toFieldErrors(error: ApiRequestError): Record<string, string> {
  return error.fieldErrors ?? {};
}
