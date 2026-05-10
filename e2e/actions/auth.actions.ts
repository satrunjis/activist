import { expect, type Page } from "@playwright/test";
import type { RegisterUserInput } from "../support/test-data";
import { trackCreated, type CreatedEntities } from "../support/teardown";
import { ensureLoggedIn, readBrowserSession } from "../support/ui-auth";

export type LoginUserInput = {
  login: string;
  password: string;
};

export type AuthUserResult = {
  user: {
    id: string;
    login: string;
    first_name?: string;
    gradebook_number?: string;
  };
};

export async function openAnonymousAuthScreen(
  page: Page,
  _entities: CreatedEntities
): Promise<void> {
  await page.goto("/");
  if ((await page.getByRole("button", { name: "Выйти" }).count()) > 0) {
    await page.getByRole("button", { name: "Выйти" }).click();
  }
  await expect(page.getByRole("heading", { name: "Добро пожаловать" })).toBeVisible();
  await expect(page.getByRole("tab", { name: "Вход" })).toBeVisible();
  await expect(page.getByRole("tab", { name: "Регистрация" })).toBeVisible();
}

export async function registerUser(
  page: Page,
  entities: CreatedEntities,
  input: RegisterUserInput
): Promise<{ body: AuthUserResult }> {
  await openAnonymousAuthScreen(page, entities);

  await page.getByRole("tab", { name: "Регистрация" }).click();
  const registerForm = page.locator('form:has(button:has-text("Зарегистрироваться"))');

  await registerForm.locator('input[name="login"]').fill(input.login);
  await registerForm.locator('input[name="password"]').fill(input.password);
  await registerForm.locator('input[name="first_name"]').fill(input.firstName);
  await registerForm.locator('input[name="gradebook_number"]').fill(input.gradebookNumber);
  await registerForm.locator('input[name="group_number"]').fill(input.groupNumber);
  await registerForm.locator('input[name="institute"]').fill(input.institute);
  await registerForm.locator('input[name="birth_date"]').fill(input.birthDate);
  await registerForm.getByRole("button", { name: "Зарегистрироваться" }).click();

  await expect(page.getByRole("button", { name: "Выйти" })).toBeVisible();

  const session = await readBrowserSession(page);
  const id = session?.user?.id?.trim();
  if (!id) {
    throw new Error("registerUser failed: authenticated user id is not present in session");
  }

  await page.getByTestId("nav-profile").click();
  await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();
  await expect(page.getByText("Логин")).toBeVisible();
  await expect(page.getByTestId("app-shell-content").getByText(input.login)).toBeVisible();

  const body: AuthUserResult = {
    user: {
      id,
      login: input.login,
      first_name: input.firstName,
      gradebook_number: input.gradebookNumber
    }
  };
  trackCreated(entities, "users", body.user.id);
  return { body };
}

export async function loginUser(
  page: Page,
  _entities: CreatedEntities,
  input: LoginUserInput
): Promise<{ body: AuthUserResult }> {
  const session = await ensureLoggedIn(page, input);
  const id = session.user?.id?.trim();
  if (!id) {
    throw new Error("loginUser failed: authenticated user id is not present in session");
  }

  return {
    body: {
      user: {
        id,
        login: input.login
      }
    }
  };
}

export async function logoutUser(page: Page, _entities: CreatedEntities): Promise<void> {
  await page.getByRole("button", { name: "Выйти" }).click();
  await expect(page.getByRole("heading", { name: "Добро пожаловать" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Войти" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Выйти" })).toHaveCount(0);
}

export async function reloadWithActiveSession(
  page: Page,
  _entities: CreatedEntities
): Promise<{ userId: string }> {
  await page.reload();
  await expect(page.getByRole("button", { name: "Выйти" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Войти" })).toHaveCount(0);

  const session = await readBrowserSession(page);
  const userId = session?.user?.id?.trim();
  if (!userId) {
    throw new Error("reloadWithActiveSession failed: authenticated user id is not present in session");
  }

  return { userId };
}
