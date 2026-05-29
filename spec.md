# Spec: Mini BigC Backend

## Objective

Build and maintain the Mini BigC backend as the operational API for branch dashboards, revenue analysis, stock alerts, deliveries, suggestions, and AI-assisted branch questions.

The backend must provide:

- Health and readiness signal for the API and database.
- Role-restricted REST APIs under `/api/v1`.
- Store-scoped dashboard aggregate data for the frontend.
- Raw operational data endpoints for stores, sales, products, inventory, deliveries, and suggestions.
- AI chat endpoint that answers branch questions using role-aware dashboard context.
- PostgreSQL-backed persistence with deterministic seed/bootstrap behavior for local development and tests.

Success means the frontend can load protected operational pages from backend APIs, manager-only data stays restricted from staff, and API behavior is testable through Go unit tests and Playwright API tests.

## Assumptions

- Backend package is `miniBigCProject_Backend`.
- Runtime API port defaults to `5001`.
- PostgreSQL runs locally on host port `5433`.
- Current implementation uses `net/http`, `chi`, GORM, and PostgreSQL.
- Current role authorization uses frontend-provided role headers, not a production JWT session.
- First-release roles are `manager` and `staff`.
- Dashboard APIs are read-only for the current release.
- Database bootstrap uses GORM `AutoMigrate`, index creation, and seed loading.
- AI chat uses Gemini, with optional MCP tool integration.

## Tech Stack

- Language: Go 1.26.3 as declared in `go.mod`.
- Router: `github.com/go-chi/chi/v5`.
- HTTP server: Go `net/http`.
- Database: PostgreSQL.
- ORM: GORM with `gorm.io/driver/postgres`.
- Env loading: `github.com/joho/godotenv`.
- AI client: `google.golang.org/genai`.
- API tests: Playwright.
- Load tests: k6 scripts in `tests/load`.
- Local DB tooling: Docker Compose with PostgreSQL and pgAdmin.

Note: Project AI-resource rules describe a stricter future target using clean architecture, `fasthttp`, and `pgx`. The current spec documents the existing backend. Any migration toward those rules should be planned as a separate architecture change, not mixed into feature work.

## Commands

Run commands from `miniBigCProject_Backend`.

```bash
docker compose up -d
go run ./cmd/server/main.go
go run ./cmd/seed/main.go
go build ./...
go test ./... -count=1
go vet ./...
make check
npm run test:api
make loadtest-smoke
make loadtest
```

Local URLs:

```text
API:      http://localhost:5001
Health:   http://localhost:5001/health
Postgres: localhost:5433
pgAdmin:  http://localhost:5050
```

## Runtime Configuration

Configuration is loaded from environment variables with defaults in `internal/util/env.go`.

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_NAME` | `Mini BigC API` | Root endpoint display name |
| `PORT` | `5001` | HTTP listen port |
| `CORS_ORIGINS` | Local frontend origins | Allowed browser origins |
| `DB_USER` | `admin` | PostgreSQL user |
| `DB_PASSWORD` | `root` | PostgreSQL password |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5433` | PostgreSQL host port |
| `DB_NAME` | `test_db` | PostgreSQL database |
| `GEMINI_API_KEY` | empty | Gemini API key |
| `GEMINI_MODEL` | `gemini-2.5-flash` | Gemini model name |
| `AI_TIMEOUT_SECONDS` | `90` | AI request timeout |
| `MCP_ENABLED` | `true` | Enables local MCP tool bridge |
| `MCP_SERVER_COMMAND` | `mcp_server/venv/bin/python` | MCP process command |
| `MCP_SERVER_ARGS` | `mcp_server/server.py` | MCP process arguments |
| `MCP_TIMEOUT_SECONDS` | `10` | MCP call timeout |
| `MINIBIGC_API_BASE_URL` | `http://localhost:{PORT}` | Backend URL exposed to MCP tools |

Secrets such as `GEMINI_API_KEY` must come from environment or local `.env`; never commit them.

## Project Structure

```text
miniBigCProject_Backend/
├── cmd/
│   ├── server/                 # API server entrypoint
│   └── seed/                   # Manual seed entrypoint
├── internal/
│   ├── db/                     # Database bootstrap, migrations, indexes
│   ├── handler/                # HTTP handlers
│   ├── middleware/             # CORS and role authorization
│   ├── model/                  # GORM and JSON models
│   ├── repo/                   # GORM repository implementations
│   ├── router/                 # chi route wiring and dependency setup
│   ├── service/                # Business/application services
│   └── util/                   # Env config and seed utilities
├── mcp_server/                 # Local MCP server used by AI integration
├── tests/
│   ├── api/                    # Playwright API tests
│   └── load/                   # k6 load tests
├── bruno/                      # API collection
├── docs/AI-resource/           # Project AI rules
├── database.md                 # Database schema reference
├── docker-compose.yml          # PostgreSQL and pgAdmin local stack
└── Makefile                    # Build, test, run, seed, API/load commands
```

## Architecture

Current request flow:

```text
HTTP request
→ chi router
→ middleware
→ handler
→ service
→ repository
→ GORM
→ PostgreSQL
```

Responsibilities:

- `cmd/server`: load env, open DB, configure pool, bootstrap database, start server.
- `internal/router`: instantiate repositories, services, handlers, middleware, and routes.
- `internal/middleware`: enforce CORS and role restrictions.
- `internal/handler`: parse HTTP input, validate path/body basics, call services, map errors to status codes.
- `internal/service`: coordinate application logic, role-aware context building, AI calls, and repository calls.
- `internal/repo`: perform database reads and assemble dashboard aggregates.
- `internal/model`: define current JSON and GORM shapes.
- `internal/db`: migrate schema, ensure indexes, and seed data when needed.

Architecture boundaries:

- Handlers should not contain query logic.
- Repositories should not contain HTTP response logic.
- Services should wrap errors with useful context.
- Route setup should remain centralized in `internal/router/router.go`.
- New features should preserve the existing handler → service → repo flow unless an approved architecture migration is underway.

## API Surface

### Public Routes

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/` | Welcome message |
| `GET` | `/health` | Database-backed health check |
| `POST` | `/api/ai/chat` | Legacy AI chat path |

### Shared Manager/Staff Routes

These routes require `X-User-Role`, `X-Frontend-Role`, or `X-Role` with value `manager` or `staff`.

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/v1/ai/chat` | AI branch assistant |
| `GET` | `/api/v1/dashboard` | Default store dashboard aggregate |
| `GET` | `/api/v1/stores` | Store list |
| `GET` | `/api/v1/stores/{storeID}` | Store detail |
| `GET` | `/api/v1/stores/{storeID}/dashboard` | Store dashboard aggregate |
| `GET` | `/api/v1/stores/{storeID}/inventory-items` | Store inventory items |
| `GET` | `/api/v1/stores/{storeID}/expiring-inventory` | Store expiring inventory |
| `GET` | `/api/v1/stores/{storeID}/low-stock-alerts` | Store low-stock alerts |
| `GET` | `/api/v1/stores/{storeID}/deliveries` | Store deliveries |
| `GET` | `/api/v1/categories` | Categories |
| `GET` | `/api/v1/products` | Products |
| `GET` | `/api/v1/inventory-items` | Inventory items |
| `GET` | `/api/v1/expiring-inventory` | Expiring inventory |
| `GET` | `/api/v1/low-stock-alerts` | Low-stock alerts |
| `GET` | `/api/v1/deliveries` | Deliveries |

### Manager-Only Routes

These routes require role `manager`.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/stores/{storeID}/sales/hourly` | Store hourly sales |
| `GET` | `/api/v1/stores/{storeID}/sales/daily` | Store daily sales |
| `GET` | `/api/v1/stores/{storeID}/sales/monthly` | Store monthly sales |
| `GET` | `/api/v1/stores/{storeID}/category-sales` | Store category sales |
| `GET` | `/api/v1/stores/{storeID}/payment-mix` | Store payment mix |
| `GET` | `/api/v1/stores/{storeID}/top-products` | Store top products |
| `GET` | `/api/v1/stores/{storeID}/suggestions` | Store suggestions |
| `GET` | `/api/v1/payment-methods` | Payment methods |
| `GET` | `/api/v1/sales/hourly` | Raw hourly sales |
| `GET` | `/api/v1/sales/daily` | Raw daily sales |
| `GET` | `/api/v1/sales/monthly` | Raw monthly sales |
| `GET` | `/api/v1/category-sales` | Raw category sales |
| `GET` | `/api/v1/payment-mix` | Raw payment mix |
| `GET` | `/api/v1/top-products` | Raw top products |
| `GET` | `/api/v1/suggestions` | Raw suggestions |

## Authentication And Authorization

Current release authorization is role-header based.

Accepted role headers:

```text
X-User-Role
X-Frontend-Role
X-Role
```

Rules:

- Missing role on protected routes returns `401`.
- Unsupported role on protected routes returns `403`.
- Role normalization trims spaces and lowercases values.
- `OPTIONS` requests should pass through role middleware for CORS compatibility.
- Staff must not access manager-only routes.
- AI chat must validate role before building dashboard context.

Future production auth:

- JWT parsing should happen in middleware.
- Controllers should read identity and role from request context, not request body.
- Backend should enforce store-level access independently of frontend masking.

## Response And Error Contract

Current implementation returns raw JSON payloads for success and a simple error body:

```json
{ "error": "message" }
```

Current status code expectations:

| Status | Use |
| --- | --- |
| `200` | Successful reads and AI replies |
| `400` | Invalid JSON, missing AI message/role, invalid `storeID` |
| `401` | Missing role on role-protected API |
| `403` | Role lacks permission or AI store access denied |
| `404` | Store or record not found |
| `500` | Internal database/configuration failure |
| `502` | AI provider/MCP failure |
| `504` | AI request timeout |

OpenAPI generation rules in `docs/AI-resource/openapi-gl 1.md` use a stricter `{ code, data }` success envelope and `{ code, message }` error envelope. Do not change existing API response shape without coordinating frontend changes and updating tests.

## Dashboard Aggregate Contract

`GET /api/v1/dashboard` and `GET /api/v1/stores/{storeID}/dashboard` return `model.DashboardData`.

Required top-level fields:

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

Requirements:

- Store ID must be positive for store-scoped routes.
- Missing store returns `404`.
- Arrays should return `[]`, not `null`, where possible.
- Sales and inventory rows must be scoped by `store_id`.
- Joined display fields such as category names, payment method names, and product names must be included where the frontend depends on them.
- Response ordering should be stable for charts and lists.

## Data Domain Requirements

### Stores

Stores contain branch metadata, manager profile labels, staff profile labels, and bilingual address/name fields.

Acceptance criteria:

- Store codes are unique.
- Thai and English names are required.
- Default dashboard can resolve a default store ID.

### Sales

Sales data includes hourly, daily, monthly, category, payment mix, and top-product records.

Acceptance criteria:

- Hour values must be `0` through `23`.
- Month values must be `1` through `12`.
- Store/date uniqueness must prevent duplicate chart points.
- Manager-only sales endpoints must reject staff requests.

### Inventory Alerts

Inventory includes current stock, low-stock alerts, and expiring inventory.

Acceptance criteria:

- Low-stock alerts include SKU, category names, current stock, reorder quantity, and location.
- Expiring inventory includes expiry date, quantity, price, category names, and location.
- Alert reads are available to both manager and staff.

### Deliveries

Deliveries represent customer fulfillment state.

Acceptance criteria:

- Delivery IDs are stable strings.
- Status values must support frontend states such as preparing, en route, and delivered.
- Delivery records include bilingual customer, address, and driver labels.
- Shared delivery APIs are available to manager and staff.

### Suggestions

Suggestions represent action recommendations for promotions, events, operations, inventory, risk, or customer insight.

Acceptance criteria:

- Suggestions include type/kind, icon, bilingual title/description, upside value, confidence, duration, and target.
- Manager-only suggestion APIs reject staff.
- Confidence values should stay in a normalized range expected by the frontend.

## Database Requirements

Database schema reference lives in `database.md`.

Current database behavior:

- `internal/db/bootstrap.go` runs GORM `AutoMigrate`.
- `ensureIndexes` creates idempotent indexes.
- `internal/util/seed.go` loads default dashboard data when needed.
- `cmd/seed` can manually seed data.

Tables:

- `stores`
- `categories`
- `payment_methods`
- `products`
- `sales_hourly`
- `sales_daily`
- `sales_monthly`
- `category_sales`
- `payment_mix`
- `top_products`
- `inventory_items`
- `expiring_inventory`
- `low_stock_alerts`
- `deliveries`
- `suggestions`

Database change rules:

- Prefer additive schema changes.
- Preserve unique indexes used to prevent duplicate chart and reporting rows.
- Preserve store/date indexes used by dashboard read paths.
- Update `database.md` when schema behavior changes.
- Add repository tests or API tests when query behavior changes.
- Do not remove seed data needed by API tests without updating those tests.

## AI Chat Requirements

AI chat endpoint:

```http
POST /api/v1/ai/chat
```

Legacy endpoint:

```http
POST /api/ai/chat
```

Request:

```json
{
  "message": "What should this branch focus on today?",
  "role": "manager"
}
```

Response:

```json
{
  "reply": "..."
}
```

Behavior:

- Trim `message` and `role`.
- If body role is empty, read role from accepted role headers.
- Reject empty message with `400`.
- Reject missing role with `400`.
- Build dashboard context using role-aware store access.
- Deny manager-only context to unauthorized roles.
- Use Gemini with configured model and timeout.
- Use MCP tools when enabled, but skip recursive AI chat tool calls.
- Return stable error status for missing API key, timeout, cancellation, empty AI response, unresolved function calls, or provider failure.

Security requirements:

- Do not log prompts together with sensitive customer or credential data.
- Do not log API keys or provider tokens.
- AI context must respect role restrictions.
- AI failure must not affect non-AI dashboard APIs.

## MCP Server Requirements

The bundled MCP server in `mcp_server` exposes read-only backend data tools to Gemini.

Requirements:

- Start command comes from `MCP_SERVER_COMMAND` and `MCP_SERVER_ARGS`.
- Tool calls use `MINIBIGC_API_BASE_URL`.
- MCP timeout must be enforced.
- Disable MCP with `MCP_ENABLED=false`.
- MCP integration must not recursively call AI chat.
- Backend should continue with dashboard-context-only behavior when MCP is disabled.

## CORS Requirements

CORS middleware must allow configured local frontend origins by default:

- `http://localhost:3000`
- `http://localhost:3001`
- `http://localhost:5173`
- `http://127.0.0.1:3000`
- `http://127.0.0.1:5173`

Requirements:

- CORS configuration must be environment-driven.
- Preflight requests must succeed for protected endpoints.
- Do not use wildcard origins for production without explicit approval.

## Code Style

Use small interfaces, explicit dependencies, and request-scoped contexts.

```go
type DashboardService interface {
	GetDefaultDashboard(ctx context.Context) (model.DashboardData, error)
	GetDashboardByStoreID(ctx context.Context, storeID int64) (model.DashboardData, error)
}

type dashboardService struct {
	repository repo.DashboardRepository
}

func NewDashboardService(repository repo.DashboardRepository) DashboardService {
	return &dashboardService{repository: repository}
}

func (service *dashboardService) GetDashboardByStoreID(
	ctx context.Context,
	storeID int64,
) (model.DashboardData, error) {
	data, err := service.repository.GetDashboard(ctx, storeID)
	if err != nil {
		return model.DashboardData{}, fmt.Errorf("get dashboard by store id %d: %w", storeID, err)
	}

	return data, nil
}
```

Conventions:

- Run `gofmt` on changed Go files.
- Keep handlers focused on HTTP concerns.
- Keep services focused on orchestration and business rules.
- Keep repositories focused on data access.
- Wrap service/repository errors with context.
- Use table-driven tests for role, service, and repository behavior.
- Avoid global mutable state in production code and tests.
- Do not log credentials, auth headers, API keys, or unmasked PII.

## Testing Strategy

### Go Unit Tests

Use Go tests for:

- Middleware role authorization.
- Handler status mapping and validation.
- Service error wrapping and orchestration.
- Dashboard context building.
- AI service timeout/error behavior.
- Repository behavior where database can be isolated or seeded.
- Env config parsing.
- Seed idempotency.

Command:

```bash
go test ./... -count=1
```

### Build And Static Checks

Commands:

```bash
go build ./...
go vet ./...
make check
```

Expected result:

- Build passes.
- Tests pass.
- Vet passes.
- `go.mod` and `go.sum` remain clean after dependency changes.

### API Tests

Use Playwright API tests in `tests/api`.

Command:

```bash
npm run test:api
```

Required API test coverage:

- `/health` returns `200` when DB is reachable.
- Protected routes return `401` without role.
- Manager/staff shared routes accept both roles.
- Manager-only routes reject staff with `403`.
- Invalid `storeID` returns `400`.
- Missing store returns `404`.
- Dashboard aggregate contains all required top-level fields.
- AI chat validates message and role.

### Load Tests

Use k6 scripts for smoke and dashboard load checks.

Commands:

```bash
make loadtest-smoke
make loadtest
```

Load tests should focus on:

- `/health`
- `/api/v1/dashboard`
- `/api/v1/stores/{storeID}/dashboard`
- Role-protected read endpoints used by the frontend.

## Boundaries

Always:

- Keep role restrictions aligned with frontend `ROLE_RESTRICTED_APIS.md`.
- Keep dashboard aggregate backward-compatible unless frontend is updated in the same change.
- Use context-aware DB calls.
- Preserve database indexes and constraints that protect dashboard read paths.
- Update tests when endpoint behavior changes.
- Update `database.md` when schema behavior changes.
- Keep AI and MCP failures isolated from non-AI APIs.

Ask first:

- Changing API response envelopes.
- Changing role names or permissions.
- Replacing GORM, chi, or `net/http`.
- Migrating toward `fasthttp`, `pgx`, or a new clean architecture layout.
- Adding write endpoints that mutate operational data.
- Adding external services beyond Gemini/MCP.
- Changing seed data shape consumed by frontend or tests.

Never:

- Commit `.env` secrets or API keys.
- Log passwords, tokens, API keys, or raw sensitive user data.
- Return `200` for an actual error condition.
- Let staff access manager-only sales or suggestions endpoints.
- Remove failing tests to make verification pass.
- Introduce direct frontend-only trust for store access in production auth work.
- Make destructive database changes without a migration and rollback plan.

## Success Criteria

- `go build ./...` passes.
- `go test ./... -count=1` passes.
- `go vet ./...` passes.
- `npm run test:api` passes when local DB/server prerequisites are available.
- `/health` returns `200` with a reachable DB.
- Shared protected routes return `401` without role and `200` for manager/staff.
- Manager-only routes return `403` for staff and `200` for manager.
- Dashboard aggregate returns all required fields for a valid store.
- AI chat validates request body, role, and provider errors with stable status codes.
- Database bootstrap is idempotent for local development.

## Implementation Checklist

- [ ] Read `AGENT.md` and relevant `docs/AI-resource/*.md` before backend changes.
- [ ] Read `database.md` before database or repository changes.
- [ ] Confirm endpoint role class before changing a route.
- [ ] Add or update handler/service/repository tests for behavior changes.
- [ ] Add or update Playwright API tests for route contract changes.
- [ ] Run `gofmt` on changed Go files.
- [ ] Run `go test ./... -count=1` after narrow changes.
- [ ] Run `make check` before handoff for backend code changes.
- [ ] Run `npm run test:api` when API behavior changes and local dependencies are available.
- [ ] Update this spec when backend architecture, routes, auth, or data contracts change.

## Detailed Endpoint Specification

### `GET /health`

Purpose:

- Confirm the server can acquire the underlying SQL DB handle.
- Confirm the database responds to `PingContext`.

Success:

```text
200 ok
```

Failure:

```text
503 database unavailable
```

Acceptance criteria:

- Uses request context for ping.
- Does not bootstrap or seed data.
- Does not expose database DSN or credentials.
- Should be lightweight enough for frequent health checks.

### `GET /api/v1/dashboard`

Purpose:

- Return the default store dashboard aggregate.
- Used by frontend initial branch data load and staff default branch flow.

Headers:

```text
X-User-Role: manager | staff
```

Success:

```json
{
  "store": {},
  "hourly_sales": [],
  "daily_sales": [],
  "monthly_sales": [],
  "category_sales": [],
  "payment_mix": [],
  "top_products": [],
  "inventory_items": [],
  "expiring_inventory": [],
  "low_stock_alerts": [],
  "deliveries": [],
  "suggestions": []
}
```

Failure cases:

| Case | Status | Body |
| --- | --- | --- |
| Missing role | `401` | `{ "error": "role is required" }` |
| Role not manager/staff | `403` | `{ "error": "role is not allowed to access this resource" }` |
| No default store | `404` | `{ "error": "record not found" }` |
| Database failure | `500` | `{ "error": "internal server error" }` |

Acceptance criteria:

- Uses the lowest store ID as current default unless a future config changes this.
- Response includes every dashboard aggregate field even if the arrays are empty.
- Repository must scope all child collections to the selected store ID.
- Query ordering must remain stable for frontend charts.

### `GET /api/v1/stores`

Purpose:

- Return store options for manager branch switching and staff branch display.

Authorization:

- `manager` and `staff`.

Acceptance criteria:

- Returns a JSON array.
- Store rows include bilingual names, short names, addresses, manager/staff labels, and initials.
- Ordering should be stable, preferably by `id` or store code.
- Staff visibility is currently allowed; future store-level access may filter rows.

### `GET /api/v1/stores/{storeID}/dashboard`

Purpose:

- Return the complete dashboard aggregate for one store.

Path validation:

- `storeID` must parse as a positive integer.
- Invalid values return `400`.

Authorization:

- `manager` and `staff` in current implementation.
- Future production auth should validate store assignment for staff.

Acceptance criteria:

- Missing store returns `404`.
- Valid store with no child rows returns store plus empty arrays.
- All child data is filtered by `store_id`.
- This endpoint is the preferred frontend source after manager branch selection.

### Store-Scoped Resource Endpoints

Routes:

```text
GET /api/v1/stores/{storeID}
GET /api/v1/stores/{storeID}/inventory-items
GET /api/v1/stores/{storeID}/expiring-inventory
GET /api/v1/stores/{storeID}/low-stock-alerts
GET /api/v1/stores/{storeID}/deliveries
GET /api/v1/stores/{storeID}/sales/hourly
GET /api/v1/stores/{storeID}/sales/daily
GET /api/v1/stores/{storeID}/sales/monthly
GET /api/v1/stores/{storeID}/category-sales
GET /api/v1/stores/{storeID}/payment-mix
GET /api/v1/stores/{storeID}/top-products
GET /api/v1/stores/{storeID}/suggestions
```

Requirements:

- Every route must validate positive `storeID`.
- Shared operational endpoints may allow manager/staff.
- Revenue and suggestions endpoints must remain manager-only.
- Store-scoped endpoints must never return rows from another store.
- Store-scoped list endpoints should share repository logic with dashboard aggregate where practical to avoid contract drift.

### Raw Resource Endpoints

Routes:

```text
GET /api/v1/categories
GET /api/v1/products
GET /api/v1/inventory-items
GET /api/v1/expiring-inventory
GET /api/v1/low-stock-alerts
GET /api/v1/deliveries
GET /api/v1/payment-methods
GET /api/v1/sales/hourly
GET /api/v1/sales/daily
GET /api/v1/sales/monthly
GET /api/v1/category-sales
GET /api/v1/payment-mix
GET /api/v1/top-products
GET /api/v1/suggestions
```

Requirements:

- Shared endpoints expose operational data usable by manager/staff.
- Manager-only endpoints expose commercial performance data.
- Raw endpoints should be stable enough for API tests and MCP tools.
- If pagination is introduced later, it must be a versioned contract change or backward-compatible optional query.

### `POST /api/v1/ai/chat`

Validation:

| Field | Source | Rule |
| --- | --- | --- |
| `message` | JSON body | Required after trimming |
| `role` | JSON body or role header | Required after trimming |

Success:

```json
{ "reply": "..." }
```

Failure cases:

| Case | Status | Body |
| --- | --- | --- |
| Invalid JSON | `400` | `{ "error": "request body must be valid JSON" }` |
| Empty message | `400` | `{ "error": "message is required" }` |
| Missing role | `400` | `{ "error": "role is required" }` |
| Role/store forbidden | `403` | role/store access error |
| Missing Gemini API key | `500` | `{ "error": "GEMINI_API_KEY is not configured" }` |
| Empty provider response | `502` | `{ "error": "AI response is empty" }` |
| Unresolved Gemini function calls | `502` | `{ "error": "Gemini function calls were not resolved" }` |
| Provider unavailable | `502` | `{ "error": "AI service unavailable" }` |
| Request canceled | `408` | `{ "error": "AI request canceled" }` |
| Timeout | `504` | `{ "error": "AI request timed out" }` |

Acceptance criteria:

- Role header fallback works when request body omits `role`.
- AI service uses request context and configured timeout.
- AI errors are mapped to stable status codes.
- Logs do not include secrets.
- Non-AI endpoints work even when AI is misconfigured.

## Detailed Model Contract

### Store

JSON fields:

- `id`
- `code`
- `name_th`
- `name_en`
- `short_name_th`
- `short_name_en`
- `address_th`
- `address_en`
- `manager_name_th`
- `manager_name_en`
- `manager_initials`
- `staff_name_th`
- `staff_name_en`
- `staff_initials`

Database requirements:

- `code` is unique.
- Names, short names, addresses, and profile labels are non-empty.
- Initials are short display strings and should not exceed the configured DB size.

Frontend dependency:

- App shell, branch selector, dashboard headings, and profile page depend on these fields.

### Hourly Sale

JSON fields:

- `id`
- `store_id`
- `sale_date`
- `hour`
- `sales_value`
- `comparison_sales_value`

Database requirements:

- Unique row per `store_id`, `sale_date`, and `hour`.
- `hour` between `0` and `23`.
- Numeric values are non-negative for normal seed data.

Repository ordering:

```sql
ORDER BY sale_date, hour
```

Frontend dependency:

- Dashboard and revenue hourly charts require ascending time order.

### Daily Sale

JSON fields:

- `id`
- `store_id`
- `sale_date`
- `sales_value`
- `comparison_sales_value`

Database requirements:

- Unique row per `store_id` and `sale_date`.
- Ordered by `sale_date`.

Frontend dependency:

- Revenue MTD and dashboard daily trend calculations.

### Monthly Sale

JSON fields:

- `id`
- `store_id`
- `year`
- `month`
- `sales_value`

Database requirements:

- Unique row per `store_id`, `year`, and `month`.
- `month` between `1` and `12`.
- Ordered by `year`, then `month`.

Frontend dependency:

- YTD and monthly dashboard/revenue views.

### Category Sale

JSON fields:

- `id`
- `store_id`
- `category_id`
- `name_th`
- `name_en`
- `color`
- `sales_date`
- `sales_value`
- `share`
- `trend_percent`

Repository requirements:

- Join `category_sales` to `categories`.
- Select bilingual category names and color.
- Order by `sales_value DESC`.

Frontend dependency:

- Revenue category pacing, donut chart, dashboard facts, and suggestions source signals.

### Payment Mix

JSON fields:

- `id`
- `store_id`
- `sales_date`
- `payment_method_id`
- `name_th`
- `name_en`
- `share`

Repository requirements:

- Join `payment_mix` to `payment_methods`.
- For dashboard aggregate, load latest `sales_date` for the store.
- Order by `share DESC`.

Frontend dependency:

- Revenue payment mix display and export.

### Top Product

JSON fields:

- `id`
- `store_id`
- `sku`
- `name_th`
- `name_en`
- `sales_date`
- `sold_quantity`
- `sales_value`
- `trend_percent`

Repository requirements:

- Join `top_products` to `products`.
- Order by `sales_value DESC`.

Frontend dependency:

- Dashboard top products, revenue movers, export.

### Inventory Item

JSON fields:

- `id`
- `store_id`
- `sku`
- `name_th`
- `name_en`
- `stock_quantity`
- `reorder_quantity`
- `location_code`
- `price`

Repository requirements:

- Join `inventory_items` to `products`.
- Order by `location_code`, then `sku`.

Frontend dependency:

- Stock context, alert calculations, AI context.

### Expiring Inventory

JSON fields:

- `id`
- `store_id`
- `sku`
- `name_th`
- `name_en`
- `category_id`
- `category_th`
- `category_en`
- `expiry_date`
- `stock_quantity`
- `location_code`
- `price`

Repository requirements:

- Join `expiring_inventory` to `products` and `categories`.
- Order by `expiry_date`, then `stock_quantity DESC`.

Frontend dependency:

- Alerts page expiry tab, value-at-risk, suggestions generated from expiry risk.

### Low Stock Alert

JSON fields:

- `id`
- `store_id`
- `sku`
- `name_th`
- `name_en`
- `category_id`
- `category_th`
- `category_en`
- `stock_quantity`
- `reorder_quantity`
- `location_code`

Repository requirements:

- Join `low_stock_alerts` to `products` and `categories`.
- Order by `stock_quantity`, then `sku`.

Frontend dependency:

- Alerts page OOS tab, dashboard alert count, suggestions generated from low-stock risk.

### Delivery

JSON fields:

- `id`
- `store_id`
- `customer_name_th`
- `customer_name_en`
- `address_th`
- `address_en`
- `item_count`
- `order_value`
- `driver_name_th`
- `driver_name_en`
- `status`
- `eta_time`
- `is_late`
- `distance_km`

Repository ordering:

```sql
ORDER BY eta_time, id
```

Requirements:

- `status` values must remain compatible with frontend domain statuses.
- `order_value` is manager-sensitive even if current shared endpoint returns it.
- `is_late` must be a boolean, not inferred on the frontend.

### Suggestion

JSON fields:

- `id`
- `store_id`
- `kind`
- `icon`
- `title_th`
- `title_en`
- `description_th`
- `description_en`
- `upside_value`
- `confidence`
- `duration_th`
- `duration_en`
- `target_th`
- `target_en`
- `type`

Repository ordering:

```sql
ORDER BY kind DESC, upside_value DESC
```

Requirements:

- `kind` must be one of `promo`, `event`, `operation`, `inventory`, `risk`, or `customer`.
- `confidence` should be `0.0` through `1.0`.
- Bilingual title and description are required.

## Repository Query Specification

Dashboard aggregate load order:

1. Store.
2. Hourly sales.
3. Daily sales.
4. Monthly sales.
5. Category sales.
6. Payment mix.
7. Top products.
8. Inventory items.
9. Expiring inventory.
10. Low-stock alerts.
11. Deliveries.
12. Suggestions.

Rules:

- Store lookup happens first; if store is missing, child queries must not run.
- Every query must use `db.WithContext(ctx)`.
- Child query errors must include a short operation label such as `query hourly sales`.
- Joins must be inner joins only when missing related metadata is considered invalid data.
- If related metadata may be absent in future, change to left joins deliberately and update frontend null handling.
- Avoid introducing N+1 query loops; prefer joins or batched queries.

Performance expectations:

- Dashboard aggregate is currently multiple queries; each must use store/date/category indexes where applicable.
- Endpoint should be fast enough for manual refresh from the frontend.
- If dashboard latency grows, first optimize query/index strategy before adding frontend workarounds.

## Database Invariants

The database should preserve these invariants:

- Every operational row references a valid `store_id`.
- Every product-linked row references a valid `sku`.
- Every category-linked row references a valid `category_id`.
- Every payment mix row references a valid `payment_method_id`.
- Sales values, quantities, shares, and confidence values use sensible numeric ranges.
- Bilingual display strings are not empty for records rendered by frontend.
- Seed data is deterministic enough for tests to assert route behavior.

Recommended constraints or checks:

- `sales_hourly.hour >= 0 AND hour <= 23`.
- `sales_monthly.month >= 1 AND month <= 12`.
- `suggestions.kind IN ('promo','event','operation','inventory','risk','customer')`.
- Unique indexes on store/date/time grains used by charts.
- Store-scoped indexes on dashboard-heavy tables.

## Store Access Model

Current model:

- `manager` can access all configured stores.
- `staff` can access shared routes.
- Store-level staff scoping is not enforced yet.

Target production model:

- Auth middleware validates token.
- Middleware injects user ID, role, and allowed store IDs into request context.
- Store-scoped handlers verify requested `storeID` against allowed store IDs.
- Manager-only routes require role `manager`.
- Staff can access only assigned stores.

Migration requirements:

- Add tests before changing auth behavior.
- Keep current role-header behavior only as an explicit local/demo mode if still needed.
- Frontend must stop relying on local role as authorization source once production auth is enabled.

## AI Context Specification

Dashboard context builder should prepare enough branch data for useful answers without exposing unauthorized information.

Manager context may include:

- Revenue totals and trends.
- Category sales.
- Payment mix.
- Top products.
- Inventory alerts.
- Expiring inventory.
- Deliveries.
- Suggestions.

Staff context may include:

- Operational dashboard summary.
- Stock alerts.
- Expiring inventory.
- Delivery status.
- Non-sensitive branch metadata.

Staff context must not include:

- Manager-only revenue details.
- Manager-only suggestions if backend policy marks them restricted.
- Commercial values that frontend masks for staff.

Prompt/context requirements:

- Keep context compact and structured.
- Include store identity.
- Include role.
- Include only data needed to answer the user question.
- Respect request cancellation and timeout.
- Avoid leaking internal errors into the AI reply.

## MCP Tool Specification

MCP tool bridge requirements:

- Spawn MCP process only when enabled.
- Use configured command and args.
- Enforce timeout per tool call.
- Convert tool schemas into Gemini callable functions where supported.
- Skip the MCP `ai_chat` tool to prevent recursive calls.
- Surface unresolved function calls as `ErrUnresolvedGeminiFunctionCalls`.
- Close or clean up subprocess resources when request lifecycle ends or client is shut down.

Failure handling:

- MCP start failure should not crash server startup unless explicitly configured as required.
- MCP tool timeout should produce AI-service failure for that request only.
- MCP malformed response should be treated as provider/tool failure.

## API Security Matrix

| Route group | No role | Staff | Manager |
| --- | --- | --- | --- |
| `/health` | `200` if DB OK | `200` if DB OK | `200` if DB OK |
| `/api/v1/dashboard` | `401` | `200` | `200` |
| `/api/v1/stores` | `401` | `200` | `200` |
| `/api/v1/stores/{id}/inventory-items` | `401` | `200` | `200` |
| `/api/v1/stores/{id}/deliveries` | `401` | `200` | `200` |
| `/api/v1/stores/{id}/sales/hourly` | `401` | `403` | `200` |
| `/api/v1/payment-methods` | `401` | `403` | `200` |
| `/api/v1/suggestions` | `401` | `403` | `200` |
| `/api/v1/ai/chat` | `400` if no body role/header role | body/header role decides | body/header role decides |

Security requirements:

- Middleware-protected routes must be tested for no role, invalid role, staff, and manager.
- AI chat currently performs its own role validation and should be tested separately from route middleware.
- CORS preflight must not fail because role headers are missing.

## Error Handling Matrix

| Layer | Error source | Required behavior |
| --- | --- | --- |
| Middleware | Missing role | `401` JSON error |
| Middleware | Forbidden role | `403` JSON error |
| Handler | Invalid `storeID` | `400` JSON error |
| Handler | Invalid JSON | `400` JSON error |
| Service | Missing record via repo | Preserve or map to `404` at handler |
| Repository | `gorm.ErrRecordNotFound` | Convert to `repo.ErrNotFound` |
| Repository | DB query error | Wrap with query context |
| AI service | Missing key | `500` |
| AI service | Provider unavailable | `502` |
| AI service | Timeout | `504` |

Rules:

- Error bodies should be JSON on API routes.
- Internal errors should not expose SQL, DSN, stack traces, or credentials.
- Log messages should help developers locate failing subsystem without leaking sensitive data.

## Observability And Logging

Current code uses standard `log` in the server and AI handler.

Requirements:

- Server startup logs port only, not secrets.
- Database connection failures can be fatal on startup.
- Health check failures should not log noisy messages on every probe unless needed.
- AI failures should log a concise error class and wrapped error.
- Future structured logging should include request IDs if middleware is added.

Recommended future fields:

- route
- method
- status
- latency
- role
- storeID
- error code

Never log:

- `GEMINI_API_KEY`
- Authorization headers
- Passwords
- Raw tokens
- Full customer PII if avoidable

## Test Data Requirements

Seed data should include at least:

- One manager-visible store.
- One staff-visible store context.
- Hourly sales with current and comparison values.
- Daily sales for MTD charts.
- Monthly sales for YTD charts.
- Multiple categories with colors.
- Multiple payment methods.
- Top products with positive and negative trends.
- Inventory items above and below reorder thresholds.
- Expiring inventory with different expiry dates.
- Deliveries in `preparing`, `enRoute`, and `delivered` states.
- At least one late delivery.
- Promotion and event suggestions.

Test data must allow API tests to assert:

- Shared route success.
- Manager-only route success/failure by role.
- Dashboard aggregate shape.
- Store-scoped filtering.
- Non-empty arrays for primary frontend pages.

## API Test Matrix

Required route contract tests:

| Test | Setup | Expected |
| --- | --- | --- |
| Health OK | DB running | `GET /health` returns `200` |
| Health DB down | DB unavailable if feasible | `GET /health` returns `503` |
| Missing role | No role header | Protected route returns `401` |
| Invalid role | `X-User-Role: viewer` | Protected route returns `403` |
| Staff shared route | `X-User-Role: staff` | Shared route returns `200` |
| Manager shared route | `X-User-Role: manager` | Shared route returns `200` |
| Staff manager-only route | `X-User-Role: staff` | Manager-only route returns `403` |
| Manager manager-only route | `X-User-Role: manager` | Manager-only route returns `200` |
| Invalid store ID | `/stores/abc/dashboard` | Returns `400` |
| Missing store ID | Valid missing numeric ID | Returns `404` |
| Dashboard shape | Valid manager request | All aggregate keys present |
| AI invalid JSON | Bad request body | Returns `400` |
| AI missing message | Empty message | Returns `400` |
| AI missing role | No body/header role | Returns `400` |

## Go Test Matrix

Suggested unit tests:

- `middleware.RequireRole` accepts configured roles.
- `middleware.RequireRole` normalizes case and whitespace.
- `middleware.RequireRole` allows `OPTIONS`.
- `parseStoreID` rejects non-numeric, zero, and negative IDs.
- `writeServiceError` maps `repo.ErrNotFound` to `404`.
- `DashboardService.GetDefaultDashboard` wraps default-store errors.
- `DashboardService.GetDashboardByStoreID` wraps store-specific errors.
- `DataService` wraps repository errors with operation names.
- `LoadConfig` handles invalid timeout values with defaults.
- `roleFromAIRequestHeader` checks accepted headers in priority order.
- `AIHandler.Chat` validates JSON, message, role, and service errors.

Repository tests, when DB-backed:

- Dashboard child rows are scoped to store ID.
- Missing store returns `repo.ErrNotFound`.
- Payment mix uses latest sales date for the store.
- Joined category/product/payment names are populated.
- Ordering matches frontend expectations.

## Load And Performance Criteria

Smoke load test should verify:

- API remains responsive under a small burst.
- Health endpoint stays lightweight.
- Dashboard endpoint does not produce DB connection pool exhaustion.

Dashboard load test should monitor:

- P95 latency for `/api/v1/dashboard`.
- Error rate.
- DB connection pool saturation.
- Slow query candidates.

Initial acceptable local targets:

- Health p95 under 100ms.
- Dashboard p95 under 1000ms on seeded local data.
- Error rate 0% under smoke test conditions.

These are local development targets, not production SLA commitments.

## Deployment And Local Operations

Local startup sequence:

1. Start Docker Compose.
2. Wait for PostgreSQL.
3. Start backend server.
4. Server loads env.
5. Server opens DB connection.
6. Server pings DB.
7. Server bootstraps schema and seed data.
8. Server listens on configured port.

Operational requirements:

- Startup should fail fast if DB cannot connect.
- Bootstrap should be idempotent.
- Seed should not duplicate unique data.
- Shutdown should close SQL connection pool.
- Future graceful shutdown should stop accepting new requests and allow in-flight requests to finish.

## OpenAPI And Contract Documentation

OpenAPI generation is governed by `docs/AI-resource/openapi-gl 1.md`.

Current blocker:

- OpenAPI generation requires complete task files under the expected task resource path.

If OpenAPI is generated later:

- Use OpenAPI 3.x YAML.
- Define schemas under `components/schemas`.
- Avoid inline schemas.
- Decide whether to preserve current raw response bodies or migrate to the required envelope.
- Update frontend API types from the OpenAPI source or verify them manually.
- Add API tests for every documented endpoint.

## Backend-Frontend Alignment

The backend and frontend specs must stay aligned on:

- `/api/v1` route paths.
- Role restrictions.
- Dashboard aggregate field names.
- Date/time field parseability.
- Delivery status values.
- Suggestion confidence range.
- AI chat request and response shape.
- Error status behavior.
- CORS local origins.

Contract change process:

1. Update backend spec.
2. Update frontend spec.
3. Update backend handler/service/repository code.
4. Update frontend API types/mappers.
5. Update API tests.
6. Update frontend unit/E2E tests.
7. Run backend and frontend verification.

## Suggested Delivery Phases

### Phase 1: Stabilize Current Read APIs

- Add missing API tests for every current route.
- Add repository tests for dashboard aggregate ordering and joins.
- Confirm seed idempotency.
- Confirm frontend mappers match backend JSON.

### Phase 2: Production Auth Design

- Specify JWT claims.
- Add auth middleware design.
- Define store-access model.
- Preserve local/demo mode only if needed.
- Update staff access rules.

### Phase 3: Write API Planning

- Specify alert acknowledgment endpoint.
- Specify delivery status update endpoint.
- Specify profile update endpoint.
- Specify suggestion review/launch endpoint.
- Add validation and conflict rules before implementation.

### Phase 4: API Contract Formalization

- Create task files for OpenAPI generation.
- Decide envelope migration.
- Generate or maintain OpenAPI.
- Use contract tests to prevent drift.

### Phase 5: Architecture Modernization Decision

- Decide whether to keep GORM/chi/net-http or migrate to AI-resource target stack.
- If migrating, do it in a dedicated branch and not alongside feature changes.
- Define adapter compatibility and data migration plan.

## Open Questions

- When should role-header auth be replaced with JWT middleware?
- Should API responses migrate to the `{ code, data }` and `{ code, message }` envelopes required by the OpenAPI directive?
- Should staff store access be scoped to a specific store in backend authorization?
- Which endpoints should become write-enabled first: alerts acknowledgment, delivery status, profile, or suggestions?
- Should the backend migrate from GORM to pgx as the AI-resource rules specify?
- Should OpenAPI be generated from task files for the current API surface?
- What production SLA is required for dashboard refresh latency?
- Should AI chat responses be cached or audited for operational review?
