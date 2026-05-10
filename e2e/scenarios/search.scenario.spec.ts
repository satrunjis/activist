import { randomUUID } from "node:crypto";
import type { APIRequestContext } from "@playwright/test";
import { assertContactVisibility, assertSearchRows, runUserSearch } from "../actions/search.actions";
import { dbQuery } from "../fixtures/db";
import { expect, test } from "../fixtures/base.fixture";
import { env } from "../support/env";
import { trackCreated, type CreatedEntities } from "../support/teardown";

const ADMIN = {
  login: "admin",
  password: "admin"
} as const;

const LIMITED_OPERATOR = {
  login: "test0",
  password: "test0"
} as const;

type AuthInput = {
  login: string;
  password: string;
};

type RegisterInput = {
  login: string;
  firstName: string;
  groupNumber?: string;
  institute?: string;
};

function uniqueSuffix(): string {
  return randomUUID().replace(/-/g, "").slice(0, 10);
}

function buildUserInput(prefix: string, overrides: Partial<RegisterInput> = {}): RegisterInput {
  const suffix = uniqueSuffix();
  return {
    login: `${prefix}_${suffix}`,
    firstName: `Имя_${suffix}`,
    groupNumber: `GR-${suffix.slice(0, 4)}`,
    institute: `INST-${suffix.slice(0, 5)}`,
    ...overrides
  };
}

async function registerRuntimeUser(
  seed: {
    registerUser: (input: {
      login: string;
      password: string;
      firstName: string;
      gradebookNumber: string;
      groupNumber: string;
      institute: string;
      birthDate: string;
    }) => Promise<{ body: { user: { id: string; login: string } } }>;
  },
  input: RegisterInput
): Promise<{ id: string; login: string }> {
  const result = await seed.registerUser({
    login: input.login,
    password: input.login,
    firstName: input.firstName,
    gradebookNumber: `GB-${uniqueSuffix().slice(0, 6)}`,
    groupNumber: input.groupNumber ?? "GR-0000",
    institute: input.institute ?? "INST-DEFAULT",
    birthDate: "2000-01-01"
  });
  return {
    id: result.body.user.id,
    login: result.body.user.login
  };
}

async function fetchCsrfToken(request: APIRequestContext): Promise<string> {
  const response = await request.get(`${env.apiBaseUrl}/auth/session`);
  if (!response.ok()) {
    throw new Error(`fetchCsrfToken failed: expected 200, got ${response.status()}`);
  }
  const payload = (await response.json()) as { csrf_token?: string };
  const token = payload.csrf_token?.trim();
  if (!token) {
    throw new Error("fetchCsrfToken failed: csrf_token is missing");
  }
  return token;
}

async function updateUserProfileSql(
  userId: string,
  payload: {
    about?: string;
    phone?: string;
    socialLinks?: Array<{ platform: string; value: string }>;
  }
): Promise<void> {
  await dbQuery(
    `
      UPDATE users
      SET
        about = COALESCE($2, about),
        phone = COALESCE($3, phone),
        social_links = COALESCE($4::jsonb, social_links),
        updated_at = now()
      WHERE id = $1
    `,
    [userId, payload.about ?? null, payload.phone ?? null, payload.socialLinks ? JSON.stringify(payload.socialLinks) : null]
  );
}

async function resolveRoleId(roleName: string): Promise<string> {
  const result = await dbQuery<{ id: string }>("SELECT id FROM roles WHERE name = $1 LIMIT 1", [roleName]);
  const id = result.rows[0]?.id?.trim();
  if (!id) {
    throw new Error(`resolveRoleId failed: role ${roleName} not found`);
  }
  return id;
}

async function createDivisionSql(
  entities: CreatedEntities,
  input: { id: string; shortName: string; parentId: string }
): Promise<{ id: string; shortName: string }> {
  await dbQuery(
    `
      INSERT INTO divisions (id, parent_id, short_name, full_name, description, regulation_url, media_links, is_archived, created_at, updated_at)
      VALUES ($1, $2, $3, '', '', '', '[]'::jsonb, false, now(), now())
    `,
    [input.id, input.parentId, input.shortName]
  );
  trackCreated(entities, "divisions", input.id);
  return { id: input.id, shortName: input.shortName };
}

async function createPositionSql(
  entities: CreatedEntities,
  input: { id: string; divisionId: string; title: string; roleName: string }
): Promise<{ id: string; title: string; divisionId: string }> {
  const roleId = await resolveRoleId(input.roleName);
  await dbQuery(
    `
      INSERT INTO positions (id, title, role_id, division_id, max_count, is_archived)
      VALUES ($1, $2, $3, $4, NULL, false)
    `,
    [input.id, input.title, roleId, input.divisionId]
  );
  trackCreated(entities, "positions", input.id);
  return { id: input.id, title: input.title, divisionId: input.divisionId };
}

async function assignMembershipSql(
  entities: CreatedEntities,
  input: { userId: string; positionId: string }
): Promise<void> {
  await dbQuery(
    `
      INSERT INTO memberships (user_id, position_id)
      VALUES ($1, $2)
      ON CONFLICT (user_id, position_id) DO NOTHING
    `,
    [input.userId, input.positionId]
  );
  trackCreated(entities, "memberships", `${input.userId}::${input.positionId}`);
}

async function archivePositionSql(positionId: string): Promise<void> {
  await dbQuery(
    `
      UPDATE positions
      SET is_archived = true
      WHERE id = $1
    `,
    [positionId]
  );
}

test("SRCH-01 user field combination search", async ({ page, seed }) => {
  const institute = `INST-SRCH01-${uniqueSuffix().slice(0, 6)}`;
  const targetInput = buildUserInput("srch01_target", { firstName: "Алексей", institute });
  const decoyInput = buildUserInput("srch01_decoy", { firstName: "Алексей", institute });

  const target = await registerRuntimeUser(seed, targetInput);
  await registerRuntimeUser(seed, decoyInput);
  await updateUserProfileSql(target.id, {
    about: "инициативная группа взаимопомощи"
  });

  await runUserSearch(page, {
    actor: ADMIN,
    first_name: "Алексей",
    institute,
    about: "взаимопомощ"
  });

  await assertSearchRows(page, [target.login]);
  await expect(page.locator('li[data-testid^="search-result-"]')).toHaveCount(1);
});

test("SRCH-02 active role/position search", async ({ page, seed, entities }) => {
  const division = await createDivisionSql(entities, {
    id: `srch02-div-${uniqueSuffix()}`,
    shortName: `srch02-${uniqueSuffix().slice(0, 6)}`,
    parentId: "seed-ops"
  });
  const targetInput = buildUserInput("srch02_target");
  const target = await registerRuntimeUser(seed, targetInput);
  const targetPosition = await createPositionSql(entities, {
    id: `srch02-pos-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Координатор-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });
  await assignMembershipSql(entities, {
    userId: target.id,
    positionId: targetPosition.id
  });

  await runUserSearch(page, {
    actor: ADMIN,
    position_title: targetPosition.title,
    role_name: "Seed Standard"
  });

  await assertSearchRows(page, [target.login]);
});

test("SRCH default excludes archived", async ({ page, seed, entities }) => {
  const institute = `INST-SRCH03-${uniqueSuffix().slice(0, 6)}`;
  const division = await createDivisionSql(entities, {
    id: `srch03-div-${uniqueSuffix()}`,
    shortName: `srch03-${uniqueSuffix().slice(0, 6)}`,
    parentId: "seed-ops"
  });

  const activeInput = buildUserInput("srch03_active", { institute });
  const archivedInput = buildUserInput("srch03_arch", { institute });
  const active = await registerRuntimeUser(seed, activeInput);
  const archived = await registerRuntimeUser(seed, archivedInput);

  const activePosition = await createPositionSql(entities, {
    id: `srch03-pos-active-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Активный-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });
  const archivedPosition = await createPositionSql(entities, {
    id: `srch03-pos-arch-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Архивный-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });

  await assignMembershipSql(entities, {
    userId: active.id,
    positionId: activePosition.id
  });
  await assignMembershipSql(entities, {
    userId: archived.id,
    positionId: archivedPosition.id
  });
  await archivePositionSql(archivedPosition.id);

  await runUserSearch(page, {
    actor: ADMIN,
    institute,
    role_name: "Seed Standard"
  });

  await assertSearchRows(page, [active.login]);
});

test("SRCH include_archived=true includes archived matches", async ({ page, seed, entities }) => {
  const institute = `INST-SRCH04-${uniqueSuffix().slice(0, 6)}`;
  const division = await createDivisionSql(entities, {
    id: `srch04-div-${uniqueSuffix()}`,
    shortName: `srch04-${uniqueSuffix().slice(0, 6)}`,
    parentId: "seed-ops"
  });

  const activeInput = buildUserInput("srch04_active", { institute });
  const archivedInput = buildUserInput("srch04_arch", { institute });
  const active = await registerRuntimeUser(seed, activeInput);
  const archived = await registerRuntimeUser(seed, archivedInput);

  const activePosition = await createPositionSql(entities, {
    id: `srch04-pos-active-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Активный-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });
  const archivedPosition = await createPositionSql(entities, {
    id: `srch04-pos-arch-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Архивный-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });

  await assignMembershipSql(entities, {
    userId: active.id,
    positionId: activePosition.id
  });
  await assignMembershipSql(entities, {
    userId: archived.id,
    positionId: archivedPosition.id
  });
  await archivePositionSql(archivedPosition.id);

  await runUserSearch(page, {
    actor: ADMIN,
    institute,
    role_name: "Seed Standard",
    include_archived: true
  });

  await assertSearchRows(page, [active.login, archived.login]);
});

test("SRCH ViewContacts scope enforcement", async ({ page, seed, entities }) => {
  const division = await createDivisionSql(entities, {
    id: `srch05-div-${uniqueSuffix()}`,
    shortName: `srch05-${uniqueSuffix().slice(0, 6)}`,
    parentId: "seed-ops-field"
  });
  const targetInput = buildUserInput("srch05_target");
  const target = await registerRuntimeUser(seed, targetInput);
  const position = await createPositionSql(entities, {
    id: `srch05-pos-${uniqueSuffix()}`,
    divisionId: division.id,
    title: `Контакты-${uniqueSuffix().slice(0, 5)}`,
    roleName: "Seed Standard"
  });
  await assignMembershipSql(entities, {
    userId: target.id,
    positionId: position.id
  });

  await updateUserProfileSql(target.id, {
    phone: "+70000000099",
    socialLinks: [{ platform: "tg", value: `@${target.login}` }]
  });

  await runUserSearch(page, {
    actor: LIMITED_OPERATOR,
    login: target.login
  });
  await assertSearchRows(page, [target.login]);
  await assertContactVisibility(page, target.login, false);

  await runUserSearch(page, {
    actor: ADMIN,
    login: target.login
  });
  await assertSearchRows(page, [target.login]);
  await assertContactVisibility(page, target.login, true);
});
