"""MCP server for the Mini BigC backend API.

Run with:
    python server.py

Environment:
    MINIBIGC_API_BASE_URL=http://localhost:5001
    MINIBIGC_API_TIMEOUT_SECONDS=30
"""

from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from typing import Any, Literal

from fastmcp import FastMCP


API_BASE_URL = os.getenv("MINIBIGC_API_BASE_URL", "http://localhost:5001").rstrip("/")
API_TIMEOUT_SECONDS = float(os.getenv("MINIBIGC_API_TIMEOUT_SECONDS", "30"))
API_ROLE = os.getenv("MINIBIGC_API_ROLE", "manager").strip()

mcp = FastMCP("Mini BigC Backend API")

GlobalEndpoint = Literal[
    "dashboard",
    "stores",
    "categories",
    "payment-methods",
    "products",
    "sales/hourly",
    "sales/daily",
    "sales/monthly",
    "category-sales",
    "payment-mix",
    "top-products",
    "inventory-items",
    "expiring-inventory",
    "low-stock-alerts",
    "deliveries",
    "suggestions",
]

StoreEndpoint = Literal[
    "dashboard",
    "sales/hourly",
    "sales/daily",
    "sales/monthly",
    "category-sales",
    "payment-mix",
    "top-products",
    "inventory-items",
    "expiring-inventory",
    "low-stock-alerts",
    "deliveries",
    "suggestions",
]


def _url(path: str) -> str:
    return f"{API_BASE_URL}/{path.lstrip('/')}"


def _decode_response(response: urllib.response.addinfourl) -> Any:
    body = response.read().decode("utf-8")
    if not body:
        return None
    try:
        return json.loads(body)
    except json.JSONDecodeError:
        return body


def _api_request(method: str, path: str, payload: dict[str, Any] | None = None) -> Any:
    body = None
    headers = {"Accept": "application/json"}
    if API_ROLE:
        headers["X-User-Role"] = API_ROLE

    if payload is not None:
        body = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    request = urllib.request.Request(_url(path), data=body, headers=headers, method=method)

    try:
        with urllib.request.urlopen(request, timeout=API_TIMEOUT_SECONDS) as response:
            return _decode_response(response)
    except urllib.error.HTTPError as error:
        error_body = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(
            f"Backend API returned HTTP {error.code} for {method} {path}: {error_body}"
        ) from error
    except urllib.error.URLError as error:
        raise RuntimeError(
            f"Could not reach Mini BigC backend at {API_BASE_URL}. "
            "Start the Go backend or set MINIBIGC_API_BASE_URL."
        ) from error


@mcp.tool()
def ai_chat(message: str, role: str, use_legacy_path: bool = False) -> dict[str, Any]:
    """Ask the Mini BigC backend AI chat endpoint a dashboard question.

    The backend requires both message and role. By default this calls
    /api/v1/ai/chat. Set use_legacy_path=true to call /api/ai/chat.
    """

    path = "/api/ai/chat" if use_legacy_path else "/api/v1/ai/chat"
    return _api_request("POST", path, {"message": message, "role": role})


@mcp.tool()
def get_default_dashboard() -> dict[str, Any]:
    """Return the default store dashboard aggregate."""

    return _api_request("GET", "/api/v1/dashboard")


@mcp.tool()
def get_store_dashboard(store_id: int) -> dict[str, Any]:
    """Return the dashboard aggregate for one store."""

    return _api_request("GET", f"/api/v1/stores/{store_id}/dashboard")


@mcp.tool()
def get_store(store_id: int) -> dict[str, Any]:
    """Return one store by ID."""

    return _api_request("GET", f"/api/v1/stores/{store_id}")


@mcp.tool()
def get_global_data(endpoint: GlobalEndpoint) -> Any:
    """Call a global read-only /api/v1 endpoint.

    Allowed values include stores, categories, products, sales/hourly,
    sales/daily, sales/monthly, category-sales, payment-mix, top-products,
    inventory-items, expiring-inventory, low-stock-alerts, deliveries,
    suggestions, and dashboard.
    """

    return _api_request("GET", f"/api/v1/{endpoint}")


@mcp.tool()
def get_store_data(store_id: int, endpoint: StoreEndpoint) -> Any:
    """Call a store-scoped read-only /api/v1/stores/{store_id} endpoint."""

    return _api_request("GET", f"/api/v1/stores/{store_id}/{endpoint}")


@mcp.tool()
def list_stores() -> list[dict[str, Any]]:
    """Return all stores."""

    return _api_request("GET", "/api/v1/stores")


@mcp.tool()
def list_products() -> list[dict[str, Any]]:
    """Return all products."""

    return _api_request("GET", "/api/v1/products")


@mcp.tool()
def list_suggestions(store_id: int | None = None) -> list[dict[str, Any]]:
    """Return suggestions globally, or for one store when store_id is provided."""

    if store_id is None:
        return _api_request("GET", "/api/v1/suggestions")
    return _api_request("GET", f"/api/v1/stores/{store_id}/suggestions")


@mcp.tool()
def list_low_stock_alerts(store_id: int | None = None) -> list[dict[str, Any]]:
    """Return low stock alerts globally, or for one store when store_id is provided."""

    if store_id is None:
        return _api_request("GET", "/api/v1/low-stock-alerts")
    return _api_request("GET", f"/api/v1/stores/{store_id}/low-stock-alerts")


@mcp.tool()
def list_deliveries(store_id: int | None = None) -> list[dict[str, Any]]:
    """Return deliveries globally, or for one store when store_id is provided."""

    if store_id is None:
        return _api_request("GET", "/api/v1/deliveries")
    return _api_request("GET", f"/api/v1/stores/{store_id}/deliveries")


def main() -> None:
    mcp.run()


if __name__ == "__main__":
    main()
