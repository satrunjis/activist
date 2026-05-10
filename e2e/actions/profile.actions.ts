import { expect, type Page } from "@playwright/test";
import { trackCreated, type CreatedEntities } from "../support/teardown";
import { readBrowserSession } from "../support/ui-auth";

export type RegisterOwnProfileInput = {
  login: string;
  password: string;
  firstName: string;
  gradebookNumber: string;
  groupNumber: string;
  institute: string;
  birthDate: string;
};

export type UpdateOwnProfileInput = {
  auth?: RegisterOwnProfileInput;
  firstName?: string;
  lastName?: string;
  middleName?: string;
  gradebookNumber?: string;
  groupNumber?: string;
  institute?: string;
  birthDate?: string;
  phone?: string;
  about?: string;
};

export type ManageProfileSocialLinksInput = {
  auth?: RegisterOwnProfileInput;
  reset?: boolean;
  socialLinks: Array<{ platform: string; value: string }>;
};

export type ViewOwnMembershipsInput = {
  auth?: RegisterOwnProfileInput;
};

export type RemoveOwnMembershipInput = {
  auth?: RegisterOwnProfileInput;
  positionId: string;
  userId?: string;
};

export type CancelMembershipRemovalInput = {
  auth?: RegisterOwnProfileInput;
  positionId: string;
  userId?: string;
  mode: "button" | "timeout";
};

type SessionIdentity = {
  userId: string;
  login?: string;
};

type MembershipIdentity = {
  key: string;
  userId: string;
  positionId: string;
  removeButtonTestId: string;
};

async function resolveSessionIdentity(page: Page): Promise<SessionIdentity | null> {
  const session = await readBrowserSession(page);
  const userId = session?.user?.id?.trim();
  if (!userId) {
    return null;
  }
  return {
    userId,
    login: session?.user?.login?.trim() || undefined
  };
}

async function ensureAuthenticatedOwnProfile(
  page: Page,
  entities: CreatedEntities,
  auth?: RegisterOwnProfileInput
): Promise<SessionIdentity> {
  const active = await resolveSessionIdentity(page);
  if (active) {
    await page.goto("/");
    await page.getByTestId("nav-profile").click();
    await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();
    return active;
  }

  if (!auth) {
    throw new Error("Authentication precondition missing: pass input.auth for profile actions in anonymous state");
  }

  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Добро пожаловать" })).toBeVisible();
  await page.getByRole("tab", { name: "Регистрация" }).click();
  const registerForm = page.locator('form:has(button:has-text("Зарегистрироваться"))');
  await registerForm.locator('input[name="login"]').fill(auth.login);
  await registerForm.locator('input[name="password"]').fill(auth.password);
  await registerForm.locator('input[name="first_name"]').fill(auth.firstName);
  await registerForm.locator('input[name="gradebook_number"]').fill(auth.gradebookNumber);
  await registerForm.locator('input[name="group_number"]').fill(auth.groupNumber);
  await registerForm.locator('input[name="institute"]').fill(auth.institute);
  await registerForm.locator('input[name="birth_date"]').fill(auth.birthDate);
  await registerForm.getByRole("button", { name: "Зарегистрироваться" }).click();
  await expect(page.getByRole("button", { name: "Выйти" })).toBeVisible();
  await page.getByTestId("nav-profile").click();
  await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();

  const current = await resolveSessionIdentity(page);
  if (!current) {
    throw new Error("Failed to resolve authenticated self-profile identity after registration");
  }
  trackCreated(entities, "users", current.userId);
  return current;
}

function toFieldMap(input: UpdateOwnProfileInput): Array<{ name: string; value: string | undefined }> {
  return [
    { name: "first_name", value: input.firstName },
    { name: "last_name", value: input.lastName },
    { name: "middle_name", value: input.middleName },
    { name: "gradebook_number", value: input.gradebookNumber },
    { name: "group_number", value: input.groupNumber },
    { name: "institute", value: input.institute },
    { name: "birth_date", value: input.birthDate },
    { name: "phone", value: input.phone },
    { name: "about", value: input.about }
  ];
}

function profileSummaryLabel(name: string): string {
  switch (name) {
    case "first_name":
      return "Имя";
    case "last_name":
      return "Фамилия";
    case "middle_name":
      return "Отчество";
    case "gradebook_number":
      return "Зачётная книжка";
    case "group_number":
      return "Группа";
    case "institute":
      return "Институт";
    case "birth_date":
      return "Дата рождения";
    case "phone":
      return "Телефон";
    case "about":
      return "О себе";
    default:
      return name;
  }
}

async function saveProfileForm(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Сохранить" }).click();
}

function parseMembershipIdentityFromRemoveTestId(testId: string, fallbackUserId: string): MembershipIdentity | null {
  const prefix = "remove-membership-";
  if (!testId.startsWith(prefix)) {
    return null;
  }

  const body = testId.slice(prefix.length);
  let positionId = "";
  let userId = fallbackUserId;

  const explicitSuffix = `-${fallbackUserId}`;
  if (body.endsWith(explicitSuffix)) {
    positionId = body.slice(0, -explicitSuffix.length);
  } else {
    const boundary = body.lastIndexOf("-");
    if (boundary <= 0 || boundary >= body.length - 1) {
      return null;
    }
    positionId = body.slice(0, boundary);
    userId = body.slice(boundary + 1) || fallbackUserId;
  }

  if (!positionId) {
    return null;
  }

  return {
    positionId,
    userId,
    key: `${userId}::${positionId}`,
    removeButtonTestId: testId
  };
}

export async function updateOwnProfile(
  page: Page,
  entities: CreatedEntities,
  input: UpdateOwnProfileInput
): Promise<{ identity: SessionIdentity; updated: Record<string, string> }> {
  const identity = await ensureAuthenticatedOwnProfile(page, entities, input.auth);
  await page.getByRole("button", { name: "Редактировать" }).click();

  const updated: Record<string, string> = {};
  for (const field of toFieldMap(input)) {
    if (field.value === undefined) {
      continue;
    }
    const label = profileSummaryLabel(field.name);
    const fieldRow = page.locator(".meta-block__field").filter({ hasText: label }).first();
    await expect(fieldRow).toBeVisible();
    await fieldRow.locator("input").fill(field.value);
    updated[field.name] = field.value;
  }

  await saveProfileForm(page);

  for (const [name, value] of Object.entries(updated)) {
    const fieldRow = page.locator(".meta-block__field").filter({ hasText: profileSummaryLabel(name) }).first();
    await expect(fieldRow).toContainText(value);
  }

  return {
    identity,
    updated
  };
}

export async function manageProfileSocialLinks(
  page: Page,
  entities: CreatedEntities,
  input: ManageProfileSocialLinksInput
): Promise<{ identity: SessionIdentity; socialLinks: Array<{ platform: string; value: string }> }> {
  const identity = await ensureAuthenticatedOwnProfile(page, entities, input.auth);
  await page.getByRole("button", { name: "Редактировать" }).click();

  if (input.reset) {
    let removeButtons = page.getByRole("button", { name: "Удалить ссылку" });
    while ((await removeButtons.count()) > 0) {
      await removeButtons.first().click();
      removeButtons = page.getByRole("button", { name: "Удалить ссылку" });
    }
  }

  for (const link of input.socialLinks) {
    await page.getByRole("button", { name: "Добавить ссылку" }).click();
    await page.getByLabel("Платформа").last().fill(link.platform);
    await page.getByLabel("Ссылка").last().fill(link.value);
  }

  await saveProfileForm(page);

  for (const link of input.socialLinks) {
    await expect(page.getByRole("link", { name: link.platform })).toBeVisible();
  }

  return {
    identity,
    socialLinks: input.socialLinks
  };
}

export async function viewOwnMemberships(
  page: Page,
  entities: CreatedEntities,
  input: ViewOwnMembershipsInput = {}
): Promise<{ identity: SessionIdentity; memberships: MembershipIdentity[]; isEmpty: boolean }> {
  const identity = await ensureAuthenticatedOwnProfile(page, entities, input.auth);
  const membershipsCard = page.getByTestId("memberships-card");
  await expect(membershipsCard).toBeVisible();

  const removeButtons = membershipsCard.locator('[data-testid^="remove-membership-"]');
  const count = await removeButtons.count();
  const memberships: MembershipIdentity[] = [];
  for (let index = 0; index < count; index += 1) {
    const removeTestId = await removeButtons.nth(index).getAttribute("data-testid");
    if (!removeTestId) {
      continue;
    }
    const parsed = parseMembershipIdentityFromRemoveTestId(removeTestId, identity.userId);
    if (!parsed) {
      continue;
    }
    memberships.push(parsed);
  }

  const emptyState = membershipsCard.getByText("Нет членств");
  const isEmpty = memberships.length === 0;
  if (isEmpty) {
    await expect(emptyState).toBeVisible();
  } else {
    await expect(emptyState).toHaveCount(0);
  }

  return {
    identity,
    memberships,
    isEmpty
  };
}

export async function removeOwnMembership(
  page: Page,
  entities: CreatedEntities,
  input: RemoveOwnMembershipInput
): Promise<{ key: string; positionId: string; userId: string }> {
  const identity = await ensureAuthenticatedOwnProfile(page, entities, input.auth);
  const userId = input.userId ?? identity.userId;
  const key = `${userId}::${input.positionId}`;

  const removeButton = page.getByTestId(`remove-membership-${input.positionId}-${userId}`);
  await expect(removeButton).toBeVisible();
  await removeButton.click();

  const dialog = page.getByRole("dialog", { name: "Исключить из членства?" });
  await expect(dialog).toBeVisible();
  const confirmButton = dialog.getByRole("button", { name: "Исключить" });
  const removeRequest = page.waitForResponse((response) => {
    if (response.request().method() !== "DELETE") {
      return false;
    }
    return response.url().includes(
      `/api/v1/positions/${encodeURIComponent(input.positionId)}/members/${encodeURIComponent(userId)}`
    );
  });
  await confirmButton.click();
  const removeResponse = await removeRequest;
  if (!removeResponse.ok()) {
    throw new Error(`remove membership failed with status ${removeResponse.status()}`);
  }

  await expect(page.getByTestId(`remove-membership-${input.positionId}-${userId}`)).toHaveCount(0);
  trackCreated(entities, "memberships", key);

  return {
    key,
    positionId: input.positionId,
    userId
  };
}

export async function cancelMembershipRemoval(
  page: Page,
  entities: CreatedEntities,
  input: CancelMembershipRemovalInput
): Promise<{ key: string; positionId: string; userId: string; mode: "button" | "timeout" }> {
  const identity = await ensureAuthenticatedOwnProfile(page, entities, input.auth);
  const userId = input.userId ?? identity.userId;
  const key = `${userId}::${input.positionId}`;

  const removeButton = page.getByTestId(`remove-membership-${input.positionId}-${userId}`);
  await expect(removeButton).toBeVisible();
  await removeButton.click();

  const dialog = page.getByRole("dialog", { name: "Исключить из членства?" });
  await expect(dialog).toBeVisible();
  const confirmButton = dialog.getByRole("button", { name: "Исключить" });

  if (input.mode === "button") {
    await dialog.getByRole("button", { name: "Отмена" }).click();
  } else {
    await page.keyboard.press("Escape");
  }

  await expect(page.getByRole("dialog", { name: "Исключить из членства?" })).toHaveCount(0);
  await expect(page.getByTestId(`remove-membership-${input.positionId}-${userId}`)).toBeVisible();
  trackCreated(entities, "memberships", key);

  return {
    key,
    positionId: input.positionId,
    userId,
    mode: input.mode
  };
}
