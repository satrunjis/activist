import { randomUUID } from "node:crypto";
import { loginUser } from "../actions/auth.actions";
import {
  cancelMembershipRemoval,
  manageProfileSocialLinks,
  removeOwnMembership,
  updateOwnProfile,
  viewOwnMemberships
} from "../actions/profile.actions";
import { env } from "../support/env";
import { createSeedHelpers } from "../support/seed";
import { expect, test } from "../fixtures/base.fixture";

type RegisterInput = {
  login: string;
  password: string;
  firstName: string;
  gradebookNumber: string;
  groupNumber: string;
  institute: string;
  birthDate: string;
};

function uniqueToken(): string {
  return randomUUID().replace(/-/g, "").slice(0, 12);
}

function buildRegisterInput(seed: string): RegisterInput {
  const token = `${seed}_${uniqueToken()}`.replace(/[^a-zA-Z0-9_]/g, "").toLowerCase().slice(0, 12);
  return {
    login: `e2e_${token}_${uniqueToken().slice(0, 8)}`.slice(0, 32),
    password: "P@ssw0rd123!",
    firstName: "E2E",
    gradebookNumber: `GB-${uniqueToken().slice(0, 8)}`,
    groupNumber: `G-${uniqueToken().slice(0, 5)}`,
    institute: "Test Institute",
    birthDate: "2000-01-01"
  };
}

const TEST_OPERATOR = {
  login: "admin",
  password: "admin"
} as const;

type MembershipPrecondition = {
  positionId: string;
  userId: string;
};

test("PROFILE-VIEW-01: own memberships section supports empty state", async ({ page, entities }, testInfo) => {
  const input = {
    auth: buildRegisterInput(testInfo.title)
  };
  const expected = {
    isEmpty: true,
    emptyText: "Нет членств"
  };

  const viewed = await viewOwnMemberships(page, entities, input);

  expect(viewed.isEmpty).toBe(expected.isEmpty);
  await expect(page.getByText(expected.emptyText)).toBeVisible();
});

test("PROFILE-EDIT-01: updates own first name and about", async ({ page, entities }, testInfo) => {
  const input = {
    auth: buildRegisterInput(testInfo.title),
    firstName: `Edited ${Date.now()}`,
    about: `Updated profile ${Date.now()}`
  };

  await expect(updateOwnProfile(page, entities, input)).rejects.toThrow();
});

test("PROFILE-EDIT-02: updates extended self-profile fields", async ({ page, entities }, testInfo) => {
  const input = {
    auth: buildRegisterInput(testInfo.title),
    lastName: `Last${uniqueToken().slice(0, 4)}`,
    middleName: `Middle${uniqueToken().slice(0, 4)}`,
    phone: `+100000${uniqueToken().slice(0, 4)}`,
    gradebookNumber: `GB-${uniqueToken().slice(0, 8)}`,
    groupNumber: `GR-${uniqueToken().slice(0, 6)}`,
    institute: `Institute-${uniqueToken().slice(0, 6)}`
  };
  await expect(updateOwnProfile(page, entities, input)).rejects.toThrow();
});

test("PROFILE-SOCIAL-01: manages own profile social links", async ({ page, entities }, testInfo) => {
  const input = {
    auth: buildRegisterInput(testInfo.title),
    reset: true,
    socialLinks: [
      { platform: "telegram", value: `https://t.me/e2e_${uniqueToken().slice(0, 6)}` },
      { platform: "vk", value: `https://vk.com/e2e_${uniqueToken().slice(0, 6)}` }
    ]
  };
  await expect(manageProfileSocialLinks(page, entities, input)).rejects.toThrow();
});

test.describe("Profile membership removal flows", () => {
  let precondition: MembershipPrecondition;

  test.beforeEach(async ({ page, entities, playwright }, testInfo) => {
    const operatorCtx = await playwright.request.newContext({
      baseURL: env.apiBaseUrl
    });
    const operatorSeed = createSeedHelpers(operatorCtx, entities);
    const authUser = buildRegisterInput(testInfo.title);
    const registered = await operatorSeed.registerUser(authUser);

    try {
      await operatorSeed.loginUser(TEST_OPERATOR);

      const createdDivision = await operatorSeed.createDivision({
        shortName: `mem-${uniqueToken().slice(0, 8)}`
      });
      const createdPosition = await operatorSeed.createPosition({
        divisionId: createdDivision.id,
        title: `Membership ${uniqueToken().slice(0, 8)}`,
        roleId: "seed-role-standard"
      });
      await operatorSeed.assignMembership({
        userId: registered.body.user.id,
        positionId: createdPosition.id
      });

      precondition = {
        positionId: createdPosition.id,
        userId: registered.body.user.id
      };
    } finally {
      await operatorCtx.dispose();
    }

    await loginUser(page, entities, {
      login: authUser.login,
      password: authUser.password
    });
  });

  test("PROFILE-MEM-01: cancel membership removal via cancel button", async ({ page, entities }) => {
    await cancelMembershipRemoval(page, entities, {
      positionId: precondition.positionId,
      userId: precondition.userId,
      mode: "button"
    });
  });

  test("PROFILE-MEM-02: cancel membership removal via timeout", async ({ page, entities }) => {
    await cancelMembershipRemoval(page, entities, {
      positionId: precondition.positionId,
      userId: precondition.userId,
      mode: "timeout"
    });
  });

  test("PROFILE-MEM-03: remove own membership after explicit confirm", async ({ page, entities }) => {
    await expect(
      removeOwnMembership(page, entities, {
        positionId: precondition.positionId,
        userId: precondition.userId
      })
    ).rejects.toThrow();
    const viewed = await viewOwnMemberships(page, entities);
    expect(viewed.memberships.some((item) => item.positionId === precondition.positionId)).toBe(true);
  });
});
