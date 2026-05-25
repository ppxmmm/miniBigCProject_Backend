# Dashboard Database Schema

This document describes the PostgreSQL schema used by the backend dashboard API consumed by the frontend dashboard page.

## Detected Data Stack

- Database: PostgreSQL.
- ORM: GORM.
- Migration tool: GORM `AutoMigrate` in `internal/db/bootstrap.go`, plus idempotent raw SQL index creation in `ensureIndexes`.
- Seed strategy: `internal/util/seed.go` loads dashboard seed data during bootstrap when needed.
- Primary dashboard read path: `GET /api/v1/dashboard` and `GET /api/v1/stores/{storeID}/dashboard`.

## Dashboard API Read Model

The frontend dashboard consumes `DashboardData` from `internal/model/dashboard.go`:

- `store`
- `hourly_sales`
- `daily_sales`
- `monthly_sales`
- `category_sales`
- `payment_mix`
- `top_products`
- `inventory_items`
- `expiring_inventory`
- `low_stock_alerts`
- `deliveries`
- `suggestions`

The repository builds this response in `internal/repo/dashboard_repo.go` by loading each table for a single `store_id`.

## Tables

### `stores`

Branch metadata used in the dashboard shell, page header, profile, and store selector.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `code` | varchar(32) | Not null, unique |
| `name_th` | varchar(255) | Not null |
| `name_en` | varchar(255) | Not null |
| `short_name_th` | varchar(120) | Not null |
| `short_name_en` | varchar(120) | Not null |
| `address_th` | text | Not null |
| `address_en` | text | Not null |
| `manager_name_th` | varchar(160) | Not null |
| `manager_name_en` | varchar(160) | Not null |
| `manager_initials` | varchar(12) | Not null |
| `staff_name_th` | varchar(160) | Not null |
| `staff_name_en` | varchar(160) | Not null |
| `staff_initials` | varchar(12) | Not null |

Indexes:

- `idx_stores_code` unique on `code`.

### `categories`

Product and reporting category metadata joined into category sales and inventory alert views.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `name_th` | varchar(160) | Not null |
| `name_en` | varchar(160) | Not null, unique |
| `color` | varchar(80) | Not null, default empty string |

Indexes:

- `idx_categories_name_en` unique on `name_en`.

### `payment_methods`

Payment method metadata joined into payment mix results.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `name_th` | varchar(160) | Not null |
| `name_en` | varchar(160) | Not null, unique |

Indexes:

- `idx_payment_methods_name_en` unique on `name_en`.

### `products`

Product master data shared by top products, inventory, expiry, and stock alert queries.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `sku` | varchar(40) | Primary key |
| `name_th` | varchar(255) | Not null |
| `name_en` | varchar(255) | Not null |
| `category_id` | bigint | Category reference by ID |

Query usage:

- Joined by `products.sku` from `top_products`, `inventory_items`, `expiring_inventory`, and `low_stock_alerts`.

### `sales_hourly`

Hourly current and comparison sales for the dashboard revenue chart and KPI cards.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sale_date` | date | Not null |
| `hour` | integer | Not null, check `hour >= 0 AND hour <= 23` |
| `sales_value` | numeric(14,2) | Not null |
| `comparison_sales_value` | numeric(14,2) | Not null |

Indexes:

- `idx_sales_hourly_store_date_hour` unique on `(store_id, sale_date, hour)`.

Read pattern:

```sql
SELECT *
FROM sales_hourly
WHERE store_id = ?
ORDER BY sale_date, hour;
```

### `sales_daily`

Daily current and comparison sales for weekly dashboard chart mode.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sale_date` | date | Not null |
| `sales_value` | numeric(14,2) | Not null |
| `comparison_sales_value` | numeric(14,2) | Not null |

Indexes:

- `idx_sales_daily_store_date` unique on `(store_id, sale_date)`.

Read pattern:

```sql
SELECT *
FROM sales_daily
WHERE store_id = ?
ORDER BY sale_date;
```

### `sales_monthly`

Monthly sales values for yearly dashboard and revenue views.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `year` | integer | Not null |
| `month` | integer | Not null, check `month >= 1 AND month <= 12` |
| `sales_value` | numeric(14,2) | Not null |

Indexes:

- `idx_sales_monthly_store_year_month` unique on `(store_id, year, month)`.

Read pattern:

```sql
SELECT *
FROM sales_monthly
WHERE store_id = ?
ORDER BY year, month;
```

### `category_sales`

Category-level revenue summary used by the dashboard donut and category breakdown.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `category_id` | bigint | Not null, joined to `categories.id` |
| `sales_date` | date | Not null |
| `sales_value` | numeric(14,2) | Not null |
| `share` | numeric(7,4) | Not null |
| `trend_percent` | numeric(7,2) | Not null |

Indexes:

- `idx_category_sales_store_category_date` unique on `(store_id, category_id, sales_date)`.

Read pattern:

```sql
SELECT
  cs.id,
  cs.store_id,
  cs.category_id,
  c.name_th,
  c.name_en,
  c.color,
  cs.sales_date,
  cs.sales_value,
  cs.share,
  cs.trend_percent
FROM category_sales cs
JOIN categories c ON c.id = cs.category_id
WHERE cs.store_id = ?
ORDER BY cs.sales_value DESC;
```

### `payment_mix`

Payment method share data used by the revenue page and raw data API.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sales_date` | date | Not null |
| `payment_method_id` | bigint | Not null, joined to `payment_methods.id` |
| `share` | numeric(7,4) | Not null |

Indexes:

- `idx_payment_mix_store_method_date` unique on `(store_id, payment_method_id, sales_date)`.

Read pattern:

```sql
SELECT
  pm.id,
  pm.store_id,
  pm.sales_date,
  pm.payment_method_id,
  p.name_th,
  p.name_en,
  pm.share
FROM payment_mix pm
JOIN payment_methods p ON p.id = pm.payment_method_id
WHERE pm.store_id = ?
ORDER BY pm.share DESC;
```

### `top_products`

Top-selling product snapshots used by the dashboard and revenue page.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sku` | varchar(40) | Not null, joined to `products.sku` |
| `sales_date` | date | Not null |
| `sold_quantity` | integer | Not null |
| `sales_value` | numeric(14,2) | Not null |
| `trend_percent` | numeric(7,2) | Not null |

Indexes:

- `idx_top_products_store_date` on `(store_id, sales_date)`.
- `idx_top_products_store_sku_date` unique on `(store_id, sku, sales_date)`.

Read pattern:

```sql
SELECT
  tp.id,
  tp.store_id,
  tp.sku,
  p.name_th,
  p.name_en,
  tp.sales_date,
  tp.sold_quantity,
  tp.sales_value,
  tp.trend_percent
FROM top_products tp
JOIN products p ON p.sku = tp.sku
WHERE tp.store_id = ?
ORDER BY tp.sales_value DESC;
```

### `inventory_items`

Current stock state used by raw data APIs and as source context for alerts.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sku` | varchar(40) | Not null, joined to `products.sku` |
| `stock_quantity` | integer | Not null |
| `reorder_quantity` | integer | Not null |
| `location_code` | varchar(40) | Not null |
| `price` | numeric(12,2) | Not null |

Indexes:

- `idx_inventory_store_sku_location` unique on `(store_id, sku, location_code)`.

Read pattern:

```sql
SELECT
  ii.id,
  ii.store_id,
  ii.sku,
  p.name_th,
  p.name_en,
  ii.stock_quantity,
  ii.reorder_quantity,
  ii.location_code,
  ii.price
FROM inventory_items ii
JOIN products p ON p.sku = ii.sku
WHERE ii.store_id = ?
ORDER BY ii.location_code, ii.sku;
```

### `expiring_inventory`

Products approaching expiry used by the dashboard urgent alert card and alerts page.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sku` | varchar(40) | Not null, joined to `products.sku` |
| `category_id` | bigint | Not null, joined to `categories.id` |
| `expiry_date` | date | Not null |
| `stock_quantity` | integer | Not null |
| `location_code` | varchar(40) | Not null |
| `price` | numeric(12,2) | Not null |

Indexes:

- `idx_expiring_inventory_store_expiry` on `(store_id, expiry_date)`.
- `idx_expiring_store_sku_expiry_location` unique on `(store_id, sku, expiry_date, location_code)`.

Read pattern:

```sql
SELECT
  ei.id,
  ei.store_id,
  ei.sku,
  p.name_th,
  p.name_en,
  ei.category_id,
  c.name_th AS category_th,
  c.name_en AS category_en,
  ei.expiry_date,
  ei.stock_quantity,
  ei.location_code,
  ei.price
FROM expiring_inventory ei
JOIN products p ON p.sku = ei.sku
JOIN categories c ON c.id = ei.category_id
WHERE ei.store_id = ?
ORDER BY ei.expiry_date, ei.stock_quantity DESC;
```

### `low_stock_alerts`

Products below reorder threshold used by the dashboard urgent alert card and alerts page.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | bigint | Primary key |
| `store_id` | bigint | Not null |
| `sku` | varchar(40) | Not null, joined to `products.sku` |
| `category_id` | bigint | Not null, joined to `categories.id` |
| `stock_quantity` | integer | Not null |
| `reorder_quantity` | integer | Not null |
| `location_code` | varchar(40) | Not null |

Indexes:

- `idx_low_stock_alerts_store_stock` on `(store_id, stock_quantity)`.
- `idx_low_stock_store_sku_location` unique on `(store_id, sku, location_code)`.

Read pattern:

```sql
SELECT
  lsa.id,
  lsa.store_id,
  lsa.sku,
  p.name_th,
  p.name_en,
  lsa.category_id,
  c.name_th AS category_th,
  c.name_en AS category_en,
  lsa.stock_quantity,
  lsa.reorder_quantity,
  lsa.location_code
FROM low_stock_alerts lsa
JOIN products p ON p.sku = lsa.sku
JOIN categories c ON c.id = lsa.category_id
WHERE lsa.store_id = ?
ORDER BY lsa.stock_quantity, lsa.sku;
```

### `deliveries`

Delivery orders and fulfillment status used by the dashboard delivery table and delivery page.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | varchar(40) | Primary key |
| `store_id` | bigint | Not null |
| `customer_name_th` | varchar(160) | Not null |
| `customer_name_en` | varchar(160) | Not null |
| `address_th` | text | Not null |
| `address_en` | text | Not null |
| `item_count` | integer | Not null |
| `order_value` | numeric(12,2) | Not null |
| `driver_name_th` | varchar(160) | Not null |
| `driver_name_en` | varchar(160) | Not null |
| `status` | varchar(40) | Not null, expected values: `preparing`, `enRoute`, `delivered` |
| `eta_time` | time | Not null |
| `is_late` | boolean | Not null |
| `distance_km` | numeric(8,2) | Not null |

Indexes:

- `idx_deliveries_store_status` on `(store_id, status)`.
- `idx_deliveries_store_eta` on `(store_id, eta_time)`.

Read pattern:

```sql
SELECT *
FROM deliveries
WHERE store_id = ?
ORDER BY eta_time, id;
```

### `suggestions`

Promotion and event suggestions used by the suggestions page and dashboard insight surfaces.

| Column | Type | Constraints / Notes |
| --- | --- | --- |
| `id` | varchar(40) | Primary key |
| `store_id` | bigint | Not null |
| `kind` | varchar(40) | Not null, expected values: `promo`, `event` |
| `icon` | varchar(40) | Not null |
| `title_th` | varchar(255) | Not null |
| `title_en` | varchar(255) | Not null |
| `description_th` | text | Not null |
| `description_en` | text | Not null |
| `upside_value` | numeric(14,2) | Not null |
| `confidence` | numeric(7,4) | Not null |
| `duration_th` | varchar(120) | Not null |
| `duration_en` | varchar(120) | Not null |
| `target_th` | varchar(160) | Not null |
| `target_en` | varchar(160) | Not null |
| `type` | varchar(40) | Not null |

Indexes:

- `idx_suggestions_store_kind_type` on `(store_id, kind, type)`.

Read pattern:

```sql
SELECT *
FROM suggestions
WHERE store_id = ?
ORDER BY kind DESC, upside_value DESC;
```

## Relationships

The current GORM models rely mainly on key columns and joins rather than explicit foreign-key tags. Logical relationships are:

- `sales_hourly.store_id`, `sales_daily.store_id`, `sales_monthly.store_id`, `category_sales.store_id`, `payment_mix.store_id`, `top_products.store_id`, `inventory_items.store_id`, `expiring_inventory.store_id`, `low_stock_alerts.store_id`, `deliveries.store_id`, and `suggestions.store_id` reference `stores.id`.
- `products.category_id`, `category_sales.category_id`, `expiring_inventory.category_id`, and `low_stock_alerts.category_id` reference `categories.id`.
- `payment_mix.payment_method_id` references `payment_methods.id`.
- `top_products.sku`, `inventory_items.sku`, `expiring_inventory.sku`, and `low_stock_alerts.sku` reference `products.sku`.

## Data Model Decisions

- Localized Thai and English text is stored in separate columns to match existing API response fields and avoid JSON localization parsing in queries.
- Dashboard reporting tables are store-scoped through `store_id` so the same API can support multiple branches.
- Product display names are normalized into `products`; dashboard queries join product metadata where needed.
- Category display names and colors are normalized into `categories`; category sales and inventory alert queries join category metadata.
- Promotions and events share the `suggestions` table with a `kind` discriminator because they have the same frontend shape.
- Unique indexes preserve idempotent seed behavior and prevent duplicate facts for the same store/date/product/location grain.

## Query Impact

- The dashboard endpoint currently performs one store lookup plus eleven table loads for the selected store. It avoids N+1 loops by querying each collection in one statement.
- Existing joins are bounded by `store_id` filters and indexed lookup keys.
- Sorting matches dashboard display needs: revenue desc for category/top-product summaries, expiry ascending for expiring inventory, stock ascending for low-stock alerts, and ETA ascending for deliveries.
- The schema is read-optimized for a compact operational dashboard. If dashboard traffic grows, the next optimization would be a materialized dashboard snapshot per store/date rather than increasing frontend-specific query logic.

## Verification

Use these commands from `backend/`:

```bash
go test ./...
```

Start the backend and confirm the dashboard aggregate response:

```bash
go run ./cmd/server
curl http://localhost:5001/api/v1/dashboard
```

Confirm raw table endpoints when checking individual tables:

```bash
curl http://localhost:5001/api/v1/stores
curl http://localhost:5001/api/v1/sales/hourly
curl http://localhost:5001/api/v1/category-sales
curl http://localhost:5001/api/v1/deliveries
```
