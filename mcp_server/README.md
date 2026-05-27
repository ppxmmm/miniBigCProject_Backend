# Mini BigC MCP Server

Python MCP server that lets an AI client call the Mini BigC Go backend API.

## Backend Target

By default the server calls:

```bash
http://localhost:5001
```

Override it when needed:

```bash
export MINIBIGC_API_BASE_URL="http://localhost:5001"
```

Protected `/api/v1` routes require a frontend role header. The MCP server sends
`X-User-Role` from `MINIBIGC_API_ROLE`, defaulting to `manager`:

```bash
export MINIBIGC_API_ROLE="manager"
```

## Install

From this folder:

```bash
python -m venv venv
source venv/bin/activate
pip install -e .
```

## Run

```bash
python server.py
```

The server uses MCP stdio transport, so it is normally started by an MCP client.

## MCP Client Config Example

Use the absolute path to this folder in your client config:

```json
{
  "mcpServers": {
    "minibigc-backend": {
      "command": "/Users/imppariya/Documents/Big C/miniBigC_Project/backend/mcp_server/venv/bin/python",
      "args": ["/Users/imppariya/Documents/Big C/miniBigC_Project/backend/mcp_server/server.py"],
      "env": {
        "MINIBIGC_API_BASE_URL": "http://localhost:5001",
        "MINIBIGC_API_ROLE": "manager"
      }
    }
  }
}
```

## Available Tools

- `ai_chat(message, role, use_legacy_path=false)`
- `get_default_dashboard()`
- `get_store_dashboard(store_id)`
- `get_store(store_id)`
- `get_global_data(endpoint)`
- `get_store_data(store_id, endpoint)`
- `list_stores()`
- `list_products()`
- `list_suggestions(store_id=null)`
- `list_low_stock_alerts(store_id=null)`
- `list_deliveries(store_id=null)`

`get_global_data` supports:

```text
dashboard, stores, categories, payment-methods, products, sales/hourly,
sales/daily, sales/monthly, category-sales, payment-mix, top-products,
inventory-items, expiring-inventory, low-stock-alerts, deliveries, suggestions
```

`get_store_data` supports:

```text
dashboard, sales/hourly, sales/daily, sales/monthly, category-sales,
payment-mix, top-products, inventory-items, expiring-inventory,
low-stock-alerts, deliveries, suggestions
```
