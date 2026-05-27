import { defineConfig } from "@playwright/test";

const baseURL = process.env.API_BASE_URL ?? "http://localhost:5001";

export default defineConfig({
  testDir: "./tests/api",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL,
    extraHTTPHeaders: {
      Accept: "application/json",
    },
  },
});
