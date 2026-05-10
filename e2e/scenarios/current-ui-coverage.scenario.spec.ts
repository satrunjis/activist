import { randomUUID } from "node:crypto";
import type { Page } from "@playwright/test";

import { expect, test } from "../fixtures/base.fixture";
import { buildRegisterUserInput } from "../support/test-data";
import { trackCreated } from "../support/teardown";
import { ensureLoggedIn, openPrimaryRoute } from "../support/ui-auth";

const ADMIN = {
  login: "admin",
  password: "admin"
} as const;

function uniqueSuffix(): string {
  return randomUUID().replace(/-/g, "").slice(0, 10);
}

async function clickVisibleTestId(page: Page, testId: string): Promise<void> {
  const target = page.locator(`[data-testid="${testId}"]:visible`).first();
  await expect(target).toBeVisible();
  await target.evaluate((node) => (node as HTMLElement).click());
}

test("CURRENT-DIV-01: division tree selection shows info tab and positions tab", async ({ page }) => {
  await ensureLoggedIn(page, ADMIN);

  await expect(page.getByTestId("division-tree-viewport")).toBeVisible();
  await clickVisibleTestId(page, "division-node-seed-ops");
  await expect(page.getByTestId("division-right-panel")).toContainText("Полное название");
  await expect(page.getByTestId("panel-tab-info")).toBeVisible();

  await page.getByTestId("panel-tab-positions").click();
  await expect(page.getByTestId("position-list-panel")).toBeVisible();
  await expect(page.getByTestId("division-right-panel")).toContainText("Должности");
});

test("CURRENT-DIV-02: division create and edit use the right panel", async ({ page, entities }) => {
  await ensureLoggedIn(page, ADMIN);

  const suffix = uniqueSuffix();
  const shortName = `ui-${suffix}`;
  const editedShortName = `ui-edit-${suffix.slice(0, 6)}`;

  await page.getByTestId("create-division").click();
  await expect(page.getByTestId("create-division-form")).toBeVisible();
  await page.getByTestId("create-division-parent").selectOption("seed-ops");
  await page.getByTestId("create-division-full-name").fill(`UI Division ${suffix}`);
  await page.getByTestId("create-division-short-name").fill(shortName);
  await page.getByTestId("create-division-description").fill("Created from current UI coverage");

  const createResponsePromise = page.waitForResponse((response) =>
    response.request().method() === "POST" &&
    response.url().endsWith("/api/v1/divisions") &&
    response.status() === 201
  );
  await page.getByTestId("create-division-submit").click();
  const created = (await createResponsePromise).json() as Promise<{ id: string }>;
  const createdBody = await created;
  trackCreated(entities, "divisions", createdBody.id);

  await expect(page.getByTestId(`division-node-${createdBody.id}`)).toBeVisible();
  await page.getByTestId("edit-division").click();
  await expect(page.getByTestId("edit-division-form")).toBeVisible();
  await page.getByTestId("edit-division-short-name").fill(editedShortName);

  const editResponsePromise = page.waitForResponse((response) =>
    response.request().method() === "PATCH" &&
    response.url().includes(`/api/v1/divisions/${encodeURIComponent(createdBody.id)}`) &&
    response.status() === 200
  );
  await page.getByTestId("edit-division-submit").click();
  await editResponsePromise;
  await expect(page.getByTestId(`division-node-${createdBody.id}`)).toContainText(editedShortName);
});

test("CURRENT-ROLE-01: role create and edit use the right panel active view", async ({ page, entities }) => {
  await openPrimaryRoute(page, ADMIN, "nav-roles", "role-management-page");

  const suffix = uniqueSuffix();
  const roleName = `AAA UI Role ${suffix}`;
  const editedRoleName = `AAA UI Role Edited ${suffix}`;

  await page.getByTestId("create-role-btn").click();
  await expect(page.getByTestId("create-role-form")).toBeVisible();
  await page.getByTestId("role-name").fill(roleName);

  const createResponsePromise = page.waitForResponse((response) =>
    response.request().method() === "POST" &&
    response.url().endsWith("/api/v1/roles") &&
    response.status() === 201
  );
  await page.getByTestId("role-submit").click();
  const createdRole = (await createResponsePromise).json() as Promise<{ id: string }>;
  const createdRoleBody = await createdRole;
  trackCreated(entities, "roles", createdRoleBody.id);

  await expect(page.getByTestId("role-list")).toContainText(roleName);
  await page.getByTestId(`role-item-${createdRoleBody.id}`).click();
  await expect(page.getByTestId("division-right-panel")).toHaveCount(0);
  await page.getByRole("button", { name: "Редактировать" }).click();
  await expect(page.getByTestId("edit-role-form")).toBeVisible();
  await page.getByTestId("role-name").fill(editedRoleName);
  await page.getByTestId("role-submit").click();

  await expect(page.getByTestId("role-list")).toContainText(editedRoleName);
});

test("CURRENT-SEARCH-01: search query returns visible results", async ({ page }) => {
  await openPrimaryRoute(page, ADMIN, "nav-search", "search-panel");

  await page.getByTestId("search-login").fill("admin");
  const searchResponse = page.waitForResponse(
    (response) => response.request().method() === "GET" && response.url().includes("/api/v1/search/users")
  );
  await page.getByTestId("search-submit").click();
  await searchResponse;

  await expect(page.getByTestId("search-results")).toBeVisible();
  await expect(page.getByTestId("search-results")).toContainText("admin");
});

test("CURRENT-PROFILE-01: memberships can be removed through ConfirmDialog", async ({ page, seed }) => {
  await seed.loginUser(ADMIN);
  const role = await seed.createRole({
    name: `remove-member-${uniqueSuffix()}`,
    permissions: ["can_remove_member"]
  });
  const division = await seed.createDivision({
    shortName: `rem-${uniqueSuffix()}`
  });
  const position = await seed.createPosition({
    divisionId: division.id,
    title: `Removable ${uniqueSuffix()}`,
    roleId: role.id
  });
  const userInput = buildRegisterUserInput();
  const user = await seed.registerUser(userInput);
  await seed.loginUser(ADMIN);
  await seed.assignMembership({
    userId: user.body.user.id,
    positionId: position.id
  });

  await openPrimaryRoute(page, { login: userInput.login, password: userInput.password }, "nav-profile");
  await expect(page.getByTestId("memberships-card")).toContainText(position.title);
  await page.getByTestId(`remove-membership-${position.id}-${user.body.user.id}`).click();

  const dialog = page.getByRole("dialog", { name: "Исключить из членства?" });
  await expect(dialog).toBeVisible();
  await dialog.getByRole("button", { name: "Исключить" }).click();
  await expect(page.getByTestId(`remove-membership-${position.id}-${user.body.user.id}`)).toHaveCount(0);
});
