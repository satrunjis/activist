import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { RolesPage } from "./RolesPage";
import type { RoleListResponse } from "../../shared/api/types";
import { ToastProvider } from "../../shared/ui/feedback";

const mockRequest = vi.hoisted(() => vi.fn());

vi.mock("../../shared/api/client", () => ({
  request: mockRequest
}));

const rolesResponse: RoleListResponse = {
  items: [
    {
      id: "role-chair",
      name: "Председатель объединения",
      permissions: [
        { code: "can_add_member", scope: "current_division" },
        { code: "can_create_subdivision", scope: "current_and_descendants" }
      ]
    }
  ],
  limit: 100,
  offset: 0,
  total: 1
};

function renderRolesPage(permissions: readonly string[] = []) {
  mockRequest.mockResolvedValue(rolesResponse);

  return render(
    <ToastProvider>
      <RolesPage permissions={permissions} />
    </ToastProvider>
  );
}

describe("RolesPage behavior", () => {
  beforeEach(() => {
    mockRequest.mockReset();
  });

  it("lets a non-admin inspect role details without edit actions", async () => {
    const user = userEvent.setup();
    renderRolesPage();

    await user.click(await screen.findByTestId("role-item-role-chair"));

    expect(screen.getByText("Добавление участников")).toBeInTheDocument();
    expect(screen.getByText("Создание подподразделений")).toBeInTheDocument();
    expect(screen.getByText("Текущее")).toBeInTheDocument();
    expect(screen.getByText("С потомками")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /создать роль/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /редактировать/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /удалить/i })).not.toBeInTheDocument();
  });

  it("keeps role edit actions available for system admins", async () => {
    const user = userEvent.setup();
    renderRolesPage(["system_admin"]);

    await user.click(await screen.findByTestId("role-item-role-chair"));

    expect(screen.getByRole("button", { name: /редактировать/i })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /удалить/i })).not.toHaveLength(0);
  });
});
