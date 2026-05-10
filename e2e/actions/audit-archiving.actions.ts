import { expect, type Page } from "@playwright/test";
import type { CreatedEntities } from "../support/teardown";
import { ensureLoggedIn, readBrowserSession } from "../support/ui-auth";

type Credentials = {
  login: string;
  password: string;
};

type OpenAuditLogInput = {
  auth: Credentials;
};

type AuditFilterInput = {
  eventType?: string;
  subjectType?: string;
  subjectId?: string;
};

type ExpectAuditEntryInput = {
  eventType: string;
  subjectType: string;
  subjectId: string;
  actorId?: string;
};

type AssertAuditForbiddenInput = {
  auth: Credentials;
};

async function login(page: Page, auth: Credentials): Promise<void> {
  await ensureLoggedIn(page, auth);
}

export async function openAuditLog(
  page: Page,
  _entities: CreatedEntities,
  input: OpenAuditLogInput
): Promise<{ actorId: string }> {
  await login(page, input.auth);

  const actorId = (await readBrowserSession(page))?.user?.id?.trim();
  if (!actorId) {
    throw new Error("openAuditLog failed: authenticated user id is not present in session");
  }

  const navAudit = page.getByTestId("nav-audit-log");
  await expect(navAudit).toBeVisible();
  await navAudit.click();

  await expect(page).toHaveURL(/\/audit-log$/);
  await expect(page.getByTestId("audit-log-page")).toBeVisible();

  return { actorId };
}

async function filterAuditLog(page: Page, input: AuditFilterInput): Promise<void> {
  await expect(page.getByTestId("audit-log-page")).toBeVisible();

  if (input.eventType !== undefined) {
    await page.getByTestId("audit-filter-event_type").fill(input.eventType);
  }
  if (input.subjectType !== undefined) {
    await page.getByTestId("audit-filter-subject_type").fill(input.subjectType);
  }
  if (input.subjectId !== undefined) {
    await page.getByTestId("audit-filter-subject_id").fill(input.subjectId);
  }

  await page.getByTestId("audit-filter-submit").click();
}

export async function expectAuditLogEntryVisible(
  page: Page,
  _entities: CreatedEntities,
  input: ExpectAuditEntryInput
): Promise<void> {
  await filterAuditLog(page, {
    eventType: input.eventType,
    subjectType: input.subjectType,
    subjectId: input.subjectId
  });

  const table = page.getByTestId("audit-log-table");
  await expect(table).toBeVisible();
  await expect(table.getByText(input.eventType)).toBeVisible();
  await expect(table.getByText(input.subjectType)).toBeVisible();
  await expect(table.getByText(input.subjectId)).toBeVisible();
  if (input.actorId) {
    await expect(table.getByText(input.actorId)).toBeVisible();
  }
}

export async function assertAuditLogForbiddenForActor(
  page: Page,
  _entities: CreatedEntities,
  input: AssertAuditForbiddenInput
): Promise<void> {
  await login(page, input.auth);

  await expect(page.getByTestId("nav-audit-log")).toHaveCount(0);
  await page.goto("/audit-log");

  await expect(page).toHaveURL(/\/audit-log$/);
  await expect(page.getByTestId("audit-log-forbidden")).toBeVisible();
  await expect(page.getByTestId("audit-log-forbidden")).toContainText("Операция запрещена.");
}
