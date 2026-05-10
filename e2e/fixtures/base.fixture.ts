import { test as base } from "@playwright/test";
import { env } from "../support/env";
import { createSeedHelpers, type SeedHelpers } from "../support/seed";
import { createCreatedEntities, teardownCreatedEntities, type CreatedEntities } from "../support/teardown";

type E2EFixtures = {
  entities: CreatedEntities;
  seed: SeedHelpers;
};

export const test = base.extend<E2EFixtures>({
  entities: async ({}, use) => {
    const entities = createCreatedEntities();
    try {
      await use(entities);
    } finally {
      await teardownCreatedEntities(entities);
    }
  },
  seed: async ({ playwright, entities }, use) => {
    const apiRequest = await playwright.request.newContext({
      baseURL: env.apiBaseUrl
    });
    try {
      await use(createSeedHelpers(apiRequest, entities));
    } finally {
      await apiRequest.dispose();
    }
  }
});

export { expect } from "@playwright/test";
