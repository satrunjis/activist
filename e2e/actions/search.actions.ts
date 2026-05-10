import { expect, type Page } from "@playwright/test";
import { openPrimaryRoute } from "../support/ui-auth";

type LoginInput = {
  login: string;
  password: string;
};

export type UserSearchInput = {
  actor?: LoginInput;
  first_name?: string;
  last_name?: string;
  middle_name?: string;
  login?: string;
  group_number?: string;
  institute?: string;
  about?: string;
  position_title?: string;
  role_name?: string;
  include_archived?: boolean;
  submit?: boolean;
};

const FIELD_TO_TEST_ID: Record<string, string> = {
  first_name: "search-first_name",
  last_name: "search-last_name",
  middle_name: "search-middle_name",
  login: "search-login",
  group_number: "search-group_number",
  institute: "search-institute",
  about: "search-about",
  position_title: "search-position_title",
  role_name: "search-role_name"
};

async function ensureAuthenticated(page: Page, actor: LoginInput): Promise<void> {
  await openPrimaryRoute(page, actor, "nav-search", "search-panel");
}

export async function runUserSearch(page: Page, input: UserSearchInput): Promise<void> {
  if (input.actor) {
    await ensureAuthenticated(page, input.actor);
  }

  for (const [field, testId] of Object.entries(FIELD_TO_TEST_ID)) {
    const value = input[field as keyof UserSearchInput];
    await page.getByTestId(testId).fill(typeof value === "string" ? value : "");
  }

  {
    const checkbox = page.getByTestId("search-include_archived");
    const includeArchived = input.include_archived ?? false;
    if (includeArchived) {
      await checkbox.check();
    } else {
      await checkbox.uncheck();
    }
  }

  if (input.submit !== false) {
    const responsePromise = page.waitForResponse(
      (response) => response.request().method() === "GET" && response.url().includes("/api/v1/search/users")
    );
    await page.getByTestId("search-submit").click();
    await responsePromise;
  }
}

export async function assertSearchRows(page: Page, expectedLogins: string[]): Promise<void> {
  const list = page.getByTestId("search-results");
  await expect(list).toBeVisible();

  for (const login of expectedLogins) {
    await expect(list.getByText(login)).toBeVisible();
  }

  const rows = list.locator('li[data-testid^="search-result-"]');
  const count = await rows.count();
  const discoveredLogins: string[] = [];
  for (let index = 0; index < count; index += 1) {
    const text = (await rows.nth(index).textContent()) ?? "";
    const match = expectedLogins.find((login) => text.includes(login));
    if (match) discoveredLogins.push(match);
  }

  for (const login of discoveredLogins) {
    expect(expectedLogins).toContain(login);
  }
}

export async function assertContactVisibility(page: Page, login: string, visible: boolean): Promise<void> {
  const row = page.locator('li[data-testid^="search-result-"]').filter({ hasText: login }).first();
  await expect(row).toBeVisible();
  await row.click();

  const details = page.getByTestId("search-user-details-panel");
  await expect(details).toBeVisible();
  const phoneField = details.locator(".meta-block__field").filter({ hasText: "Телефон" }).first();
  const phoneText = (await phoneField.textContent()) ?? "";
  const hasPhone = !phoneText.includes("—");
  if (visible) {
    expect(hasPhone).toBeTruthy();
    return;
  }

  expect(hasPhone).toBeFalsy();
}
