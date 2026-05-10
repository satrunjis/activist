import { buildRegisterUserInput } from "../support/test-data";
import { registerUser } from "../actions/auth.actions";
import { test, expect } from "../fixtures/base.fixture";

test("registers a user via browser and reaches authenticated state", async ({ page, entities }) => {
  const input = buildRegisterUserInput();
  const expected = {
    firstName: input.firstName,
    login: input.login
  };

  const created = await registerUser(page, entities, input);

  expect(created.body.user.login).toBe(expected.login);
  await expect(page.getByTestId("app-shell-content").getByText(expected.firstName).first()).toBeVisible();
});
