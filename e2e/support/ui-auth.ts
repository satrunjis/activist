import { expect, type Page } from "@playwright/test";
import { env } from "./env";

export type UiCredentials = {
  login: string;
  password: string;
};

export type BrowserSession = {
  user?: {
    id?: string;
    login?: string;
    first_name?: string;
  };
};

export async function readBrowserSession(page: Page): Promise<BrowserSession | null> {
  const response = await page.context().request.get(`${env.apiBaseUrl}/auth/session`);
  if (response.status() === 401) {
    return null;
  }
  if (!response.ok()) {
    throw new Error(`GET /api/v1/auth/session returned ${response.status()}`);
  }
  return (await response.json()) as BrowserSession;
}

export async function ensureLoggedIn(page: Page, auth: UiCredentials): Promise<BrowserSession> {
  const currentSession = await readBrowserSession(page);
  const currentLogin = currentSession?.user?.login?.trim();
  if (currentLogin === auth.login && !page.url().startsWith("about:")) {
    await expect(page.getByTestId("nav-home")).toBeVisible();
    return currentSession;
  }

  await page.goto("/");

  if (currentLogin && currentLogin !== auth.login) {
    await expect(page.getByRole("button", { name: "Выйти" })).toBeVisible();
    await page.getByRole("button", { name: "Выйти" }).click();
    await expect(page.getByRole("button", { name: "Войти" })).toBeVisible();
  } else if (currentLogin === auth.login) {
    await expect(page.getByTestId("nav-home")).toBeVisible();
    if (!currentSession) {
      throw new Error("ensureLoggedIn failed: current session disappeared");
    }
    return currentSession;
  }

  const loginForm = page.locator('form:has(button:has-text("Войти"))');
  await expect(loginForm).toBeVisible();
  await loginForm.locator('input[name="login"]').fill(auth.login);
  await loginForm.locator('input[name="password"]').fill(auth.password);
  await loginForm.getByRole("button", { name: "Войти" }).click();
  await expect(page.getByTestId("nav-home")).toBeVisible();

  const nextSession = await readBrowserSession(page);
  if (!nextSession?.user?.id) {
    throw new Error("ensureLoggedIn failed: authenticated session has no user id");
  }
  return nextSession;
}

export async function openPrimaryRoute(
  page: Page,
  auth: UiCredentials,
  navTestId: "nav-home" | "nav-roles" | "nav-profile" | "nav-search" | "nav-audit-log",
  pageTestId?: string
): Promise<BrowserSession> {
  const session = await ensureLoggedIn(page, auth);
  const navButton = page.getByTestId(navTestId);
  await expect(navButton).toBeVisible();
  await navButton.click();
  if (pageTestId) {
    await expect(page.getByTestId(pageTestId)).toBeVisible();
  }
  return session;
}
