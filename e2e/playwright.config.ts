import { defineConfig } from "@playwright/test";
import { env } from "./support/env";

export default defineConfig({
  testDir: "./scenarios",
  fullyParallel: false,
  retries: 0,
  reporter: "list",
  use: {
    baseURL: env.frontendUrl,
    trace: "on-first-retry"
  }
});
