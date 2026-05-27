import http from "k6/http";
import { check, sleep } from "k6";
import { apiBaseUrl } from "./lib.js";

export const options = {
  stages: [
    { duration: "20s", target: 10 },
    { duration: "40s", target: 25 },
    { duration: "20s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.02"],
    http_req_duration: ["p(95)<1200"],
  },
};

const baseURL = apiBaseUrl();

export function setup() {
  const response = http.get(`${baseURL}/api/v1/dashboard`);
  if (response.status !== 200) {
    throw new Error(`setup failed: dashboard returned ${response.status}`);
  }

  const body = JSON.parse(response.body);
  return { storeId: body.store.id };
}

export default function dashboardLoad(data) {
  const storeId = data.storeId;

  const responses = http.batch([
    ["GET", `${baseURL}/api/v1/dashboard`, null, { tags: { name: "dashboard_default" } }],
    ["GET", `${baseURL}/api/v1/stores/${storeId}/sales/hourly`, null, { tags: { name: "hourly_sales" } }],
    ["GET", `${baseURL}/api/v1/stores/${storeId}/top-products`, null, { tags: { name: "top_products" } }],
    ["GET", `${baseURL}/api/v1/stores/${storeId}/low-stock-alerts`, null, { tags: { name: "low_stock" } }],
    ["GET", `${baseURL}/api/v1/stores/${storeId}/deliveries`, null, { tags: { name: "deliveries" } }],
  ]);

  for (const response of responses) {
    check(response, {
      "status is 200": (item) => item.status === 200,
    });
  }

  sleep(0.5);
}
