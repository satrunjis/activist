import type { ApiRequestError } from "./client";

export type UiError = {
  message: string;
  code?: string;
  status: number;
  retryable: boolean;
  fieldErrors: Record<string, string>;
  kind: "validation" | "forbidden" | "conflict" | "network" | "unknown";
};

function isApiRequestError(error: unknown): error is ApiRequestError {
  if (!error || typeof error !== "object") {
    return false;
  }
  const candidate = error as Partial<ApiRequestError>;
  return typeof candidate.message === "string" && typeof candidate.status === "number";
}

function fallbackMessage(fallback: string): string {
  const normalized = fallback.trim();
  return normalized.length > 0 ? normalized : "Что-то пошло не так";
}

export function adaptApiError(error: unknown, fallback: string): UiError {
  const fallbackText = fallbackMessage(fallback);

  if (!isApiRequestError(error)) {
    return {
      message: error instanceof Error && error.message.trim() ? error.message : fallbackText,
      status: -1,
      retryable: true,
      fieldErrors: {},
      kind: "unknown"
    };
  }

  const message = error.message.trim() || fallbackText;
  const code = error.code?.trim() || undefined;
  const status = error.status;
  const fieldErrors = error.fieldErrors ?? {};

  let kind: UiError["kind"] = "unknown";
  if (status === 0) {
    kind = "network";
  } else if (status === 409) {
    kind = "conflict";
  } else if ((status === 400 || status === 422) && Object.keys(fieldErrors).length > 0) {
    kind = "validation";
  } else if (code?.startsWith("access.") || status === 403) {
    kind = "forbidden";
  }

  return {
    message,
    code,
    status,
    retryable: kind === "network" || kind === "unknown",
    fieldErrors,
    kind
  };
}

export function getFieldErrors(error: UiError): Record<string, string> {
  return error.fieldErrors;
}

export function getRootMessage(error: UiError): string {
  return error.message;
}
