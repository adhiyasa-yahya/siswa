# SISWA — Sistem Item Stok Warehouse & Aset

Inventory system for warehouse items, categories, rack locations and per-location stock.

| Layer | Stack | Port |
| --- | --- | --- |
| Backend | Go 1.23 · Gin · pgx/v5 · PostgreSQL 16 · Redis 7 | **3069** |
| Frontend | React 19 · TypeScript · Vite · Tailwind CSS v4 · React Router v6 · TanStack Query v5 | **4069** |
| Docs | Hand-written OpenAPI 3.0 + Swagger UI | `/swagger/index.html` |

---

## Table of contents

1. [Quick start](#quick-start)
2. [Repository structure](#repository-structure)
3. [Architecture](#architecture)
4. [Database schema & indexes](#database-schema--indexes)
5. [API reference](#api-reference)
6. [Response envelope & error codes](#response-envelope--error-codes)
7. [Business rules](#business-rules)
8. [Realtime request logging](#realtime-request-logging)
9. [Stock mutation log](#stock-mutation-log)
10. [Seeder](#seeder)
11. [Frontend](#frontend)
12. [Local development (without Docker)](#local-development-without-docker)
13. [Verification log](#verification-log)
14. [Design decisions](#design-decisions)
15. [Original prompt & AI model](#original-prompt--ai-model)

---

## Quick start

Requirements: Docker + Docker Compose.

```bash
docker compose up -d --build
```

That is all — migrations run automatically on boot and the database is seeded with
1 000 items.

| What | URL |
| --- | --- |
| Frontend (SPA) | <http://localhost:4069> |
| Backend API | <http://localhost:3069/api/v1> |
| Swagger UI | <http://localhost:3069/swagger/index.html> |
| OpenAPI document | <http://localhost:3069/swagger/openapi.yaml> |
| Health check | <http://localhost:3069/health> |
| Realtime log stream (SSE) | <http://localhost:3069/api/v1/logs/stream> |

```bash
docker compose logs -f backend     # application + stock mutation log
docker compose down                # stop
docker compose down -v             # stop and wipe the database (re-seeds next boot)
```

> The frontend container also proxies `/api` and `/swagger` to the backend, so
> <http://localhost:4069/api/v1/items> works too (same origin, no CORS needed).
>
> Swagger UI loads its assets from the unpkg CDN, so that page needs internet
> access. The OpenAPI document itself is embedded in the binary and always
> available offline.

---

## Repository structure

```
siswa/
├── backend/
│   ├── cmd/api/                  # entrypoint (main.go)
│   ├── internal/
│   │   ├── cache/                # redis client
│   │   ├── config/               # env configuration
│   │   ├── database/             # pgx pool + embedded migration runner
│   │   ├── dto/                  # request/response contracts + query sanitising
│   │   ├── errs/                 # error catalogue (codes 0 / 99 / 1xxx / 2xxx)
│   │   ├── handler/              # gin handlers (transport only)
│   │   ├── logger/               # slog setup + stock mutation audit log
│   │   ├── logstore/             # realtime request log (redis + postgres + SSE hub)
│   │   ├── middleware/           # request id, request logger, recovery
│   │   ├── model/                # database entities
│   │   ├── repository/           # SQL data access
│   │   ├── response/             # the single JSON envelope + error filter
│   │   ├── router/               # route table
│   │   ├── seeder/               # 1 000 item auto-seeder
│   │   ├── service/              # business rules
│   │   └── validation/           # validator wiring (JSON field names)
│   ├── migrations/               # SQL migration files (embedded in the binary)
│   ├── docs/                     # openapi.yaml + Swagger UI handler
│   ├── Dockerfile
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── components/           # ui/ (Button, Field, Alert, Badge, Pagination,
│   │   │                         #      SearchBar, PageHeader, Spinner)
│   │   │                         # items/ (ItemTable, ItemFilters, ItemForm)
│   │   │                         # layout/ (AppShell + header menu)
│   │   ├── hooks/                # useDebounce, useClientTable, useItems, useMasters
│   │   ├── pages/                # ItemListPage, ItemFormPage,
│   │   │                         # CategoryListPage, CategoryFormPage,
│   │   │                         # WarehouseListPage, WarehouseFormPage, NotFoundPage
│   │   ├── types/                # api.ts, item.ts
│   │   └── services/             # http.ts, items.ts, masters.ts
│   ├── nginx.conf
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
└── README.md
```

---

## Architecture

```
┌────────────┐   HTTP/JSON    ┌──────────────────────────────────────────┐
│ React 19   │ ─────────────▶ │ Gin router                               │
│ TanStack   │                │  ├─ middleware: request id → recovery →  │
│ Query      │ ◀───────────── │  │   CORS → request logger               │
└────────────┘   envelope     │  ├─ handler  (bind + render envelope)    │
                              │  ├─ service  (business rules)            │
                              │  └─ repository (SQL)                     │
                              └───────┬──────────────────┬───────────────┘
                                      │                  │
                             ┌────────▼───────┐  ┌───────▼────────┐
                             │ PostgreSQL 16  │  │ Redis 7        │
                             │ master + stock │  │ capped log list│
                             │ + request_logs │  │ + pub/sub      │
                             └────────────────┘  └────────────────┘
```

Layering rule: handlers never touch SQL, services never touch `gin.Context`,
repositories never decide business outcomes. Every error leaving a service is an
`*errs.AppError` carrying both the HTTP status and the business code.

---

## Database schema & indexes

| Table | Columns |
| --- | --- |
| `categories` | `id`, `name`, `status` (`active` / `inactive`) |
| `locations` | `id`, `code`, `zone`, `type` (`rack`/`bin`/`pallet`/`cold_storage`/`staging`) |
| `items` | `id`, `sku`, `name`, `category_id`, `unit`, `created_at`, `updated_at`, `deleted_at` |
| `stocks` | `id`, `item_id`, `location_id`, `qty`, `updated_at` |
| `request_logs` | `id`, `method`, `api_url`, `payload`, `executed_time_ms`, `status`, `response_code`, `client_ip`, `user_agent`, `response_bytes`, `datetime` |
| `schema_migrations` | applied migration versions |

Indexes (all created in `backend/migrations/000001_init_schema.up.sql`):

| Index | Target | Purpose |
| --- | --- | --- |
| `ux_items_sku_active` | `items(sku) WHERE deleted_at IS NULL` | **unique SKU** on live rows → 409 |
| `idx_items_sku` | `items(sku)` | lookup by SKU |
| `idx_items_name` | `items(name)` | sort/lookup by name |
| `idx_items_sku_trgm`, `idx_items_name_trgm` | GIN `gin_trgm_ops` | fast `ILIKE '%term%'` search |
| `idx_items_category_id` | `items(category_id)` | category filter |
| `idx_items_deleted_at`, `idx_items_created_at` | `items` | soft-delete filter, default sort |
| `ux_stocks_item_location` | `stocks(item_id, location_id)` | one row per item/location (upsert key) |
| `idx_stocks_item_id` | `stocks(item_id)` | stock per item |
| `idx_stocks_location_id` | `stocks(location_id)` | stock per location |
| `categories_pkey`, `locations_pkey`, `stocks_pkey` | primary keys | `Categories.id`, `Location.id`, `Stock.id` |
| `ux_categories_name`, `idx_categories_status`, `ux_locations_code`, `idx_locations_zone` | master data | uniqueness + filters |
| `idx_request_logs_datetime/method/status` | `request_logs` | log browsing |

Migrations are embedded in the binary and applied on boot inside a transaction,
tracked in `schema_migrations` — restarting never re-applies them.

---

## API reference

Base path: `/api/v1`

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/items` | Tambah item baru (409 on duplicate SKU); optional initial warehouse placement |
| `GET` | `/items` | List item — search, filter by category, pagination, sorting |
| `GET` | `/items/:id` | Detail item |
| `PUT` | `/items/:id` | Update item |
| `DELETE` | `/items/:id` | Hapus item (**soft delete**) |
| `POST` | `/stock/receive` | Penerimaan stok masuk ke lokasi |
| `GET` | `/stock/:item_id` | Cek stok per item (semua lokasi) |
| `GET` | `/categories` | Master kategori (`?status=active\|all`) |
| `POST` | `/categories` | Tambah category (409 on duplicate name) |
| `GET` | `/locations` | Master lokasi gudang |
| `POST` | `/locations` | Tambah location / rak gudang (409 on duplicate code) |
| `GET` | `/logs` | Recent request logs (`?source=redis\|db`, `method`, `status`) |
| `GET` | `/logs/stream` | **Realtime** request log stream (SSE) |
| `GET` | `/error-codes` | Full error code catalogue |
| `GET` | `/health` | Health check (postgres + redis) |

`GET /items` query parameters: `search` (alias `q`, matches name **or** SKU),
`category_id`, `page` (default 1), `limit` (default 10, max 100),
`sort_by` (`id`/`sku`/`name`/`category_id`/`created_at`), `sort_dir` (`asc`/`desc`).

`POST /items` optionally accepts `location_id` and `initial_qty` (default `0`).
When a location is given, the item row and its stock row are written in **one
transaction**, so an item is never left half-placed; `initial_qty: 0` simply
reserves the slot. The placement is appended to the stock mutation log as a
`PLACEMENT` entry.

### Examples

```bash
# create
curl -X POST http://localhost:3069/api/v1/items \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-GUDANG-001","name":"Besi Hollow 4x4 Galvanis","category_id":3,"unit":"batang"}'

# search + filter + pagination
curl "http://localhost:3069/api/v1/items?search=hollow&category_id=3&page=1&limit=10"

# receive stock
curl -X POST http://localhost:3069/api/v1/stock/receive \
  -H 'Content-Type: application/json' \
  -d '{"item_id":1,"location_id":12,"qty":120,"reference":"PO-2026-00231"}'

# stock per item, across every location
curl http://localhost:3069/api/v1/stock/1

# watch the realtime request log
curl -N http://localhost:3069/api/v1/logs/stream
```

Swagger UI documents every endpoint with request/response **examples** for the
success and the error paths: <http://localhost:3069/swagger/index.html>

---

## Response envelope & error codes

Every endpoint — including 404s and panics — answers with the same envelope.

```jsonc
// success
{
  "success": true,
  "code": 0,
  "message": "Items retrieved successfully",
  "data": [ /* ... */ ],
  "meta": { "page": 1, "limit": 10, "total": 1000, "total_pages": 100 }
}

// error
{
  "success": false,
  "code": 1003,
  "message": "SKU already exists",
  "errors": [{ "field": "sku", "reason": "Duplicate entry: SKU-GUDANG-001" }]
}
```

`code` is the business code; `0` = sukses and `99` = generic error, as required.
Specific codes make failures machine-readable:

| Code | Name | HTTP | Meaning |
| ---: | --- | ---: | --- |
| 0 | `SUCCESS` | 200/201 | Request processed successfully |
| 99 | `ERROR` | 500 | Generic / unclassified error |
| 1001 | `VALIDATION_ERROR` | 400 | One or more fields failed validation |
| 1002 | `NOT_FOUND` | 404 | Resource does not exist (or is soft deleted) |
| 1003 | `DUPLICATE_SKU` | 409 | SKU already exists |
| 1004 | `NEGATIVE_STOCK` | 422 | Mutation would make the stock negative |
| 1005 | `INVALID_CATEGORY` | 422 | Category missing or inactive |
| 1006 | `INVALID_LOCATION` | 422 | Location not registered |
| 1007 | `ITEM_DELETED` | 410 | Item has been soft deleted |
| 1008 | `INVALID_PAYLOAD` | 400 | Body is not valid JSON |
| 1009 | `INVALID_QUANTITY` | 400 | Quantity must be positive |
| 1010 | `ROUTE_NOT_FOUND` | 404 | Endpoint/method not registered |
| 1011 | `DUPLICATE_CATEGORY` | 409 | Category name already exists |
| 1012 | `DUPLICATE_LOCATION` | 409 | Location code already exists |
| 2001 | `DATABASE_ERROR` | 500 | Database command failed |
| 2002 | `CACHE_ERROR` | 500 | Redis command failed |
| 2003 | `INTERNAL_ERROR` | 500 | Unexpected server error |

The live list is served by `GET /api/v1/error-codes`.

**Error filter.** `errs.Filter()` normalises *any* error into an `*AppError`, and
`response.Error()` is the only way a handler can emit a failure. Driver level
messages (constraint names, SQL text) are logged server-side and never reach the
client; validation errors are translated into the `errors[]` array with their
JSON field names.

---

## Business rules

| Rule | Where it is enforced |
| --- | --- |
| **SKU unik, duplikat → 409** | `ItemService.Create/Update` checks first (`ExistsSKU`); the partial unique index `ux_items_sku_active` is the final authority and a race is mapped back to code 1003 |
| **Stock tidak boleh negatif — divalidasi di service** | `StockService.Receive` passes a `validate(qtyBefore, delta)` callback that runs **inside** the transaction while the stock row is locked (`SELECT … FOR UPDATE`); the `stocks_qty_non_negative` CHECK constraint is only the second line of defence |
| **Soft delete** | `DELETE /items/:id` sets `deleted_at`; the row stays in the database. Every read filters `deleted_at IS NULL`, so deleted items vanish from list/detail/stock-receive but remain auditable |
| **Setiap mutasi stok dicatat** | `logger.StockLogger` writes each receipt to stdout (structured) **and** appends it as JSON to `STOCK_LOG_FILE` (`backend/logs/stock_mutation.log`, mounted to the host) |

Master data uniqueness follows the same shape: a duplicate category name returns
**409 / code 1011** and a duplicate location code **409 / code 1012**, each
checked in the service first and backed by a unique index.

Concurrency: receiving stock creates the `(item_id, location_id)` row with
`INSERT … ON CONFLICT DO NOTHING`, then locks it with `FOR UPDATE`, so parallel
receipts of the same item/location serialise instead of losing an update.
Creating an item with a placement uses a single transaction for the item and its
stock row, so a failed placement rolls the item back too.

---

## Realtime request logging

Every request lifecycle is recorded as
`method | api url | payload | executed time | status | datetime`
(plus request id, response code, client IP, user agent and response size).

* **Redis** — capped list `siswa:request_logs` (`REQUEST_LOG_MAX_ENTRIES`, default 1 000)
  + pub/sub channel `siswa:request_logs:stream`.
* **PostgreSQL** — durable history in `request_logs`.
* **SSE** — `GET /api/v1/logs/stream` pushes each entry the moment the request finishes.

Writes happen on a background worker with a buffered queue, so logging never adds
latency to the request and never fails it. If Redis is down the API keeps serving
and falls back to PostgreSQL only. Request bodies are truncated at
`REQUEST_LOG_PAYLOAD_LIMIT` bytes (default 2 KB) and multipart uploads are skipped.

```bash
curl -N http://localhost:3069/api/v1/logs/stream   # live
curl "http://localhost:3069/api/v1/logs?limit=10&method=POST&status=409"
curl "http://localhost:3069/api/v1/logs?source=db"
docker compose exec redis redis-cli LRANGE siswa:request_logs 0 4
```

---

## Stock mutation log

```json
{"request_id":"ed961c74-…","type":"RECEIVE","item_id":1001,"sku":"SKU-GUDANG-001",
 "location_id":12,"location_code":"A-03-P04","qty_before":0,"qty_delta":120,
 "qty_after":120,"reference":"PO-2026-00231","note":"Penerimaan vendor",
 "at":"2026-10-19T09:30:48.085176799Z"}
```

Console: `docker compose logs backend | grep "stock mutation"`
File: `backend/logs/stock_mutation.log` (bind-mounted from the container).

---

## Seeder

Runs automatically on boot when `SEED_ENABLED=true` (default) and is **idempotent** —
a table that already has rows is left untouched, so restarts never duplicate data.

| Data | Count |
| --- | --- |
| Categories | 12 (10 active + 2 inactive, so the "inactive category" rule is testable) |
| Locations | 72 (zones A–F × 3 aisles × 4 slots) |
| **Items** | **1 000** (`SKU-00001` … `SKU-01000`) |
| Stock rows | ~1 960 (each item spread over 1–3 locations) |

It uses `COPY` (`pgx.CopyFrom`) with a fixed random seed, so the dataset is
identical on every machine and seeding takes ~300 ms.
Tune with `SEED_ENABLED` / `SEED_ITEM_COUNT`.

---

## Frontend

Header menu: **Daftar Item | Daftar Category | Daftar Warehouse**.
Each menu entry is a list view with its own "add" button:

| Menu | List view | Add button → |
| --- | --- | --- |
| Daftar Item | `/items` | `Tambah Item` → `/items/new` |
| Daftar Category | `/categories` | `Tambah Category` → `/categories/new` |
| Daftar Warehouse | `/warehouses` | `Tambah Location` → `/warehouses/new` |

**Daftar Item — `/items`:**
* Table of SKU, nama, kategori, satuan (+ total stok and item id).
* Search by name/SKU with a **300 ms debounced** input (`useDebounce`).
* Category filter dropdown fed by `GET /categories`.
* **Server-side pagination** (page + page size), page resets on every filter change.
* Buttons to open Create / Edit, and an inline two-step delete confirmation.

**Tambah / Edit Item — `/items/new`, `/items/:id/edit`:**
* Client-side validation: SKU, nama, kategori and unit are mandatory
  (plus length/charset rules) — nothing is sent until the form is valid.
* **Penempatan awal (create only):** an optional *Lokasi gudang* dropdown fed by
  `GET /locations`, plus a *Qty awal* field that stays disabled until a location
  is chosen. The item and its stock row are then created in one request/transaction.
* Submits to `POST /items` or `PUT /items/:id`; the placement is never sent on
  update, because stock is moved through `POST /stock/receive`.
* Loading state on the submit button and on the form while the request runs.
* Success and error messages come straight from the API envelope, including the
  business code; `errors[].field` is mapped back onto the matching input (a 409
  duplicate SKU is shown both as a banner and under the SKU field).

**Daftar Category — `/categories`** — id, nama, status badge (aktif/nonaktif),
debounced search, pagination. `Tambah Category` posts to `POST /categories` with
client-side validation and the same success/error handling (409 → code 1011).

**Daftar Warehouse — `/warehouses`** — id, kode lokasi, zona, tipe badge,
debounced search, pagination. `Tambah Location` posts to `POST /locations`
(code/zone/type, 409 → code 1012).

Composition: `services/` (pure API calls) → `hooks/` (TanStack Query) →
`components/ui` (presentational: `Button`, `Field`, `Alert`, `Badge`,
`Pagination`, `SearchBar`, `PageHeader`, `Spinner`) + `components/items`
(feature) → `pages/`. Items paginate on the server (1 000+ rows); the two master
data lists are small, so they reuse the shared `useClientTable` hook for
debounced search + pagination in the browser. Caching, retry policy and query
keys live in `lib/queryClient.ts`, `hooks/useItems.ts` and `hooks/useMasters.ts`;
4xx responses are never retried.

---

## Local development (without Docker)

Infrastructure only:

```bash
docker compose up -d postgres redis
```

Backend:

```bash
cd backend
cp .env.example .env        # DB_HOST=localhost, REDIS_ADDR=localhost:6379
go run ./cmd/api            # http://localhost:3069
```

Frontend:

```bash
cd frontend
npm install
npm run dev                 # http://localhost:4069, proxies /api to :3069
```

Useful commands:

```bash
cd backend  && go vet ./... && go build ./...
cd frontend && npm run build        # tsc -b + vite build
```

Key environment variables (full list in `backend/.env.example`):

| Variable | Default | Meaning |
| --- | --- | --- |
| `APP_PORT` | `3069` | HTTP port |
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `localhost` / `5432` / `siswa` / `siswa_secret` / `siswa` | PostgreSQL |
| `REDIS_ADDR` | `localhost:6379` | Redis (optional — API degrades gracefully) |
| `REQUEST_LOG_ENABLED` / `_TO_REDIS` / `_TO_DB` | `true` | Request log sinks |
| `REQUEST_LOG_MAX_ENTRIES` | `1000` | Redis list cap |
| `STOCK_LOG_FILE` | `logs/stock_mutation.log` | Stock audit log |
| `SEED_ENABLED` / `SEED_ITEM_COUNT` | `true` / `1000` | Seeder |
| `CORS_ALLOW_ORIGINS` | `http://localhost:4069,…` | Allowed browser origins |

---

## Verification log

Everything below was executed against the running stack (`docker compose up -d --build`).

| # | Check | Result |
| --- | --- | --- |
| 1 | `go vet ./...` + `go build ./...` (golang:1.23 container) | pass |
| 2 | `tsc -b && vite build` | pass |
| 3 | Boot: migrations applied, seeder inserted 1 000 items / 1 961 stock rows / 72 locations | pass (290 ms) |
| 4 | `GET /health` → postgres `up`, redis `up` | pass |
| 5 | `GET /items?page=1&limit=2` → meta `total: 1000, total_pages: 500` | pass |
| 6 | `GET /items?search=hollow` → 26 matches | pass |
| 7 | `POST /items` → 201, SKU normalised to upper case | pass |
| 8 | Duplicate SKU → **409**, code **1003**, `errors[0].field = "sku"` | pass |
| 9 | Empty payload → **400**, code 1001, four field errors (`sku`, `name`, `category_id`, `unit`) | pass |
| 10 | Inactive category → **422**, code 1005 | pass |
| 11 | `POST /stock/receive` → 201, `qty_before` 0 → `qty_after` 120 | pass |
| 12 | `GET /stock/:item_id` → per-location breakdown + `total_qty` 200 | pass |
| 13 | `qty: -5` → 400 (rejected before the mutation) | pass |
| 14 | Unknown location → **422**, code 1006 | pass |
| 15 | `PUT /items/:id` to an existing SKU → **409** | pass |
| 16 | `DELETE /items/:id` → row still present with `deleted_at` set; subsequent `GET` → 404, receive → 404 | pass |
| 17 | Unknown route → 404 with code **1010** in the standard envelope | pass |
| 18 | Stock mutation written to console **and** `backend/logs/stock_mutation.log` | pass |
| 19 | Request logs in Redis (`LLEN` 20) **and** PostgreSQL (21 rows, avg 5.136 ms) | pass |
| 20 | `GET /api/v1/logs/stream` pushed `event: request` entries live | pass |
| 21 | Swagger UI 200, `openapi.yaml` 200 (32 KB) | pass |
| 22 | SPA 200, deep link `/items/5/edit` 200, `/api` proxy through nginx 200 | pass |
| 23 | Browser: item grid renders 1 000 items; debounced search "hollow" → 26 rows | pass |
| 24 | Browser: empty form submit → four inline client-side errors, no request sent | pass |
| 25 | Browser: duplicate SKU → red alert `SKU already exists · code 1003` + inline field error | pass |
| 26 | Browser: valid submit → item created, redirect to the list, total 1 001, no console errors | pass |

Second round, after the menu/master-data change:

| # | Check | Result |
| --- | --- | --- |
| 27 | `go vet ./...` + `go build ./...`, `tsc -b && vite build` | pass |
| 28 | `POST /categories` → 201; duplicate name (different casing) → **409 / 1011** | pass |
| 29 | `POST /locations` → 201, code upper-cased; duplicate code → **409 / 1012** | pass |
| 30 | `POST /locations` with `type: "shelf"` → 400 / 1001 listing the allowed types | pass |
| 31 | `POST /items` with `location_id` + `initial_qty: 75` → 201, `total_qty` 75, stock row at that location | pass |
| 32 | `POST /items` with `location_id` and no qty → 201, slot reserved with qty 0 | pass |
| 33 | `POST /items` with unknown location → **422 / 1006**, and no orphan item row created | pass |
| 34 | `POST /items` with `initial_qty: -3` → 400, rejected before any write | pass |
| 35 | Placement written to the mutation log as a `PLACEMENT` entry | pass |
| 36 | OpenAPI re-validated: 10 paths, 15 operations, 25 refs, none broken | pass |
| 37 | Browser: header shows Daftar Item / Daftar Category / Daftar Warehouse, active state follows the section | pass |
| 38 | Browser: `/categories` lists 13 categories with status badges; `/warehouses` lists 73 locations | pass |
| 39 | Browser: item form shows "Penempatan awal", Qty awal disabled until a location is picked | pass |
| 40 | Browser: Tambah Location form → created, redirect to list, count 73 → 74, no console errors | pass |

Test rows created during verification were removed afterwards; the database holds
exactly the seeded 1 000 items, 12 categories and 72 locations.

---

## Design decisions

* **SKU uniqueness is partial** (`WHERE deleted_at IS NULL`). A soft-deleted item
  therefore releases its SKU, which is what warehouse users expect; duplicates
  among live items still return 409.
* **Table names are plural** (`items`, `categories`, `locations`, `stocks`) while
  the entities keep the names from the brief.
* **OpenAPI is hand-written** (`backend/docs/openapi.yaml`) instead of generated
  from annotations: it is embedded in the binary, needs no code-generation step in
  the build, and carries richer request/response examples.
* **Redis is optional.** If it is unreachable the API logs a warning, keeps
  serving, and writes request logs to PostgreSQL only.
* **Stock logging is application level** (console + file) as the brief requires;
  the data needed to add a persistent `stock_mutations` table later is already in
  `logger.Mutation`.
* **The SSE hub de-duplicates by request id**, so a multi-replica deployment can
  fan out through Redis pub/sub without a client seeing an entry twice.
* **Item placement is part of `POST /items`**, not a second call: one request,
  one transaction, no half-created item. Moving stock afterwards stays the job of
  `POST /stock/receive`, so the update endpoint never silently changes quantities.
* **Categories and locations are created, not edited or deleted.** The brief asks
  for a list view and an add button; editing or removing master data that items
  already reference needs its own rules (re-pointing or blocking), so it was left
  out rather than half-done.

---

## Original prompt & AI model

**AI model:** Claude Opus 5 (`claude-opus-5`), via Claude Code.
**Date:** 19 October 2026.
**Human author:** project owner (prompts below), all code generated and verified in-session.

Every prompt of the session is recorded verbatim, in order.

### Prompt 1 — initial build (verbatim)

````text
init project inventory system BE: golang, FE: react19, project name: SISWA (Sistem item Stok Warehouse & Aset)

- main PostgreSQL db entity:
Item: id, sku, name, category_id, unit, created_at Master data barang
Categories: id, name, status Master category item
Location: id, code, zone, type Lokasi rak/zona di gudang
Stock: id, item_id, location_id, qty, updated_at Kuantitas item per lokasi

- Endpoint
POST /api/v1/items Tambah item baru
GET /api/v1/items List item (filter by category + pagination)
GET /api/v1/items/:id Detail item
PUT /api/v1/items/:id Update item
DELETE /api/v1/items/:id Hapus item (soft delete)
POST /api/v1/stock/receive Penerimaan stok masuk ke lokasi
GET /api/v1/stock/:item_id Cek stok per item (semua lokasi)

rules:
• SKU item harus unik; duplikat harus dikembalikan sebagai error 409.
• Stock tidak boleh negatif; validasi wajib dilakukan di level service, bukan hanya di database.
• Soft delete pada item: record tidak dihapus dari database, hanya diberi flag deleted_at.
• Setiap mutasi stok (receive) harus mencatat log transaksi minimal di level aplikasi
(console/file).

additional:
- run BE in port 3069 and FE in 4069
- implementasi realtime looging system, record request masuk sampai selesai, update time executed time, method, datetime, save in redis/db
    method | api url | payload | executed time | status | datetime
- implement filter error, custom and list error code, 0: sukses, 99: error
- implement this response as a base json
// Success Response
    { "success": true,
    "message": "Item retrieved successfully",
    "data": { ... },
    "meta": { "page": 1, "limit": 10, "total": 42 }
    }
// Error Response
    { "success": false,
    "message": "SKU already exists",
    "errors": [{ "field": "sku", "reason": "Duplicate entry" }]
    }
- implement swagger complete with the examples
- add indexing Item.sku, Item.name Location.id, Stock.id, Stock.item_id, Categories.id
- buatkan seeder 1000 data auto run
- implement docker-compose.yml
- implement simple FE react19 ts,  tailwind, React Router v6, TanStack Query, build composition and components, 2 pages:
page 1: Tabel/grid yang menampilkan daftar item gudang (SKU, nama, kategori, satuan).
• Fitur search by nama/SKU (debounced input, minimal 300ms delay).
• Filter by kategori (dropdown).
• Pagination (client-side minimal; server-side jika API mendukung).
• Tombol untuk membuka halaman Create/Edit item.

page 2:
Form Tambah / Edit Item
• Form dengan validasi client-side (field wajib: SKU, nama, kategori, unit).
• Submit form memanggil API create atau update.
• Tampilkan loading state saat request berlangsung.
• Tampilkan pesan sukses/error dari API response.
- structure project:
Struktur Repository
Repository harus mengikuti struktur berikut:
your-repo/
├── backend/
│ ├── cmd/ # entrypoint
│ ├── internal/ # handler, service, repo, model
│ ├── migrations/ # SQL migration files
│ ├── docs/ # swagger / openapi
│ ├── Dockerfile
│ └── .env.example
├── frontend/
│ ├── src/
│ │ ├── components/
│ │ ├── hooks/
│ │ ├── pages/
│ │ ├── types/
│ │ └── services/ # API call functions
│ └── package.json
├── docker-compose.yml
└── README.md
- write and update abt the project in README.md, also write all my requested prompt and ai model in README.MD
````

### Prompt 2 — Swagger parser error (verbatim)

````text
fix this error swagger doc "

Errors
Hide

Parser error on line 170
bad indentation of a mapping entry

Unable to render this definition
The provided definition does not specify a valid version field.

Please indicate a valid Swagger or OpenAPI version field. Supported version fields are swagger: "2.0" and those that match openapi: 3.x.y (for example, openapi: 3.1.0).

"
````

*Resolution:* line 170 of `backend/docs/openapi.yaml` held an unquoted scalar
containing a colon-space (`description: Search by name or SKU (alias: \`q\`)`),
which YAML parsed as a nested mapping entry. Because the document then failed to
parse at all, Swagger UI could not read `openapi: 3.0.3` either — hence the
second, misleading message. The string was quoted, the whole file was swept for
the same pattern and re-validated with a YAML parser (paths, operations and
every `$ref`).

### Prompt 3 — menu, master data pages and item location (verbatim)

````text
adjust header menu jadi daftar item | Daftar category | daftar warehouse, tampilan nya items -> list view | btn add item (add field location)
Category -> list view | btn add category
Warehouse -> list view | btn add location. Update readme.md bagian prompt
````

*Resolution:* header menu changed to **Daftar Item | Daftar Category | Daftar
Warehouse**, each with a list view and an add button
(`/items` + `/items/new`, `/categories` + `/categories/new`,
`/warehouses` + `/warehouses/new`). Two endpoints were added to support the new
forms — `POST /api/v1/categories` (409 / code 1011 on a duplicate name) and
`POST /api/v1/locations` (409 / code 1012 on a duplicate code) — and the item
create form gained an optional **Lokasi gudang** field.

*Assumption made:* a location on its own cannot express *how much* is placed
there, so an optional **Qty awal** (`initial_qty`, default `0`) accompanies it.
`POST /items` writes the item and its stock row in one transaction;
`initial_qty: 0` just reserves the slot. The placement is not part of
`PUT /items/:id` — stock keeps moving only through `POST /stock/receive`.

### How each requirement was met

| Requirement | Where |
| --- | --- |
| PostgreSQL entities (Item, Categories, Location, Stock) | `backend/migrations/000001_init_schema.up.sql`, `internal/model` |
| 7 required endpoints | `internal/router/router.go` (+ categories, locations, logs, error-codes, health) |
| SKU unique → 409 | `internal/service/item_service.go`, `ux_items_sku_active` |
| Non-negative stock validated in the **service** | `internal/service/stock_service.go` (validate callback inside the locked transaction) |
| Soft delete | `internal/repository/item_repository.go` (`SoftDelete`), `deleted_at` filters |
| Stock mutation transaction log (console/file) | `internal/logger/stock.go` |
| BE 3069 / FE 4069 | `docker-compose.yml`, `backend/.env.example`, `frontend/vite.config.ts` |
| Realtime request logging to Redis/DB | `internal/middleware/request_logger.go`, `internal/logstore` |
| Error filter + code list (0 / 99 / …) | `internal/errs`, `internal/response`, `GET /api/v1/error-codes` |
| Base JSON response | `internal/response/response.go` |
| Swagger with examples | `backend/docs/openapi.yaml` + `/swagger/index.html` |
| Indexes | migration file, see [index table](#database-schema--indexes) |
| Seeder 1 000 rows, auto-run | `internal/seeder/seeder.go` |
| docker-compose | `docker-compose.yml` (postgres, redis, backend, frontend) |
| React 19 + TS + Tailwind + Router v6 + TanStack Query, 2 pages | `frontend/src` |
| Repository structure | see [structure](#repository-structure) |
| Header menu: Daftar Item / Category / Warehouse | `frontend/src/components/layout/AppShell.tsx`, `src/App.tsx` |
| Item list view + add item button (with location field) | `pages/ItemListPage.tsx`, `components/items/ItemForm.tsx` |
| Category list view + add category button | `pages/CategoryListPage.tsx`, `pages/CategoryFormPage.tsx`, `POST /api/v1/categories` |
| Warehouse list view + add location button | `pages/WarehouseListPage.tsx`, `pages/WarehouseFormPage.tsx`, `POST /api/v1/locations` |
