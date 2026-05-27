# Role-Restricted Paths And Related APIs

This document maps frontend role-restricted paths to the backend API routes they depend on or are related to.

## Frontend Role Restrictions

| Frontend path | Allowed roles | Restriction behavior |
| --- | --- | --- |
| `/dashboard` | `manager`, `staff` | Staff can access the page, but some KPI values are masked. |
| `/revenue` | `manager` | Staff sees a restricted-access message. |
| `/alerts` | `manager`, `staff` | Staff can access the page, but manager-only action controls are hidden. |
| `/deliveries` | `manager`, `staff` | Staff can access the page, but order values and manager actions are hidden. |
| `/suggestions` | `manager` | Staff sees a restricted-access message. |
| `/profile` | `manager`, `staff` | Both roles can access the page. |

All app routes under `src/app/(app)` require authentication through `RequireAuth`.

## APIs Used By Restricted Paths

The current frontend does not call most resource-specific backend endpoints directly. It loads branch data through:

```http
GET /api/v1/dashboard
```

That request is made by `src/lib/api/dashboard.ts`, loaded by `BranchDataProvider`, and mapped into page data by `src/lib/api/mappers.ts`.

## `/revenue`

The `/revenue` page is manager-only.

Current direct frontend API:

```http
GET /api/v1/dashboard
```

Relevant fields from the dashboard response:

- `hourly_sales`
- `daily_sales`
- `monthly_sales`
- `category_sales`
- `payment_mix`
- `top_products`

Related backend routes from the available API list:

```http
GET /api/v1/dashboard
GET /api/v1/stores/{storeID}/dashboard
GET /api/v1/sales/hourly
GET /api/v1/sales/daily
GET /api/v1/sales/monthly
GET /api/v1/category-sales
GET /api/v1/payment-mix
GET /api/v1/top-products
GET /api/v1/stores/{storeID}/sales/hourly
GET /api/v1/stores/{storeID}/sales/daily
GET /api/v1/stores/{storeID}/sales/monthly
GET /api/v1/stores/{storeID}/category-sales
GET /api/v1/stores/{storeID}/payment-mix
GET /api/v1/stores/{storeID}/top-products
```

## `/suggestions`

The `/suggestions` page is manager-only.

Current direct frontend API:

```http
GET /api/v1/dashboard
```

Relevant fields from the dashboard response:

- `suggestions`
- `daily_sales`
- `category_sales`
- `low_stock_alerts`
- `expiring_inventory`
- `deliveries`

Related backend routes from the available API list:

```http
GET /api/v1/dashboard
GET /api/v1/stores/{storeID}/dashboard
GET /api/v1/suggestions
GET /api/v1/stores/{storeID}/suggestions
GET /api/v1/sales/daily
GET /api/v1/category-sales
GET /api/v1/low-stock-alerts
GET /api/v1/expiring-inventory
GET /api/v1/deliveries
GET /api/v1/stores/{storeID}/sales/daily
GET /api/v1/stores/{storeID}/category-sales
GET /api/v1/stores/{storeID}/low-stock-alerts
GET /api/v1/stores/{storeID}/expiring-inventory
GET /api/v1/stores/{storeID}/deliveries
```

## Shared Pages With Partial Manager-Only Data

These paths are not fully restricted, but they hide or mask some manager-only data for staff.

### `/dashboard`

Current direct frontend API:

```http
GET /api/v1/dashboard
```

Manager-sensitive data shown or masked on this page includes revenue KPIs, Top-30 product performance, and performance highlights.

### `/deliveries`

Current direct frontend API:

```http
GET /api/v1/dashboard
```

Related backend routes:

```http
GET /api/v1/deliveries
GET /api/v1/stores/{storeID}/deliveries
```

Staff can see delivery status data, but order values and manager actions are hidden.

### `/alerts`

Current direct frontend API:

```http
GET /api/v1/dashboard
```

Related backend routes:

```http
GET /api/v1/inventory-items
GET /api/v1/expiring-inventory
GET /api/v1/low-stock-alerts
GET /api/v1/stores/{storeID}/inventory-items
GET /api/v1/stores/{storeID}/expiring-inventory
GET /api/v1/stores/{storeID}/low-stock-alerts
```

Staff can see alert data, but manager-only action controls are hidden.

## AI Chat

AI chat is available from shared UI and is not specific to the manager-only routes.

Backend route from the provided `/api/v1` list:

```http
POST /api/v1/ai/chat
```

The frontend also has another chat helper that calls:

```http
POST /api/ai/chat
```

That route is proxied separately by `next.config.ts` and is not part of the provided `/api/v1` route list.

