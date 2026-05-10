import { randomUUID } from "node:crypto";
import {
  assertAuditLogForbiddenForActor,
  expectAuditLogEntryVisible,
  openAuditLog
} from "../actions/audit-archiving.actions";
import { assignMemberToPosition } from "../actions/positions.actions";
import { expect, test } from "../fixtures/base.fixture";
import type { SeedHelpers } from "../support/seed";
import { buildRegisterUserInput } from "../support/test-data";

const ADMIN = {
  login: "admin",
  password: "admin"
} as const;

function unique(value: string): string {
  return `${value}-${randomUUID().replace(/-/g, "").slice(0, 8)}`;
}

async function seedAssignmentFixture(seed: SeedHelpers) {
  await seed.loginUser(ADMIN);

  const division = await seed.createDivision({
    shortName: unique("audit-div")
  });
  const position = await seed.createPosition({
    divisionId: division.id,
    title: unique("Audit Position"),
    roleName: "Seed Standard"
  });
  const member = buildRegisterUserInput();
  const registered = await seed.registerUser(member);
  const userId = registered.body.user.id;
  await seed.loginUser(ADMIN);
  return {
    divisionId: division.id,
    positionId: position.id,
    userId
  };
}

async function seedAuditViewer(seed: SeedHelpers) {
  await seed.loginUser(ADMIN);

  const viewerRole = await seed.createRole({
    name: unique("audit-viewer-role"),
    permissions: ["can_view_audit_log"]
  });
  const viewerDivision = await seed.createDivision({
    shortName: unique("audit-viewer-div")
  });
  const viewerSeat = await seed.createPosition({
    divisionId: viewerDivision.id,
    title: unique("Audit Viewer Seat"),
    roleId: viewerRole.id
  });
  const viewerInput = buildRegisterUserInput();
  const viewer = await seed.registerUser(viewerInput);

  await seed.loginUser(ADMIN);
  await seed.assignMembership({
    userId: viewer.body.user.id,
    positionId: viewerSeat.id
  });

  return {
    auth: {
      login: viewerInput.login,
      password: viewerInput.password
    }
  };
}

test("D-10 assign membership writes EventLog entry with expected metadata", async ({ page, entities, seed }) => {
  const fixture = await seedAssignmentFixture(seed);
  const viewer = await seedAuditViewer(seed);
  const expected = {
    eventType: "position_assigned",
    subjectType: "membership",
    subjectId: `${fixture.userId}:${fixture.positionId}`
  };

  await assignMemberToPosition(page, entities, {
    auth: ADMIN,
    divisionId: fixture.divisionId,
    positionId: fixture.positionId,
    userId: fixture.userId
  });

  await openAuditLog(page, entities, { auth: viewer.auth });
  await expectAuditLogEntryVisible(page, entities, {
    eventType: expected.eventType,
    subjectType: expected.subjectType,
    subjectId: expected.subjectId
  });
});

test("D-11 audit log filter by event_type via browser table", async ({ page, entities, seed }) => {
  const fixture = await seedAssignmentFixture(seed);
  const viewer = await seedAuditViewer(seed);

  await assignMemberToPosition(page, entities, {
    auth: ADMIN,
    divisionId: fixture.divisionId,
    positionId: fixture.positionId,
    userId: fixture.userId
  });

  await openAuditLog(page, entities, { auth: viewer.auth });
  await expectAuditLogEntryVisible(page, entities, {
    eventType: "position_assigned",
    subjectType: "membership",
    subjectId: `${fixture.userId}:${fixture.positionId}`
  });
});

test("D-12 audit log filter by subject_id via browser table", async ({ page, entities, seed }) => {
  const fixture = await seedAssignmentFixture(seed);
  const viewer = await seedAuditViewer(seed);

  await assignMemberToPosition(page, entities, {
    auth: ADMIN,
    divisionId: fixture.divisionId,
    positionId: fixture.positionId,
    userId: fixture.userId
  });

  await openAuditLog(page, entities, { auth: viewer.auth });
  await expectAuditLogEntryVisible(page, entities, {
    eventType: "position_assigned",
    subjectType: "membership",
    subjectId: `${fixture.userId}:${fixture.positionId}`
  });
});

test("D-13 audit log defaults to limit=50 and supports paging controls", async ({ page, entities, seed }) => {
  const fixture = await seedAssignmentFixture(seed);
  const viewer = await seedAuditViewer(seed);

  await assignMemberToPosition(page, entities, {
    auth: ADMIN,
    divisionId: fixture.divisionId,
    positionId: fixture.positionId,
    userId: fixture.userId
  });

  await openAuditLog(page, entities, { auth: viewer.auth });

  await expect(page.getByTestId("audit-pagination-state")).toContainText(/Записи/);
  await expect(page.getByTestId("audit-pagination-prev")).toBeDisabled();
});

test("D-14 eventlog endpoint denies actor without ViewAuditLog permission", async ({ page, entities, seed }) => {
  await seed.loginUser(ADMIN);

  const restrictedRole = await seed.createRole({
    name: unique("no-audit-role"),
    permissions: ["can_manage_positions"]
  });
  const division = await seed.createDivision({
    shortName: unique("d14-div")
  });
  const seat = await seed.createPosition({
    divisionId: division.id,
    title: unique("D14 Seat"),
    roleId: restrictedRole.id
  });
  const actorInput = buildRegisterUserInput();
  const actor = await seed.registerUser(actorInput);
  await seed.loginUser(ADMIN);
  await seed.assignMembership({
    userId: actor.body.user.id,
    positionId: seat.id
  });

  await assertAuditLogForbiddenForActor(page, entities, {
    auth: {
      login: actorInput.login,
      password: actorInput.password
    }
  });
});
