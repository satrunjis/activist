import { describe, it, expect } from "vitest";
import { normalizeApiError, shouldRetryRequest, toFieldErrors } from "../errorPolicy";
import type { ApiRequestError } from "../errorPolicy";

describe("normalizeApiError", () => {
  it("maps backend JSON error with message to structured error preserving human-readable message", () => {
    const raw = JSON.stringify({ error: { code: "auth.invalid", message: "Неверный логин или пароль" } });
    const result = normalizeApiError(raw, 401, "GET");
    expect(result.message).toBe("Неверный логин или пароль");
    expect(result.code).toBe("auth.invalid");
    expect(result.status).toBe(401);
    expect(result.method).toBe("GET");
  });

  it("maps empty or non-JSON error to fallback message", () => {
    const result1 = normalizeApiError("", 500, "GET");
    expect(result1.message).toBe("Что-то пошло не так");

    const result2 = normalizeApiError("not json at all", 503, "POST");
    expect(result2.message).toBe("Что-то пошло не так");

    const result3 = normalizeApiError("   ", 500, "GET");
    expect(result3.message).toBe("Что-то пошло не так");
  });

  it("maps 400/422 payload and exposes field-level data for inline rendering", () => {
    const raw = JSON.stringify({
      error: {
        code: "validation.field",
        message: "Validation failed",
        field: "login"
      }
    });
    const result = normalizeApiError(raw, 422, "POST");
    expect(result.message).toBe("Validation failed");
    expect(result.code).toBe("validation.field");
    expect(result.fieldErrors).toBeDefined();
    expect(result.fieldErrors?.["login"]).toBe("Validation failed");
  });
});

describe("shouldRetryRequest", () => {
  it("returns true for GET + 5xx failures", () => {
    const error: ApiRequestError = {
      message: "Что-то пошло не так",
      code: "internal.error",
      status: 500,
      method: "GET",
      retryable: true
    };
    expect(shouldRetryRequest(error, "GET")).toBe(true);
  });

  it("returns true for GET + network-like conditions (no status / timeout)", () => {
    const error: ApiRequestError = {
      message: "Что-то пошло не так",
      code: "network.error",
      status: 0,
      method: "GET",
      retryable: true
    };
    expect(shouldRetryRequest(error, "GET")).toBe(true);
  });

  it("returns false for POST even on 5xx", () => {
    const error: ApiRequestError = {
      message: "Что-то пошло не так",
      code: "internal.error",
      status: 500,
      method: "POST",
      retryable: false
    };
    expect(shouldRetryRequest(error, "POST")).toBe(false);
  });

  it("returns false for GET with 4xx client errors", () => {
    const error: ApiRequestError = {
      message: "Not found",
      code: "not.found",
      status: 404,
      method: "GET",
      retryable: false
    };
    expect(shouldRetryRequest(error, "GET")).toBe(false);
  });

  it("returns false for DELETE + network failure", () => {
    const error: ApiRequestError = {
      message: "Что-то пошло не так",
      code: "network.error",
      status: 0,
      method: "DELETE",
      retryable: false
    };
    expect(shouldRetryRequest(error, "DELETE")).toBe(false);
  });
});

describe("toFieldErrors", () => {
  it("returns a field errors map from a field-level validation error", () => {
    const error: ApiRequestError = {
      message: "Login is required",
      code: "validation.field",
      status: 422,
      method: "POST",
      retryable: false,
      fieldErrors: { login: "Login is required" }
    };
    const result = toFieldErrors(error);
    expect(result["login"]).toBe("Login is required");
  });

  it("returns empty map when there are no field errors", () => {
    const error: ApiRequestError = {
      message: "Что-то пошло не так",
      code: "internal.error",
      status: 500,
      method: "GET",
      retryable: false
    };
    const result = toFieldErrors(error);
    expect(Object.keys(result)).toHaveLength(0);
  });
});
