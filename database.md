The frontend currently stores dashboard data as TypeScript mock constants. This document translates that mock data into a relational database shape suitable for the backend.

## Detected Data Stack

- Frontend data source: static TypeScript mock data.
- Backend database: not confirmed from the inspected frontend file.
- ORM or query builder: not confirmed from the inspected frontend file.
- Migration tool: not confirmed from the inspected frontend file.

## Tables

### `stores`

Stores Mini BigC branch metadata.

| Field              | Type          | Notes                                 |
| ------------------ | ------------- | ------------------------------------- |
| `id`               | bigint / uuid | Primary key                           |
| `code`             | varchar       | Unique store code, example `MBC-0421` |
| `name_th`          | varchar       | Thai store name                       |
| `name_en`          | varchar       | English store name                    |
| `short_name_th`    | varchar       | Thai short branch name                |
| `short_name_en`    | varchar       | English short branch name             |
| `address_th`       | text          | Thai address                          |
| `address_en`       | text          | English address                       |
| `manager_name_th`  | varchar       | Thai manager name                     |
| `manager_name_en`  | varchar       | English manager name                  |
| `manager_initials` | varchar       | Manager initials                      |
| `staff_name_th`    | varchar       | Thai staff name                       |
| `staff_name_en`    | varchar       | English staff name                    |
| `staff_initials`   | varchar       | Staff initials                        |

### `sales_hourly`

Stores current and comparison hourly sales.

| Field                    | Type          | Notes                                     |
| ------------------------ | ------------- | ----------------------------------------- |
| `id`                     | bigint / uuid | Primary key                               |
| `store_id`               | bigint / uuid | Foreign key to `stores.id`                |
| `sale_date`              | date          | Business date                             |
| `hour`                   | integer       | Hour of day, example `6` through `23`     |
| `sales_value`            | decimal       | Current sales value                       |
| `comparison_sales_value` | decimal       | Previous-period value, from `HOURLY_YEST` |

### `sales_daily`

Stores daily sales and previous-period comparison values.

| Field                    | Type          | Notes                                    |
| ------------------------ | ------------- | ---------------------------------------- |
| `id`                     | bigint / uuid | Primary key                              |
| `store_id`               | bigint / uuid | Foreign key to `stores.id`               |
| `sale_date`              | date          | Business date                            |
| `sales_value`            | decimal       | Current sales value                      |
| `comparison_sales_value` | decimal       | Previous-period value, from `DAILY_LAST` |

### `sales_monthly`

Stores monthly sales values.

| Field         | Type          | Notes                            |
| ------------- | ------------- | -------------------------------- |
| `id`          | bigint / uuid | Primary key                      |
| `store_id`    | bigint / uuid | Foreign key to `stores.id`       |
| `year`        | integer       | Calendar year                    |
| `month`       | integer       | Calendar month, `1` through `12` |
| `sales_value` | decimal       | Monthly sales value              |

### `categories`

Stores product or reporting categories.

| Field     | Type          | Notes                                |
| --------- | ------------- | ------------------------------------ |
| `id`      | bigint / uuid | Primary key                          |
| `name_th` | varchar       | Thai category name                   |
| `name_en` | varchar       | English category name                |
| `color`   | varchar       | UI color token, example `oklch(...)` |

### `category_sales`

Stores category-level sales summary data.

| Field           | Type          | Notes                                |
| --------------- | ------------- | ------------------------------------ |
| `id`            | bigint / uuid | Primary key                          |
| `store_id`      | bigint / uuid | Foreign key to `stores.id`           |
| `category_id`   | bigint / uuid | Foreign key to `categories.id`       |
| `sales_date`    | date          | Business date or reporting date      |
| `sales_value`   | decimal       | Sales value, from `v`                |
| `share`         | decimal       | Category sales share, example `0.38` |
| `trend_percent` | decimal       | Trend percentage                     |

### `payment_methods`

Stores supported payment method labels.

| Field     | Type          | Notes                       |
| --------- | ------------- | --------------------------- |
| `id`      | bigint / uuid | Primary key                 |
| `name_th` | varchar       | Thai payment method name    |
| `name_en` | varchar       | English payment method name |

### `payment_mix`

Stores payment method share data.

| Field               | Type          | Notes                               |
| ------------------- | ------------- | ----------------------------------- |
| `id`                | bigint / uuid | Primary key                         |
| `store_id`          | bigint / uuid | Foreign key to `stores.id`          |
| `sales_date`        | date          | Business date or reporting date     |
| `payment_method_id` | bigint / uuid | Foreign key to `payment_methods.id` |
| `share`             | decimal       | Payment share, example `0.38`       |

### `products`

Stores product master data shared by sales, expiry, and stock views.

| Field         | Type          | Notes                                                                  |
| ------------- | ------------- | ---------------------------------------------------------------------- |
| `sku`         | varchar       | Primary key, example `FB-0102`                                         |
| `name_th`     | varchar       | Thai product name                                                      |
| `name_en`     | varchar       | English product name                                                   |
| `category_id` | bigint / uuid | Foreign key to `categories.id`, nullable if source category is unknown |

### `top_products`

Stores top-selling product snapshots.

| Field           | Type          | Notes                           |
| --------------- | ------------- | ------------------------------- |
| `id`            | bigint / uuid | Primary key                     |
| `store_id`      | bigint / uuid | Foreign key to `stores.id`      |
| `sku`           | varchar       | Foreign key to `products.sku`   |
| `sales_date`    | date          | Business date or reporting date |
| `sold_quantity` | integer       | Units sold                      |
| `sales_value`   | decimal       | Sales value                     |
| `trend_percent` | decimal       | Trend percentage                |

### `inventory_items`

Stores current inventory state for products in a store.

| Field              | Type          | Notes                                       |
| ------------------ | ------------- | ------------------------------------------- |
| `id`               | bigint / uuid | Primary key                                 |
| `store_id`         | bigint / uuid | Foreign key to `stores.id`                  |
| `sku`              | varchar       | Foreign key to `products.sku`               |
| `stock_quantity`   | integer       | Current stock count                         |
| `reorder_quantity` | integer       | Reorder threshold or target quantity        |
| `location_code`    | varchar       | Shelf or storage location, example `F-12-B` |
| `price`            | decimal       | Current unit price                          |

### `expiring_inventory`

Stores products approaching expiry.

| Field            | Type          | Notes                          |
| ---------------- | ------------- | ------------------------------ |
| `id`             | bigint / uuid | Primary key                    |
| `store_id`       | bigint / uuid | Foreign key to `stores.id`     |
| `sku`            | varchar       | Foreign key to `products.sku`  |
| `category_id`    | bigint / uuid | Foreign key to `categories.id` |
| `expiry_date`    | date          | Expiry date                    |
| `stock_quantity` | integer       | Expiring stock count           |
| `location_code`  | varchar       | Shelf or storage location      |
| `price`          | decimal       | Unit price                     |

### `low_stock_alerts`

Stores products below reorder threshold.

| Field              | Type          | Notes                                |
| ------------------ | ------------- | ------------------------------------ |
| `id`               | bigint / uuid | Primary key                          |
| `store_id`         | bigint / uuid | Foreign key to `stores.id`           |
| `sku`              | varchar       | Foreign key to `products.sku`        |
| `category_id`      | bigint / uuid | Foreign key to `categories.id`       |
| `stock_quantity`   | integer       | Current stock count                  |
| `reorder_quantity` | integer       | Reorder threshold or target quantity |
| `location_code`    | varchar       | Shelf or storage location            |

### `deliveries`

Stores delivery orders and current fulfillment status.

| Field              | Type           | Notes                              |
| ------------------ | -------------- | ---------------------------------- |
| `id`               | varchar        | Primary key, example `BC-26052201` |
| `store_id`         | bigint / uuid  | Foreign key to `stores.id`         |
| `customer_name_th` | varchar        | Thai customer display name         |
| `customer_name_en` | varchar        | English customer display name      |
| `address_th`       | text           | Thai delivery address              |
| `address_en`       | text           | English delivery address           |
| `item_count`       | integer        | Number of order items              |
| `order_value`      | decimal        | Delivery order value               |
| `driver_name_th`   | varchar        | Thai driver name                   |
| `driver_name_en`   | varchar        | English driver name                |
| `status`           | varchar / enum | Delivery status                    |
| `eta_time`         | time           | Estimated arrival time             |
| `is_late`          | boolean        | Whether delivery is late           |
| `distance_km`      | decimal        | Delivery distance in kilometers    |

### `suggestions`

Stores promotion and event suggestions. `PROMOS` and `EVENTS` share the same TypeScript shape, so they can be stored in one table with a `kind` discriminator.

| Field            | Type           | Notes                             |
| ---------------- | -------------- | --------------------------------- |
| `id`             | varchar        | Primary key, example `p1` or `e1` |
| `store_id`       | bigint / uuid  | Foreign key to `stores.id`        |
| `kind`           | varchar / enum | `promo` or `event`                |
| `icon`           | varchar        | UI icon key                       |
| `title_th`       | varchar        | Thai title                        |
| `title_en`       | varchar        | English title                     |
| `description_th` | text           | Thai description                  |
| `description_en` | text           | English description               |
| `upside_value`   | decimal        | Estimated upside value            |
| `confidence`     | decimal        | Confidence score, example `0.92`  |
| `duration_th`    | varchar        | Thai duration label               |
| `duration_en`    | varchar        | English duration label            |
| `target_th`      | varchar        | Thai target segment               |
| `target_en`      | varchar        | English target segment            |
| `type`           | varchar / enum | Suggestion type                   |

## Enum-Like Values

### Delivery Status

From `DeliveryStatus`:

- `preparing`
- `enRoute`
- `delivered`

### Suggestion Kind

Derived from the mock constant source:

- `promo`
- `event`

### Suggestion Type

From `PROMOS` and `EVENTS`:

- `markdown`
- `bundle`
- `discount`
- `event`

## Data Model Decisions

- Keep `products` as a product master table because product fields repeat across top products, expiring inventory, and low stock alerts.
- Keep `categories` separate because category labels appear in revenue summaries and inventory-related records.
- Store Thai and English localized text as separate columns for simple querying and backend compatibility.
- Store `PROMOS` and `EVENTS` in one `suggestions` table because they share the same fields.
- Use `store_id` on operational and reporting tables so the model can support more than one branch.
- Use decimal types for money, percentages, confidence, and distance instead of floating point types.

## Suggested Constraints And Indexes

- `stores.code` should be unique.
- `products.sku` should be the primary key or have a unique constraint.
- `sales_hourly` should have a unique constraint on `(store_id, sale_date, hour)`.
- `sales_daily` should have a unique constraint on `(store_id, sale_date)`.
- `sales_monthly` should have a unique constraint on `(store_id, year, month)`.
- `category_sales` should have a unique constraint on `(store_id, category_id, sales_date)`.
- `payment_mix` should have a unique constraint on `(store_id, payment_method_id, sales_date)`.
- Add indexes for common filters:
  - `deliveries(store_id, status)`
  - `deliveries(store_id, eta_time)`
  - `expiring_inventory(store_id, expiry_date)`
  - `low_stock_alerts(store_id, stock_quantity)`
  - `top_products(store_id, sales_date)`
  - `suggestions(store_id, kind, type)`

## Query Impact

- Dashboard pages can fetch sales summaries by `store_id` and date range.
- Revenue pages can join `category_sales` to `categories` and `payment_mix` to `payment_methods`.
- Alert pages can join `expiring_inventory` and `low_stock_alerts` to `products`.
- Delivery pages can filter `deliveries` by status and sort by `eta_time`.
- Suggestion pages can filter `suggestions` by `kind` or `type`.

## Verification

No migration or database test command was run because this request only converts frontend mock data into a database table and field list.
