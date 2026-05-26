import { expect, test } from "@playwright/test";

test.describe("AI chat API", () => {
  test("POST /api/v1/ai/chat rejects invalid JSON", async ({ request }) => {
    const response = await request.post("/api/v1/ai/chat", {
      data: "{",
      headers: { "Content-Type": "application/json" },
    });
    expect(response.status()).toBe(400);

    const body = (await response.json()) as { error: string };
    expect(body.error).toContain("valid JSON");
  });

  test("POST /api/v1/ai/chat requires message and role", async ({ request }) => {
    const response = await request.post("/api/v1/ai/chat", {
      data: { message: " ", role: " " },
    });
    expect(response.status()).toBe(400);
  });

  test("POST /api/v1/ai/chat rejects unknown role", async ({ request }) => {
    const response = await request.post("/api/v1/ai/chat", {
      data: { message: "How are sales?", role: "cashier" },
    });
    expect(response.status()).toBe(403);

    const body = (await response.json()) as { error: string };
    expect(body.error).toContain("not allowed");
  });

  test("POST /api/ai/chat accepts manager role when Gemini is configured", async ({ request }) => {
    test.skip(!process.env.GEMINI_API_KEY, "GEMINI_API_KEY not set — skipping live AI call");

    const response = await request.post("/api/ai/chat", {
      data: { message: "Summarize today sales in one sentence.", role: "manager" },
    });
    expect(response.ok()).toBeTruthy();

    const body = (await response.json()) as { reply: string };
    expect(body.reply.trim().length).toBeGreaterThan(0);
  });
});
