import { expect } from "@playwright/test";
import {
  assignMemberToPosition,
  createPosition,
  enforceAssignmentLeadershipAcl,
  enforceScopedMutationAcl,
  removeMembership
} from "../actions/positions.actions";
import { loginUser } from "../actions/auth.actions";
import { viewOwnMemberships } from "../actions/profile.actions";
import { test } from "../fixtures/base.fixture";

const TEST_OPERATOR = {
  login: "test0",
  password: "test0"
} as const;

test("POS-01: create position via declarative action", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops",
    title: "Declarative Secretary",
    roleName: "Seed Standard"
  };
  const expected = {
    visibleTitle: input.title
  };

  await createPosition(page, entities, input);

  await expect(page.getByText(expected.visibleTitle).first()).toBeVisible();
});

test("MEM-01: assign member via declarative action", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops",
    positionId: "seed-pos-manager",
    userId: "seed-target"
  };
  const expected = {
    assignButtonEnabled: true
  };

  await assignMemberToPosition(page, entities, input);

  if (expected.assignButtonEnabled) {
    await expect(page.getByTestId(`assign-member-${input.positionId}`)).toBeEnabled();
  } else {
    await expect(page.getByTestId(`assign-member-${input.positionId}`)).toBeDisabled();
  }
});

test("ADM-01: leadership assignment forbidden shows leadership ACL message", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops",
    positionId: "seed-pos-leader",
    userId: "seed-target"
  };
  const expected = {
    message: "Пользователь уже назначен на эту должность."
  };

  await enforceAssignmentLeadershipAcl(page, entities, input);

  await expect(page.getByText(expected.message)).toBeVisible();
});

test("ACL-01: create mutation forbidden keeps read-only views available", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops-field",
    mutation: "create-position" as const,
    createPosition: {
      title: "Blocked by Scope",
      roleName: "Seed Standard"
    }
  };
  const expected = {
    message: "Операция недоступна в текущем контексте.",
    selectedDivisionLabel: "Field Ops",
    visiblePositionTitle: "Field Coordinator"
  };

  await enforceScopedMutationAcl(page, entities, input);

  await expect(page.getByTestId("assign-member-seed-pos-field")).toBeVisible();
  await expect(page.getByTestId("division-right-panel").getByText(expected.selectedDivisionLabel)).toBeVisible();
  await expect(page.getByTestId("division-right-panel").getByText(expected.visiblePositionTitle)).toBeVisible();
});

test("POS-GAP-01: create position with max_count", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops",
    title: "Deputy With Cap",
    roleName: "Seed Standard",
    maxCount: 2
  };
  const expected = {
    visibleTitle: input.title
  };

  await createPosition(page, entities, input);

  await expect(page.getByText(expected.visibleTitle).first()).toBeVisible();
});

test("MEM-GAP-02: duplicate assignment shows 409 UX message", async ({ page, entities }) => {
  const firstAssignment = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops",
    positionId: "seed-pos-manager",
    userId: "seed-target"
  };
  const duplicateAssignment = {
    ...firstAssignment,
    expectedOutcome: "duplicate_conflict" as const
  };
  const expected = {
    duplicateMessage: "Пользователь уже назначен на эту должность."
  };

  await assignMemberToPosition(page, entities, firstAssignment);
  await assignMemberToPosition(page, entities, duplicateAssignment);

  await expect(page.getByText(expected.duplicateMessage)).toBeVisible();
});

test("MEM-GAP-03: generic scope-forbidden assign path", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    divisionId: "seed-ops-field",
    mutation: "assign-member" as const,
    assignMember: {
      positionId: "seed-pos-field",
      userId: "seed-target"
    }
  };
  const expected = {
    message: "Операция недоступна в текущем контексте."
  };

  await enforceScopedMutationAcl(page, entities, input);

  await expect(page.getByText(expected.message)).toBeVisible();
});

test("MEM-02: remove membership shows inline confirm", async ({ page, entities }) => {
  const input = {
    auth: TEST_OPERATOR,
    positionId: "seed-pos-manager",
    userId: "seed-test0",
    mode: "preview" as const
  };

  const removed = await removeMembership(page, entities, input);

  await expect(page.getByRole("dialog", { name: "Исключить из членства?" })).toBeVisible();
});

test("VIEW-01: memberships list shows role_name", async ({ page, entities }) => {
  await loginUser(page, entities, TEST_OPERATOR);
  const viewed = await viewOwnMemberships(page, entities);

  expect(viewed.isEmpty).toBe(false);
  const membershipsCard = page.getByTestId("memberships-card");
  await expect(membershipsCard).toContainText(/Ops Coordinator/);
  await expect(membershipsCard).toContainText(/Seed Manager \(Current Division\)/);
});
