# luohua-oj

`luohua-oj` is a local-first online judge workspace with:

- `apps/api`: Go API for auth, problems, contests, submissions, admin actions
- `apps/judge-worker`: Go worker for queue-driven judging and rejudge tasks
- `apps/web`: user-facing React app
- `apps/admin-web`: admin React app
- `packages/shared`: shared schemas and route/types contracts

## Quick Start

1. Install dependencies:

```bash
pnpm install
```

2. Start local PostgreSQL and Redis:

```bash
docker compose -f infra/compose.local.yml up -d postgres redis
```

3. Apply database migrations:

```bash
goose -dir packages/database/migrations postgres "$DATABASE_URL" up
```

4. Start the full local stack:

```bash
pnpm dev:stack
```

5. Verify local services:

```bash
pnpm dev:check
```

Default local URLs:

- web: `http://localhost:5173`
- admin web: `http://localhost:5174`
- api: `http://localhost:8080`

## Environment

Base backend variables live in [`.env.example`](/Volumes/新加卷/project/luohua-oj/.env.example).
Web variables live in [`infra/env/web.env.example`](/Volumes/新加卷/project/luohua-oj/infra/env/web.env.example).
Admin-web variables live in [`infra/env/admin-web.env.example`](/Volumes/新加卷/project/luohua-oj/infra/env/admin-web.env.example).

Key variables:

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis address for the judge queue
- `SOURCE_ROOT`: local object storage root for submission sources and problem samples
- `VITE_API_BASE_URL`: user web API base URL
- `VITE_ADMIN_BASE_URL`: user web link target for the admin app
- `VITE_SUBMISSIONS_USERNAME`: optional username used by the submissions page remote history query
- `VITE_DEMO_MODE`: when `true`, the user web can fall back to local seed/demo data

## Demo Mode

`VITE_DEMO_MODE` defaults to `false`.

When it is `false`, `apps/web` does not silently fall back to local seed data for:

- problems
- contests
- contest makeup
- submissions

Those routes surface the real API error state instead.

Only set `VITE_DEMO_MODE=true` when you explicitly want a local demo without a complete backend.

## Verification

Run the aligned quality gate:

```bash
pnpm ci:check
```

This includes:

- `@oj/shared` tests
- `@oj/web` tests and build
- `@oj/admin-web` tests and build
- `go test ./apps/api/... ./apps/judge-worker/...`
