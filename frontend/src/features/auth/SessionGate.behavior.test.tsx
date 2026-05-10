import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SessionGate } from "./SessionGate";
import type { AuthSessionResponse } from "../../shared/api/types";

vi.mock("../../shared/api/client", () => ({
  request: vi.fn(),
  setCsrfToken: vi.fn(),
  clearCsrfToken: vi.fn()
}));

import { request } from "../../shared/api/client";
const mockRequest = vi.mocked(request);

const mockSession: AuthSessionResponse = {
  csrf_token: "test-csrf",
  permissions: ["can_manage_positions"],
  session: { expires_at: "2099-01-01T00:00:00Z", idle_expires_at: "2099-01-01T00:00:00Z" },
  user: {
    id: "user-1",
    login: "test_user",
    first_name: "Test",
    last_name: undefined,
    middle_name: undefined,
    gradebook_number: undefined,
    group_number: undefined,
    institute: undefined,
    birth_date: undefined,
    phone: undefined,
    social_links: [],
    about: undefined,
    memberships: []
  }
};

describe("SessionGate behavior", () => {
  it("shows loading state while bootstrapping session", () => {
    mockRequest.mockReturnValueOnce(new Promise(() => undefined));

    render(<SessionGate>{() => <div>authenticated content</div>}</SessionGate>);

    expect(screen.getByText(/проверка сессии/i)).toBeInTheDocument();
    expect(screen.queryByText(/authenticated content/)).not.toBeInTheDocument();
  });

  it("transitions to anonymous when session request returns 401", async () => {
    mockRequest.mockRejectedValueOnce(new Error("401 Unauthorized"));

    render(<SessionGate>{() => <div>authenticated content</div>}</SessionGate>);

    expect(await screen.findByRole("button", { name: /войти/i })).toBeInTheDocument();
    expect(screen.queryByText(/authenticated content/)).not.toBeInTheDocument();
  });

  it("transitions to authenticated when session request succeeds", async () => {
    mockRequest.mockResolvedValueOnce(mockSession);

    render(
      <SessionGate>
        {({ user }) => <div>Welcome {user.login}</div>}
      </SessionGate>
    );

    expect(await screen.findByText(/welcome test_user/i)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /войти/i })).not.toBeInTheDocument();
  });
});
