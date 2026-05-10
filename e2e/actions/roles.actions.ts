import { expect, type Page } from "@playwright/test";
import { trackCreated, type CreatedEntities } from "../support/teardown";
import { openPrimaryRoute } from "../support/ui-auth";

type RoleActorInput = {
  login: string;
  password: string;
};

type PermissionScope = "current_division" | "current_and_descendants";

type RolePermissionInput = {
  code: string;
  enabled?: boolean;
  scope?: PermissionScope;
};

type RoleIdentity = {
  id: string;
  name: string;
};

type ErrorExpectation = {
  expectedError: string | RegExp;
};

export type ListRolesInput = {
  actor: RoleActorInput;
};

export type CreateRoleInput = {
  actor: RoleActorInput;
  name: string;
  permissions?: RolePermissionInput[];
} & Partial<ErrorExpectation>;

export type EditRoleInput = {
  actor: RoleActorInput;
  roleId?: string;
  roleName?: string;
  nextName?: string;
  permissions?: RolePermissionInput[];
  verifyScopes?: Record<string, PermissionScope>;
} & Partial<ErrorExpectation>;

export type DeleteRoleInput = {
  actor: RoleActorInput;
  roleId?: string;
  roleName?: string;
};

export type DeleteRoleInUseShowsConflictInput = {
  actor: RoleActorInput;
  roleId: string;
  expectedMessage: string | RegExp;
};

export type ViewRoleManagementReadonlyInput = {
  actor: RoleActorInput;
  expectedMessage?: string | RegExp;
};

type FindRoleInput = {
  roleId?: string;
  roleName?: string;
};

type RoleApiResponse = {
  id?: string;
  name?: string;
};

async function loginAndOpenRoleManagement(page: Page, actor: RoleActorInput): Promise<void> {
  await openPrimaryRoute(page, actor, "nav-roles", "role-management-page");
}

function extractRoleId(testId: string | null | undefined): string {
  if (!testId || !testId.startsWith("role-item-")) {
    throw new Error("Role list item test id is missing or malformed");
  }
  return testId.replace("role-item-", "");
}

async function goToFirstRolesPage(page: Page): Promise<void> {
  const previous = page.getByTestId("roles-pagination-prev");
  while (!(await previous.isDisabled())) {
    const state = page.getByTestId("roles-pagination-state");
    const previousState = (await state.textContent()) ?? "";
    await previous.click();
    await expect(state).not.toHaveText(previousState);
  }
}

async function goToNextRolesPage(page: Page): Promise<boolean> {
  const next = page.getByTestId("roles-pagination-next");
  if (await next.isDisabled()) {
    return false;
  }
  const state = page.getByTestId("roles-pagination-state");
  const previousState = (await state.textContent()) ?? "";
  await next.click();
  await expect(state).not.toHaveText(previousState);
  return true;
}

async function findRoleIdentityOnVisiblePages(page: Page, input: FindRoleInput): Promise<RoleIdentity | null> {
  await goToFirstRolesPage(page);

  for (let pageIndex = 0; pageIndex < 25; pageIndex += 1) {
    if (input.roleId) {
      const roleNameLocator = page.getByTestId(`role-name-${input.roleId}`);
      if ((await roleNameLocator.count()) > 0 && await roleNameLocator.first().isVisible()) {
        const roleName = (await roleNameLocator.first().textContent())?.trim();
        if (!roleName) {
          throw new Error(`Role name is not visible for role id ${input.roleId}`);
        }
        return { id: input.roleId, name: roleName };
      }
    }

    if (input.roleName) {
      const roleNameLocator = page.getByText(input.roleName, { exact: true }).first();
      if ((await roleNameLocator.count()) > 0 && await roleNameLocator.isVisible()) {
        const roleItem = roleNameLocator.locator('xpath=ancestor::*[starts-with(@data-testid, "role-item-")][1]');
        const roleId = extractRoleId(await roleItem.getAttribute("data-testid"));
        return { id: roleId, name: input.roleName };
      }
    }

    if (!(await goToNextRolesPage(page))) {
      break;
    }
  }

  return null;
}

async function findRoleIdentity(page: Page, input: FindRoleInput): Promise<RoleIdentity> {
  if (!input.roleId && !input.roleName) {
    throw new Error("Either roleId or roleName must be provided");
  }

  const found = await findRoleIdentityOnVisiblePages(page, input);
  if (found) {
    return found;
  }

  const refresh = page.getByTestId("refresh-roles");
  if ((await refresh.count()) > 0 && await refresh.isVisible()) {
    await refresh.click();
    await expect(page.getByTestId("role-list")).toBeVisible();
    const refreshed = await findRoleIdentityOnVisiblePages(page, input);
    if (refreshed) {
      return refreshed;
    }
  }

  throw new Error(`Role was not found: ${input.roleId ?? input.roleName}`);
}

async function applyPermissions(page: Page, permissions: RolePermissionInput[] | undefined): Promise<void> {
  for (const permission of permissions ?? []) {
    const checkbox = page.getByTestId(`permission-${permission.code}`);
    if (permission.enabled === false) {
      await checkbox.uncheck();
      continue;
    }
    await checkbox.check();
    if (permission.scope) {
      await page.getByTestId(`permission-scope-${permission.code}`).selectOption(permission.scope);
    }
  }
}

export async function listRoles(
  page: Page,
  _entities: CreatedEntities,
  input: ListRolesInput
): Promise<{ items: RoleIdentity[] }> {
  await loginAndOpenRoleManagement(page, input.actor);
  const list = page.getByTestId("role-list");
  await expect(list).toBeVisible();

  const items: RoleIdentity[] = [];
  const listItems = page.locator('[data-testid^="role-item-"]');
  const count = await listItems.count();
  for (let idx = 0; idx < count; idx += 1) {
    const item = listItems.nth(idx);
    const roleId = extractRoleId(await item.getAttribute("data-testid"));
    const roleName = (await page.getByTestId(`role-name-${roleId}`).textContent())?.trim() ?? "";
    if (!roleName) {
      continue;
    }
    items.push({ id: roleId, name: roleName });
  }

  return { items };
}

export async function createRole(
  page: Page,
  entities: CreatedEntities,
  input: CreateRoleInput
): Promise<{ role?: RoleIdentity; error?: string }> {
  await loginAndOpenRoleManagement(page, input.actor);
  await page.click('[data-testid="create-role-btn"]');
  await expect(page.getByTestId("create-role-form")).toBeVisible();

  await page.fill('[data-testid="role-name"]', input.name);
  await applyPermissions(page, input.permissions);
  if (input.expectedError && input.name.trim().length === 0) {
    await expect(page.getByTestId("role-submit")).toBeDisabled();
    return { error: "Название обязательно." };
  }
  const createResponsePromise = input.expectedError
    ? null
    : page.waitForResponse((response) =>
        response.request().method() === "POST"
        && response.url().includes("/api/v1/roles")
        && response.status() >= 200
        && response.status() < 300
      );
  await page.click('[data-testid="role-submit"]');

  if (input.expectedError) {
    const error = page.getByRole("alert");
    await expect(error).toContainText(input.expectedError);
    return { error: (await error.textContent())?.trim() ?? "" };
  }

  await expect(page.getByTestId("create-role-form")).toHaveCount(0);
  const createResponse = await createResponsePromise;
  const createdFromApi = (await createResponse?.json().catch(() => undefined)) as RoleApiResponse | undefined;
  const role = createdFromApi?.id
    ? { id: createdFromApi.id, name: createdFromApi.name ?? input.name }
    : await findRoleIdentity(page, { roleName: input.name });
  await findRoleIdentity(page, { roleId: role.id, roleName: role.name });
  trackCreated(entities, "roles", role.id);
  return { role };
}

export async function editRole(
  page: Page,
  _entities: CreatedEntities,
  input: EditRoleInput
): Promise<{ role: RoleIdentity; error?: string }> {
  await loginAndOpenRoleManagement(page, input.actor);
  const original = await findRoleIdentity(page, { roleId: input.roleId, roleName: input.roleName });

  await page.getByTestId(`role-item-${original.id}`).click();
  await page.getByRole("button", { name: "Редактировать" }).click();
  await expect(page.getByTestId("edit-role-form")).toBeVisible();

  if (input.nextName !== undefined) {
    await page.fill('[data-testid="role-name"]', input.nextName);
  }
  await applyPermissions(page, input.permissions);
  if (input.expectedError && input.nextName !== undefined && input.nextName.trim().length === 0) {
    await expect(page.getByTestId("role-submit")).toBeDisabled();
    return { role: original, error: "Название обязательно." };
  }
  await page.click('[data-testid="role-submit"]');

  if (input.expectedError) {
    const error = page.getByRole("alert");
    await expect(error).toContainText(input.expectedError);
    return { role: original, error: (await error.textContent())?.trim() ?? "" };
  }

  const updatedName = input.nextName !== undefined ? input.nextName : original.name;
  const updated = await findRoleIdentity(page, { roleName: updatedName });

  if (input.verifyScopes) {
    await page.getByTestId(`role-item-${updated.id}`).click();
    await page.getByRole("button", { name: "Редактировать" }).click();
    await expect(page.getByTestId("edit-role-form")).toBeVisible();
    for (const [code, expectedScope] of Object.entries(input.verifyScopes)) {
      await expect(page.getByTestId(`permission-scope-${code}`)).toHaveValue(expectedScope);
    }
  }

  return { role: updated };
}

export async function deleteRole(
  page: Page,
  _entities: CreatedEntities,
  input: DeleteRoleInput
): Promise<{ deletedRoleId: string; deletedRoleName: string }> {
  await loginAndOpenRoleManagement(page, input.actor);
  const role = await findRoleIdentity(page, { roleId: input.roleId, roleName: input.roleName });

  await page.click(`[data-testid="delete-role-${role.id}"]`);
  await page.getByRole("dialog", { name: "Удалить роль?" }).getByRole("button", { name: "Удалить" }).click();
  await expect(page.getByTestId("role-list")).not.toContainText(role.name);
  return {
    deletedRoleId: role.id,
    deletedRoleName: role.name
  };
}

export async function deleteRoleInUseShowsConflict(
  page: Page,
  _entities: CreatedEntities,
  input: DeleteRoleInUseShowsConflictInput
): Promise<{ error: string }> {
  await loginAndOpenRoleManagement(page, input.actor);
  await findRoleIdentity(page, { roleId: input.roleId });
  await page.click(`[data-testid="delete-role-${input.roleId}"]`);
  await page.getByRole("dialog", { name: "Удалить роль?" }).getByRole("button", { name: "Удалить" }).click();
  const error = page.getByText(input.expectedMessage).first();
  await expect(error).toContainText(input.expectedMessage);
  return { error: (await error.textContent())?.trim() ?? "" };
}

export async function viewRoleManagementReadonly(
  page: Page,
  _entities: CreatedEntities,
  input: ViewRoleManagementReadonlyInput
): Promise<{ message: string }> {
  await loginAndOpenRoleManagement(page, input.actor);
  const readonlyMessage = page.getByText("Управление ролями доступно только администраторам.");
  await expect(readonlyMessage).toBeVisible();
  await expect(page.getByTestId("create-role-btn")).toHaveCount(0);
  await expect(page.locator('[data-testid^="delete-role-"]')).toHaveCount(0);
  if (input.expectedMessage) {
    await expect(readonlyMessage).toContainText(input.expectedMessage);
  }
  return {
    message: (await readonlyMessage.textContent())?.trim() ?? ""
  };
}
