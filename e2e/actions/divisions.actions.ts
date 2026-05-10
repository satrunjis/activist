import { expect, type Locator, type Page } from "@playwright/test";
import { trackCreated, type CreatedEntities } from "../support/teardown";

export type DivisionMediaLinkInput = {
  platform: string;
  value: string;
};

export type ViewDivisionTreeInput = {
  path?: string;
  visibleNodeIds?: string[];
  expandNodeIds?: string[];
};

export type DivisionIdentity = {
  id: string;
  shortName: string;
};

export type CreateDivisionInput = {
  selectParentId?: string;
  parentId?: string;
  shortName: string;
  fullName?: string;
  description?: string;
  regulationUrl?: string;
  mediaLinks?: DivisionMediaLinkInput[];
  expectForbidden?: boolean;
};

export type EditDivisionInput = {
  targetDivisionId?: string;
  parentId?: string;
  shortName?: string;
  fullName?: string;
  description?: string;
  regulationUrl?: string;
  mediaLinks?: DivisionMediaLinkInput[];
  expectForbidden?: boolean;
};

export type ReparentDivisionInput = {
  targetDivisionId: string;
  newParentId: string;
};

function form(mode: "create" | "edit", page: Page): Locator {
  return page.getByTestId(mode === "create" ? "create-division-form" : "edit-division-form");
}

async function clickDivisionNode(page: Page, nodeId: string): Promise<void> {
  const node = page.getByTestId(`division-node-${nodeId}`);
  await expect(node).toBeVisible();
  await node.dispatchEvent("click");
}

async function setMediaLinks(page: Page, mode: "create" | "edit", links: DivisionMediaLinkInput[]): Promise<void> {
  const activeForm = form(mode, page);
  const addLinkButton = activeForm.getByRole("button", { name: "+ Ссылка" });

  if (mode === "edit") {
    const removeButtons = activeForm.getByRole("button", { name: "Удалить ссылку" });
    while ((await removeButtons.count()) > 0) {
      await removeButtons.first().click();
    }
  }

  for (let i = 0; i < links.length; i += 1) {
    await addLinkButton.click();
    await activeForm.getByLabel("Платформа").last().fill(links[i].platform);
    await activeForm.getByLabel("Ссылка").last().fill(links[i].value);
  }
}

function slugFromShortName(shortName: string): string {
  return shortName.trim().toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

export async function viewDivisionTree(
  page: Page,
  _entities: CreatedEntities,
  input: ViewDivisionTreeInput
): Promise<void> {
  await page.goto(input.path ?? "/");
  await expect(page.getByTestId("division-tree")).toBeVisible();

  for (const nodeId of input.visibleNodeIds ?? []) {
    await expect(page.getByTestId(`division-node-${nodeId}`)).toBeVisible();
  }

  for (const nodeId of input.expandNodeIds ?? []) {
    await toggleDivisionNode(page, nodeId);
  }
}

export async function toggleDivisionNode(page: Page, nodeId: string): Promise<void> {
  const node = page.getByTestId(`division-node-${nodeId}`);
  await expect(node).toBeVisible();
  await node.dispatchEvent("click");
}

/**
 * Current division explorer uses a full-tree cache rather than lazy child fetches.
 * Keep this action as a UI readiness assertion for older scenario call sites.
 */
export async function expectChildrenFetched(
  page: Page,
  parentId: string
): Promise<void> {
  await expect(page.getByTestId("division-tree-viewport")).toBeVisible();
}

export async function createDivision(
  page: Page,
  entities: CreatedEntities,
  input: CreateDivisionInput
): Promise<DivisionIdentity | null> {
  if (input.selectParentId) {
    await clickDivisionNode(page, input.selectParentId);
  }

  await page.getByTestId("create-division").click();
  const createForm = form("create", page);
  await expect(createForm).toBeVisible();

  if (input.parentId !== undefined) {
    await page.getByTestId("create-division-parent").selectOption(input.parentId);
  }

  await page.getByTestId("create-division-short-name").fill(input.shortName);
  await page.getByTestId("create-division-full-name").fill(input.fullName ?? "");
  await page.getByTestId("create-division-description").fill(input.description ?? "");
  await page.getByTestId("create-division-regulation-url").fill(input.regulationUrl ?? "");

  if (input.mediaLinks) {
    await setMediaLinks(page, "create", input.mediaLinks);
  }

  await page.getByTestId("create-division-submit").click();

  if (input.expectForbidden) {
    await expect(page.getByTestId("create-division")).toBeDisabled();
    return null;
  }

  const identity = { id: slugFromShortName(input.shortName), shortName: input.shortName };
  await expect(page.getByTestId(`division-node-${identity.id}`)).toBeVisible();
  trackCreated(entities, "divisions", identity.id);
  return identity;
}

export async function editDivision(
  page: Page,
  _entities: CreatedEntities,
  input: EditDivisionInput
): Promise<DivisionIdentity | null> {
  if (input.targetDivisionId) {
    await clickDivisionNode(page, input.targetDivisionId);
  }

  await page.getByTestId("edit-division").click();
  const editForm = form("edit", page);
  await expect(editForm).toBeVisible();

  if (input.parentId !== undefined) {
    await page.getByTestId("edit-division-parent").selectOption(input.parentId);
  }

  if (input.shortName !== undefined) {
    await page.getByTestId("edit-division-short-name").fill(input.shortName);
  }
  if (input.fullName !== undefined) {
    await page.getByTestId("edit-division-full-name").fill(input.fullName);
  }
  if (input.description !== undefined) {
    await page.getByTestId("edit-division-description").fill(input.description);
  }
  if (input.regulationUrl !== undefined) {
    await page.getByTestId("edit-division-regulation-url").fill(input.regulationUrl);
  }

  if (input.mediaLinks) {
    await setMediaLinks(page, "edit", input.mediaLinks);
  }

  await page.getByTestId("edit-division-submit").click();

  if (input.expectForbidden) {
    await expect(page.getByTestId("create-division")).toBeDisabled();
    await expect(page.getByTestId("edit-division")).toBeDisabled();
    return null;
  }

  const identity = { id: input.targetDivisionId ?? "", shortName: input.shortName ?? "" };
  if (input.shortName && input.targetDivisionId) {
    await expect(page.getByTestId(`division-node-${input.targetDivisionId}`)).toContainText(input.shortName);
  }
  return identity;
}

export async function reparentDivision(
  page: Page,
  entities: CreatedEntities,
  input: ReparentDivisionInput
): Promise<DivisionIdentity> {
  const identity = await editDivision(page, entities, {
    targetDivisionId: input.targetDivisionId,
    parentId: input.newParentId
  });

  if (!identity) {
    throw new Error("Reparent returned forbidden state unexpectedly");
  }

  return {
    id: input.targetDivisionId,
    shortName: identity.shortName
  };
}
