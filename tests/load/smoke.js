import http from "k6/http";
import { check, sleep } from "k6";
import { apiBaseUrl } from "./lib.js";

export const options = {
  vus: 5,
  duration: "30s",
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<800"],
  },
};

const baseURL = apiBaseUrl();

export default function smoke() {
  const health = http.get(`${baseURL}/health`);
  check(health, {
    "health status 200": (response) => response.status === 200,
    "health body ok": (response) => response.body === "ok",
  });

  const dashboard = http.get(`${baseURL}/api/v1/dashboard`);
  check(dashboard, {
    "dashboard status 200": (response) => response.status === 200,
    "dashboard has store code": (response) => response.body.includes("MBC-0421"),
  });

  sleep(1);
}
