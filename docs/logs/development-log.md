# Development Log

## 2026-05-15

### Scope

- continued the `submissions` frontend split into route, feature, and components
- added structured frontend submission logging
- added API and worker local file logging

### Decisions

- local runtime log directory is `tmp/logs/`
- API file name is `api.log`
- worker file name is `judge-worker.log`
- frontend logs stay local in browser console for now
- no submission source text is allowed in structured logs

### Event Naming

- frontend: `submissions.*`
- API: `submission.*`
- worker: `judge.task.*`

### Known Limits

- frontend still uses pathname routing instead of a router library
- frontend logs are not shipped to a backend collector yet
- file rotation is not implemented in this phase

## 2026-05-16

### Scope

- tightened PostgreSQL integrity constraints for `submissions` and `submission_results`
- added repository integration coverage for DB constraints and `ListSubmissionsByUsername`
- expanded user submissions API to pagination shape with joined `problem.id/slug/title`
- aligned submissions recent list and detail page with richer problem metadata
- rebranded the frontend shell from `oj3` to `luooj`
- added a lightweight Chinese-first language toggle
- redesigned home, problems, contests, and contest workspace pages toward a more standard product UI
- completed two density-tightening passes and a responsive layout pass for medium-width screens

### Decisions

- new DB constraints are applied via additive migration instead of rewriting the initial schema
- `submission_results` now enforces uniqueness on `(submission_id, test_case_id)`
- `GET /users/{username}/submissions` now returns only the paginated object shape:
  - `{ items, total, page, pageSize }`
- joined submission problem summaries stay intentionally small:
  - `{ id, slug, title }`
- frontend i18n remains lightweight and local to the app shell instead of adopting a full translation framework
- responsive layout should prefer multi-column shells from `md`/`lg` upward to avoid low-information cards occupying full rows

### Validation

- PostgreSQL smoke test passed through real repository writes
- API submission package tests passed after constraint and pagination changes
- full `go test ./apps/api/...` passed during the submissions backend work
- `pnpm test:web` passed after the UI redesign, density pass, and responsive pass
- `pnpm build:web` passed after the same frontend changes
- in-browser checks confirmed `/`, `/problems`, `/contests`, and `/contests/spring-open` render with the new layout once the local dev server is running

### Known Limits

- the local fallback submission history still lacks rich `problem` summaries unless data is fetched from the API
- the app still uses mock contest data and does not yet consume a real contests backend
- browser validation depended on the local Vite dev server staying alive; after Codex restarts the server must be started again with `pnpm dev:web`
- there is still no shared density token layer for `button`, `badge`, `summary card`, and `side card`; pages currently reuse class patterns manually

## 2026-05-24

### Scope

- consolidated the auth/admin/judge/submission foundation work onto reviewable branch history
- created and updated draft PR `#1` from `codex/auth-admin-submission-foundation-split` to `main`
- finished the user/admin cross-entry navigation and problem workspace rollout
- fixed duplicate page actions in the web `AppShell`
- split the web app by route-level lazy loading to remove the oversized main chunk warning

### Decisions

- keep the original integrated feature scope, but rebuild history into smaller commits instead of rewriting the feature set again
- use `codex/auth-admin-submission-foundation-split` as the new review baseline
- keep the PR in draft because it still spans schema, backend, worker, and frontend layers together
- treat the current problem-page `compatUserId` path as an explicit temporary compatibility fallback
- prefer route-level `lazy()` loading before adding custom Vite `manualChunks`

### Validation

- `pnpm test` passed on the rebuilt split branch
- `pnpm --filter @oj/web build` passed after the route lazy-loading change
- `pnpm --filter @oj/admin-web build` passed during PR self-check
- `pnpm --filter @oj/web test` passed after the `AppShell` duplicate-action fix
- draft PR created and updated:
  - [PR #1](https://github.com/rippleflower/luohua-oj/pull/1)

### Current Review Baseline

- branch:
  - `codex/auth-admin-submission-foundation-split`
- latest commits:
  - `35c1438` `perf(web): lazy-load route modules`
  - `8542015` `fix(web): avoid duplicate app shell actions`
  - `2043db0` `chore: ignore admin web tsbuildinfo`

### Known Limits

- the draft PR is still large in scope even after history cleanup; reviewer load is lower, but integration risk remains cross-layer
- `apps/web` no longer triggers the `>500 kB` warning, but there is still no deliberate chunk-group strategy beyond route lazy loading
- the problem workspace still exposes a compatibility user ID fallback because the user identity chain is not fully normalized yet
- the next highest-risk area is not the shell UX anymore; it is backend/admin write-path completeness and PR review readiness
