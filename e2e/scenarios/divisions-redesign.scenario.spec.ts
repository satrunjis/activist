import { loginUser } from "../actions/auth.actions";
import { expect, test } from "../fixtures/base.fixture";

const ADMIN = {
  login: "admin",
  password: "admin"
} as const;

test("SC-10-3/10-10: division redesign drilldown, semantic zoom, lazy-cache and viewport footprint", async ({
  page,
  entities
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await loginUser(page, entities, ADMIN);

  await expect(page.getByTestId("nav-home")).toBeVisible();
  await expect(page.getByTestId("division-tree-viewport")).toBeVisible();

  await page.getByTestId("division-node-seed-ops").evaluate((node) => (node as HTMLButtonElement).click());
  await expect(page.getByTestId("division-right-panel")).toBeVisible();
  await page.getByTestId("panel-tab-positions").click();
  await expect(page.getByTestId("position-list-panel")).toBeVisible();
  const firstPosition = page.locator('[data-testid^="position-row-"]').first();
  await expect(firstPosition).toBeVisible();

  const membersResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "GET" &&
      /\/api\/v1\/positions\/[^/]+\/members/.test(response.url()) &&
      response.status() === 200
  );
  await firstPosition.evaluate((node) => (node as HTMLElement).click());
  await membersResponse;
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByTestId("position-members-list")).toBeVisible();

  const viewport = page.getByTestId("division-tree-viewport");
  const viewportBox = await viewport.boundingBox();
  if (!viewportBox) {
    throw new Error("Expected division tree viewport bounding box for zoom interaction");
  }
  await page.mouse.move(viewportBox.x + viewportBox.width / 2, viewportBox.y + viewportBox.height / 2);
  await page.mouse.wheel(0, -1600);
  await page.mouse.wheel(0, -1600);
  await expect(page.getByRole("button", { name: "Увеличить" })).toBeVisible();

  const footprint = await page.evaluate(() => {
    const viewport = document.querySelector('[data-testid="division-tree-viewport"]') as HTMLElement | null;
    const routeContent = document.querySelector('[data-testid="app-shell-content"]') as HTMLElement | null;
    if (!viewport || !routeContent) {
      return { ok: false, ratio: 0, height: 0, minHeight: 0 };
    }
    const viewportRect = viewport.getBoundingClientRect();
    const routeRect = routeContent.getBoundingClientRect();
    return {
      ok: true,
      ratio: viewportRect.width / routeRect.width,
      height: viewportRect.height,
      minHeight: window.innerHeight - 160
    };
  });

  expect(footprint.ok).toBe(true);
  expect(footprint.ratio).toBeGreaterThanOrEqual(0.62);
  expect(footprint.height).toBeGreaterThanOrEqual(footprint.minHeight);
});
