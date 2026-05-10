import AxeBuilder from "@axe-core/playwright";

import { loginUser } from "../actions/auth.actions";
import { expect, test } from "../fixtures/base.fixture";

const ADMIN = {
  login: "admin",
  password: "admin"
} as const;

const CURRENT_UI_KNOWN_AXE_RULES = new Set([
  "color-contrast",
  "scrollable-region-focusable"
]);

test("redesign accessibility smoke (critical/serious)", async ({ page, entities }) => {
  await loginUser(page, entities, ADMIN);

  await expect(page.getByTestId("app-shell-content")).toBeVisible();
  const homeA11y = await new AxeBuilder({ page }).analyze();
  const homeBlocking = homeA11y.violations.filter((violation) =>
    !CURRENT_UI_KNOWN_AXE_RULES.has(violation.id) && (violation.impact === "critical" || violation.impact === "serious")
  );
  expect(homeBlocking, JSON.stringify(homeBlocking, null, 2)).toEqual([]);

  await page.getByTestId("nav-audit-log").click();
  await expect(page.getByTestId("audit-log-page")).toBeVisible();
  const auditA11y = await new AxeBuilder({ page }).analyze();
  const auditBlocking = auditA11y.violations.filter((violation) =>
    !CURRENT_UI_KNOWN_AXE_RULES.has(violation.id) && (violation.impact === "critical" || violation.impact === "serious")
  );
  expect(auditBlocking, JSON.stringify(auditBlocking, null, 2)).toEqual([]);
});
