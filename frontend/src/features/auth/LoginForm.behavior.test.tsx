import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LoginForm } from "./LoginForm";
import { RegisterForm } from "./RegisterForm";
import { ToastProvider } from "../../shared/ui/feedback";

const mockRequest = vi.fn();

vi.mock("../../shared/api/client", () => ({
  request: (...args: unknown[]) => mockRequest(...args),
  setCsrfToken: vi.fn(),
  clearCsrfToken: vi.fn(),
  getCsrfToken: vi.fn()
}));

describe("LoginForm behavior", () => {
  beforeEach(() => {
    mockRequest.mockReset();
  });

  function renderLoginForm(onLoggedIn = vi.fn(async () => undefined)) {
    render(
      <ToastProvider>
        <LoginForm onLoggedIn={onLoggedIn} />
      </ToastProvider>
    );
    return { onLoggedIn };
  }

  it("blocks login submit and shows validation errors on invalid inputs", async () => {
    const user = userEvent.setup();
    renderLoginForm();

    await user.type(screen.getByLabelText(/логин/i), "ab");
    await user.click(screen.getByRole("button", { name: /^войти$/i }));

    await waitFor(() => {
      expect(screen.getByText(/логин должен соответствовать шаблону/i)).toBeInTheDocument();
      expect(screen.getByText(/пароль обязателен/i)).toBeInTheDocument();
    });
    expect(mockRequest).not.toHaveBeenCalled();
  });

  it("shows toast error and keeps page state on invalid credentials", async () => {
    const user = userEvent.setup();
    const onLoggedIn = vi.fn(async () => undefined);

    mockRequest.mockRejectedValueOnce({
      message: "Неверный логин или пароль",
      code: "auth.invalid",
      status: 401,
      method: "POST",
      retryable: false
    });

    renderLoginForm(onLoggedIn);

    await user.type(screen.getByLabelText(/логин/i), "tester");
    await user.type(screen.getByLabelText(/пароль/i), "wrong-password");
    await user.click(screen.getByRole("button", { name: /^войти$/i }));

    await waitFor(() => {
      expect(screen.getByText("Не удалось выполнить вход")).toBeInTheDocument();
      expect(screen.getByText("Неверный логин или пароль")).toBeInTheDocument();
    });

    expect(mockRequest).toHaveBeenCalledWith("/api/v1/auth/login", expect.objectContaining({
      skipUnauthorizedRedirect: true
    }));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(onLoggedIn).not.toHaveBeenCalled();
  });

  it("blocks register submit and shows validation errors for missing required fields", async () => {
    const user = userEvent.setup();
    const onRegistered = vi.fn(async () => undefined);
    render(<RegisterForm onRegistered={onRegistered} />);

    await user.click(screen.getByRole("button", { name: /^зарегистрироваться$/i }));

    await waitFor(() => {
      expect(screen.getByText(/пароль обязателен/i)).toBeInTheDocument();
      expect(screen.getByText(/имя обязательно/i)).toBeInTheDocument();
    });
    expect(mockRequest).not.toHaveBeenCalled();
    expect(onRegistered).not.toHaveBeenCalled();
  });
});
