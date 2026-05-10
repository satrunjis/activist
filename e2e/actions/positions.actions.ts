import { expect, type Locator, type Page } from "@playwright/test";
import { trackCreated, type CreatedEntities } from "../support/teardown";
import { ensureLoggedIn, readBrowserSession } from "../support/ui-auth";

const FORBIDDEN_SCOPE_MESSAGE = "Операция недоступна в текущем контексте.";
const FORBIDDEN_LEADERSHIP_MESSAGE = "Пользователь уже назначен на эту должность.";
const DUPLICATE_ASSIGNMENT_MESSAGE = "Пользователь уже назначен на эту должность.";

type LoginInput = {
  login: string;
  password: string;
};

export type CreatePositionInput = {
  auth: LoginInput;
  divisionId: string;
  title: string;
  roleName: string;
  maxCount?: number;
};

export type CreatePositionResult = {
  id: string;
  title: string;
  divisionId: string;
};

export type AssignMemberToPositionInput = {
  auth: LoginInput;
  divisionId: string;
  positionId: string;
  userId: string;
  expectedOutcome?: "success" | "duplicate_conflict";
};

export type AssignMemberToPositionResult = {
  positionId: string;
  userId: string;
  membershipKey?: string;
};

export type EnforceAssignmentLeadershipAclInput = {
  auth: LoginInput;
  divisionId: string;
  positionId: string;
  userId: string;
};

export type EnforceScopedMutationAclInput = {
  auth: LoginInput;
  divisionId: string;
  mutation: "create-position" | "assign-member";
  createPosition?: {
    title: string;
    roleName: string;
    maxCount?: number;
  };
  assignMember?: {
    positionId: string;
    userId: string;
  };
};

export type RemoveMembershipInput = {
  auth: LoginInput;
  positionId: string;
  userId?: string;
  mode?: "preview" | "confirm";
};

export type RemoveMembershipResult = {
  positionId: string;
  userId: string;
  mode: "preview" | "confirm";
};

async function login(page: Page, auth: LoginInput): Promise<void> {
  await ensureLoggedIn(page, auth);
}

async function ensureDivisionNodeVisible(page: Page, divisionId: string): Promise<void> {
  const node = page.getByTestId(`division-node-${divisionId}`);

  for (let attempt = 0; attempt < 20; attempt += 1) {
    const nodeCount = await node.count();
    const nodeVisible = nodeCount > 0 ? await node.first().isVisible() : false;
    if (nodeVisible) {
      return;
    }

    await page.waitForTimeout(200);
  }
}

async function clickHTMLElement(locator: Locator): Promise<void> {
  await locator.evaluate((element) => {
    (element as HTMLElement).click();
  });
}

async function selectDivision(page: Page, divisionId: string): Promise<void> {
  await expect(page.getByTestId("division-tree")).toBeVisible();
  await expect(page.getByTestId("division-tree-viewport")).toBeVisible();
  await ensureDivisionNodeVisible(page, divisionId);
  const node = page.locator(`[data-testid="division-node-${divisionId}"]:visible`).first();
  await expect(node).toBeVisible();
  await clickHTMLElement(node);
  await page.getByTestId("panel-tab-positions").click();
  await expect(page.getByTestId("position-list-panel")).toBeVisible();
}

export async function createPosition(
  page: Page,
  entities: CreatedEntities,
  input: CreatePositionInput
): Promise<CreatePositionResult> {
  await login(page, input.auth);
  await selectDivision(page, input.divisionId);

  await page.getByTestId("create-position").click();
  await expect(page.getByTestId("create-position-form")).toBeVisible();

  const createResponsePromise = page.waitForResponse((response) => {
    return (
      response.request().method() === "POST" &&
      /\/api\/v1\/divisions\/[^/]+\/positions$/.test(response.url()) &&
      response.status() === 201
    );
  });

  await page.getByTestId("position-title").fill(input.title);
  await page.getByTestId("position-role").click();
  await clickHTMLElement(page.getByRole("option", { name: input.roleName }).first());
  if (input.maxCount !== undefined) {
    await page.getByTestId("position-max-count").fill(String(input.maxCount));
  }
  await page.getByTestId("create-position-submit").click();

  const createResponse = await createResponsePromise;
  await expect(page.getByTestId("create-position-form")).not.toBeVisible();

  const body = (await createResponse.json()) as { id?: string; title?: string; division_id?: string };
  const id = body.id?.trim();
  if (!id) {
    throw new Error("createPosition failed: API response does not contain created position id");
  }

  trackCreated(entities, "positions", id);

  return {
    id,
    title: body.title?.trim() || input.title,
    divisionId: body.division_id?.trim() || input.divisionId
  };
}

export async function assignMemberToPosition(
  page: Page,
  entities: CreatedEntities,
  input: AssignMemberToPositionInput
): Promise<AssignMemberToPositionResult> {
  await login(page, input.auth);
  await selectDivision(page, input.divisionId);

  await page.getByTestId(`assign-member-${input.positionId}`).click();
  await expect(page.getByTestId("assign-member-form")).toBeVisible();

  await page.getByTestId("assign-user-id").fill(input.userId);
  await page.getByTestId("assign-submit").click();

  const expectedOutcome = input.expectedOutcome ?? "success";
  if (expectedOutcome === "duplicate_conflict") {
    await expect(page.getByText(DUPLICATE_ASSIGNMENT_MESSAGE)).toBeVisible();
    return {
      positionId: input.positionId,
      userId: input.userId
    };
  }

  const duplicateMessage = page.getByText(DUPLICATE_ASSIGNMENT_MESSAGE);
  if (await duplicateMessage.waitFor({ state: "visible", timeout: 3000 }).then(() => true).catch(() => false)) {
    await page.getByRole("button", { name: "Отмена" }).click();
    await expect(page.getByTestId("position-list-panel")).toBeVisible();
    return {
      positionId: input.positionId,
      userId: input.userId,
      membershipKey: `${input.userId}::${input.positionId}`
    };
  }

  await expect(page.getByTestId("assign-member-form")).not.toBeVisible();

  const membershipKey = `${input.userId}::${input.positionId}`;
  trackCreated(entities, "memberships", membershipKey);

  return {
    positionId: input.positionId,
    userId: input.userId,
    membershipKey
  };
}

export async function enforceAssignmentLeadershipAcl(
  page: Page,
  _entities: CreatedEntities,
  input: EnforceAssignmentLeadershipAclInput
): Promise<void> {
  await login(page, input.auth);
  await selectDivision(page, input.divisionId);

  await page.getByTestId(`assign-member-${input.positionId}`).click();
  await expect(page.getByTestId("assign-member-form")).toBeVisible();

  await page.getByTestId("assign-user-id").fill(input.userId);
  await page.getByTestId("assign-submit").click();

  await expect(page.getByText(FORBIDDEN_LEADERSHIP_MESSAGE)).toBeVisible();
}

export async function enforceScopedMutationAcl(
  page: Page,
  _entities: CreatedEntities,
  input: EnforceScopedMutationAclInput
): Promise<void> {
  await login(page, input.auth);
  await selectDivision(page, input.divisionId);

  if (input.mutation === "create-position") {
    if (!input.createPosition) {
      throw new Error("enforceScopedMutationAcl requires createPosition payload for create-position mutation");
    }

    await page.getByTestId("create-position").click();
    await expect(page.getByTestId("create-position-form")).toBeVisible();

    await page.getByTestId("position-title").fill(input.createPosition.title);
    await page.getByTestId("position-role").click();
    await clickHTMLElement(page.getByRole("option", { name: input.createPosition.roleName }).first());
    if (input.createPosition.maxCount !== undefined) {
      await page.getByTestId("position-max-count").fill(String(input.createPosition.maxCount));
    }
    await page.getByTestId("create-position-submit").click();

    await expect(page.getByText(FORBIDDEN_SCOPE_MESSAGE).first()).toBeVisible();
    await page.getByRole("button", { name: "Отмена" }).click();
    await expect(page.getByTestId("position-list-panel")).toBeVisible();
    return;
  }

  if (!input.assignMember) {
    throw new Error("enforceScopedMutationAcl requires assignMember payload for assign-member mutation");
  }

  await page.getByTestId(`assign-member-${input.assignMember.positionId}`).click();
  await expect(page.getByTestId("assign-member-form")).toBeVisible();

  await page.getByTestId("assign-user-id").fill(input.assignMember.userId);
  await page.getByTestId("assign-submit").click();

  await expect(page.getByText(FORBIDDEN_SCOPE_MESSAGE)).toBeVisible();
}

export async function removeMembership(
  page: Page,
  _entities: CreatedEntities,
  input: RemoveMembershipInput
): Promise<RemoveMembershipResult> {
  await login(page, input.auth);

  const sessionUserId = (await readBrowserSession(page))?.user?.id?.trim();
  if (!sessionUserId) {
    throw new Error("removeMembership failed: cannot resolve authenticated user id from session toolbar");
  }
  const userId = input.userId ?? sessionUserId;
  const mode = input.mode ?? "preview";

  await page.getByTestId("nav-profile").click();
  await expect(page.getByTestId("memberships-card")).toBeVisible();
  const removeButton = page.getByTestId(`remove-membership-${input.positionId}-${userId}`);
  await expect(removeButton).toBeVisible();
  await removeButton.click();

  const dialog = page.getByRole("dialog", { name: "Исключить из членства?" });
  await expect(dialog).toBeVisible();
  const confirmButton = dialog.getByRole("button", { name: "Исключить" });

  if (mode === "confirm") {
    await confirmButton.click();
    await expect(page.getByTestId(`remove-membership-${input.positionId}-${userId}`)).toHaveCount(0);
  }

  return {
    positionId: input.positionId,
    userId,
    mode
  };
}
