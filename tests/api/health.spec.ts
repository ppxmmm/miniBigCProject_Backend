import { expect, test } from "@playwright/test";

test.describe("Health & root", () => {
  test("GET /health returns ok", async ({ request }) => {
    const response = await request.get("/health");
    expect(response.status()).toBe(200);
    expect(await response.text()).toBe("ok");
  });

  test("GET / returns welcome message", async ({ request }) => {
    const response = await request.get("/");
    expect(response.status()).toBe(200);
    expect(await response.text()).toContain("Mini BigC");
  });
});
