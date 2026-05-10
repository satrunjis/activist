import {
  loginUser,
  logoutUser,
  openAnonymousAuthScreen,
  registerUser,
  reloadWithActiveSession
} from "../actions/auth.actions";
import { updateOwnProfile } from "../actions/profile.actions";
import { test, expect } from "../fixtures/base.fixture";
import { buildRegisterUserInput } from "../support/test-data";

test("anonymous state shows Register and Login forms", async ({ page, entities }) => {
  const expected = {
    registerHeading: "Регистрация",
    loginHeading: "Вход"
  };

  await openAnonymousAuthScreen(page, entities);

  await expect(page.getByRole("heading", { name: "Добро пожаловать" })).toBeVisible();
  await expect(page.getByRole("tab", { name: expected.registerHeading })).toBeVisible();
  await expect(page.getByRole("tab", { name: expected.loginHeading })).toBeVisible();
  await expect(page.getByRole("button", { name: "Войти" })).toBeVisible();
});

test("register flow reaches authenticated profile screen", async ({ page, entities }) => {
  const input = buildRegisterUserInput();
  const expected = {
    firstName: input.firstName,
    membershipsHeading: "Членства"
  };

  await registerUser(page, entities, input);

  await expect(page.getByRole("heading", { name: expected.membershipsHeading })).toBeVisible();
  await expect(page.getByTestId("app-shell-content").getByText(expected.firstName).first()).toBeVisible();
});

test("logout then login flow works", async ({ page, entities }) => {
  const input = buildRegisterUserInput();

  const registered = await registerUser(page, entities, input);
  await logoutUser(page, entities);
  const loggedIn = await loginUser(page, entities, {
    login: input.login,
    password: input.password
  });

  expect(loggedIn.body.user.id).toBe(registered.body.user.id);
});

test("self-profile page shows memberships with empty-state support", async ({ page, entities }) => {
  const input = buildRegisterUserInput();
  const expected = {
    emptyMembershipsMessage: "Нет членств"
  };

  await registerUser(page, entities, input);

  await expect(page.getByText(expected.emptyMembershipsMessage)).toBeVisible();
});

test("session persists after page reload while cookie is valid", async ({ page, entities }) => {
  const input = buildRegisterUserInput();

  const registered = await registerUser(page, entities, input);
  const reloaded = await reloadWithActiveSession(page, entities);

  expect(reloaded.userId).toBe(registered.body.user.id);
  await expect(page.getByTestId("app-shell-content").getByText(input.login)).toBeVisible();
});

test("profile edit flow saves and reflects updated values", async ({ page, entities }) => {
  await loginUser(page, entities, {
    login: "admin",
    password: "admin"
  });

  const updatedFirstName = `Admin-${Date.now()}`;
  const updatedAbout = `Updated via e2e ${Date.now()}`;

  const updated = await updateOwnProfile(page, entities, {
    firstName: updatedFirstName,
    about: updatedAbout
  });

  expect(updated.updated.first_name).toBe(updatedFirstName);
  expect(updated.updated.about).toBe(updatedAbout);
});
