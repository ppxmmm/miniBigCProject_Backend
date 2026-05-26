import { type APIRequestContext, expect, test } from "@playwright/test";

type Store = { id: number; code: string; name_en: string };

type DashboardPayload = {
  store: Store;
  hourly_sales: unknown[];
  daily_sales: unknown[];
  low_stock_alerts: unknown[];
};

async function fetchDefaultStoreId(request: APIRequestContext) {
  const response = await request.get("/api/v1/dashboard");
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as DashboardPayload;
  expect(body.store.code).toBe("MBC-0421");
  return body.store.id;
}

test.describe("Dashboard API", () => {
  test("GET /api/v1/dashboard returns seeded default store", async ({ request }) => {
    const response = await request.get("/api/v1/dashboard");
    expect(response.ok()).toBeTruthy();

    const body = (await response.json()) as DashboardPayload;
    expect(body.store.code).toBe("MBC-0421");
    expect(body.store.name_en).toContain("Thonglor");
    expect(Array.isArray(body.hourly_sales)).toBeTruthy();
    expect(body.hourly_sales.length).toBeGreaterThan(0);
    expect(Array.isArray(body.daily_sales)).toBeTruthy();
    expect(Array.isArray(body.low_stock_alerts)).toBeTruthy();
  });

  test("GET /api/v1/stores lists seeded branches", async ({ request }) => {
    const response = await request.get("/api/v1/stores");
    expect(response.ok()).toBeTruthy();

    const stores = (await response.json()) as Store[];
    expect(stores.length).toBeGreaterThanOrEqual(2);
    expect(stores.some((store) => store.code === "MBC-0421")).toBeTruthy();
    expect(stores.some((store) => store.code === "MBC-0178")).toBeTruthy();
  });

  test("store-scoped routes return data for valid store id", async ({ request }) => {
    const storeId = await fetchDefaultStoreId(request);

    const endpoints = [
      `/api/v1/stores/${storeId}`,
      `/api/v1/stores/${storeId}/dashboard`,
      `/api/v1/stores/${storeId}/sales/hourly`,
      `/api/v1/stores/${storeId}/sales/daily`,
      `/api/v1/stores/${storeId}/top-products`,
      `/api/v1/stores/${storeId}/low-stock-alerts`,
      `/api/v1/stores/${storeId}/deliveries`,
      `/api/v1/stores/${storeId}/suggestions`,
    ];

    for (const path of endpoints) {
      const response = await request.get(path);
      expect(response.ok(), `${path} should succeed`).toBeTruthy();
    }
  });

  test("invalid store id returns 400", async ({ request }) => {
    const response = await request.get("/api/v1/stores/not-a-number/top-products");
    expect(response.status()).toBe(400);

    const body = (await response.json()) as { error: string };
    expect(body.error).toContain("positive integer");
  });

  test("unknown store id returns 404", async ({ request }) => {
    const response = await request.get("/api/v1/stores/999999/dashboard");
    expect(response.status()).toBe(404);
  });
});
