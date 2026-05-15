# Online Judge Tech Stack Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Design and implement a production-ready Online Judge platform that supports problem management, submissions, sandboxed judging, rankings, and an ergonomic web UI.

**Architecture:** Use a modular monolith for the first release, with clear package boundaries for auth, problems, submissions, judging, contests, and administration. Run code execution in isolated judge workers through a queue so the web API remains responsive and the sandbox can scale separately. Keep the database schema normalized and event-friendly so the system can later split into services without rewriting the domain model.

**Tech Stack:** TypeScript, Next.js, NestJS, PostgreSQL, Prisma, Redis, BullMQ, Docker, nsjail or isolate, MinIO/S3, Playwright, Vitest, Jest, ESLint, Prettier, GitHub Actions.

---

## 1. Product Scope

The first version should support:

- User registration, login, session management, and role-based access control.
- Problem browsing with tags, difficulty, statement, examples, constraints, and hidden/public tests.
- Code submission in selected languages.
- Asynchronous judging with statuses: `PENDING`, `RUNNING`, `ACCEPTED`, `WRONG_ANSWER`, `TIME_LIMIT_EXCEEDED`, `MEMORY_LIMIT_EXCEEDED`, `RUNTIME_ERROR`, `COMPILE_ERROR`, `SYSTEM_ERROR`.
- Submission detail pages with compiler output, test summary, runtime, and memory.
- Admin problem management with test case upload.
- Basic contest model with start/end time, problem set, frozen scoreboard option, and ranking.
- Audit-friendly logs for judge execution.

Out of scope for the first version:

- Distributed microservices.
- Real-time collaborative editing.
- Full plagiarism detection.
- Custom interactive problems.
- Multi-tenant school/team isolation.

## 2. Recommended Repository Layout

```text
oj3.0/
  apps/
    web/                         # Next.js frontend
    api/                         # NestJS HTTP API
    judge-worker/                # Queue consumer and sandbox runner
  packages/
    database/                    # Prisma schema, migrations, seed scripts
    shared/                      # Shared DTOs, validators, enums
    eslint-config/               # Shared lint config
    tsconfig/                    # Shared TS config
  infra/
    docker/
      api.Dockerfile
      web.Dockerfile
      judge-worker.Dockerfile
    compose.local.yml
    nginx.local.conf
  docs/
    architecture/
      online-judge-tech-stack.md
      judge-sandbox.md
      data-model.md
    plans/
      2026-05-11-online-judge-tech-stack.md
  tests/
    e2e/
```

## 3. Technology Stack Decisions

### Frontend

- **Framework:** Next.js with App Router.
- **Language:** TypeScript.
- **UI:** Tailwind CSS plus a small component layer using Radix UI primitives.
- **Code editor:** Monaco Editor.
- **Data fetching:** TanStack Query for client-side server state.
- **Forms:** React Hook Form plus Zod validation.
- **Testing:** Vitest for component logic, Playwright for user flows.

Rationale: Next.js is suitable for SEO-visible problem pages and fast admin/product iteration. Monaco gives users a familiar coding experience. TanStack Query keeps submission polling and cache invalidation explicit.

### Backend API

- **Framework:** NestJS.
- **Language:** TypeScript.
- **Validation:** Zod or class-validator; choose one and use it consistently. Prefer Zod if sharing schemas with frontend.
- **ORM:** Prisma.
- **Database:** PostgreSQL 16+.
- **Cache/queue:** Redis plus BullMQ.
- **Auth:** Email/password with Argon2 hashing, HTTP-only session cookies, CSRF protection for browser writes.
- **Authorization:** Role-based access control with `USER`, `PROBLEM_SETTER`, `ADMIN`.
- **Testing:** Jest for unit/integration tests, Testcontainers where practical.

Rationale: NestJS gives a clear module structure and dependency injection without forcing microservices. Prisma is productive for schema-first development and readable migrations.

### Judge Worker

- **Runtime:** Node.js TypeScript worker using BullMQ.
- **Sandbox:** Prefer `nsjail` on Linux. Use `isolate` if the deployment target is based on traditional competitive-programming judge hosts.
- **Container boundary:** Run judge workers in dedicated Docker containers with restricted privileges. The sandbox still runs inside the worker; Docker is not the only isolation layer.
- **Language support V1:** C++17, C++20, Java 17, Python 3.11.
- **Resource control:** Per-test CPU time, wall time, memory limit, output limit, process count, and file-system restrictions.
- **Artifacts:** Store source code, compile output, and detailed judge logs in database/object storage according to retention policy.

Rationale: Executing untrusted code is the highest-risk part of the system. The worker must be separately deployable, locked down, observable, and easy to disable without taking the product offline.

### Storage

- **Relational data:** PostgreSQL.
- **Object storage:** MinIO locally, S3-compatible storage in production.
- **Redis:** Queue state, rate limits, and short-lived cache only. Do not make Redis the source of truth.

### Observability

- **Logs:** Pino JSON logs from API and worker.
- **Metrics:** Prometheus-compatible metrics endpoint.
- **Tracing:** OpenTelemetry from API request to queue job to judge result.
- **Error tracking:** Sentry or an equivalent self-hosted service.

### Deployment

- **Local development:** Docker Compose for PostgreSQL, Redis, MinIO, API, web, and judge worker.
- **Production V1:** One web service, one API service, a scalable judge-worker pool, managed PostgreSQL, managed Redis, S3-compatible object storage.
- **CI:** GitHub Actions running lint, typecheck, unit tests, migration checks, and Playwright smoke tests.

## 4. Core Data Model

Use these entities as the first schema baseline:

```prisma
model User {
  id           String   @id @default(cuid())
  email        String   @unique
  username     String   @unique
  passwordHash String
  role         Role     @default(USER)
  createdAt    DateTime @default(now())
  updatedAt    DateTime @updatedAt

  submissions  Submission[]
}

enum Role {
  USER
  PROBLEM_SETTER
  ADMIN
}

model Problem {
  id             String     @id @default(cuid())
  slug           String     @unique
  title          String
  difficulty     Difficulty
  statementMd    String
  inputMd        String
  outputMd       String
  constraintsMd  String
  timeLimitMs    Int
  memoryLimitKb  Int
  isPublished    Boolean    @default(false)
  createdAt      DateTime   @default(now())
  updatedAt      DateTime   @updatedAt

  testCases      TestCase[]
  submissions    Submission[]
}

enum Difficulty {
  EASY
  MEDIUM
  HARD
}

model TestCase {
  id           String   @id @default(cuid())
  problemId    String
  inputObject  String
  outputObject String
  isSample     Boolean  @default(false)
  weight       Int      @default(1)
  createdAt    DateTime @default(now())

  problem      Problem  @relation(fields: [problemId], references: [id])
}

model Submission {
  id             String           @id @default(cuid())
  userId         String
  problemId      String
  language       Language
  sourceObject   String
  status         SubmissionStatus @default(PENDING)
  score          Int              @default(0)
  compileOutput  String?
  maxTimeMs      Int?
  maxMemoryKb    Int?
  createdAt      DateTime         @default(now())
  judgedAt       DateTime?

  user           User             @relation(fields: [userId], references: [id])
  problem        Problem          @relation(fields: [problemId], references: [id])
  results        SubmissionResult[]
}

model SubmissionResult {
  id            String           @id @default(cuid())
  submissionId  String
  testCaseId    String
  status        SubmissionStatus
  timeMs        Int?
  memoryKb      Int?
  outputSnippet String?
  errorSnippet  String?

  submission    Submission       @relation(fields: [submissionId], references: [id])
}

enum Language {
  CPP17
  CPP20
  JAVA17
  PYTHON311
}

enum SubmissionStatus {
  PENDING
  RUNNING
  ACCEPTED
  WRONG_ANSWER
  TIME_LIMIT_EXCEEDED
  MEMORY_LIMIT_EXCEEDED
  RUNTIME_ERROR
  COMPILE_ERROR
  SYSTEM_ERROR
}
```

## 5. Security Requirements

- Run all submitted code as a non-root user.
- Disable network access inside sandboxed executions.
- Mount a temporary working directory per execution.
- Enforce CPU, wall-clock, memory, file size, process count, and output size limits.
- Store source code and test cases outside the web root.
- Rate-limit login and submission endpoints.
- Keep hidden test case inputs/outputs inaccessible from frontend and normal users.
- Validate every API input at the boundary.
- Never interpolate user input into shell commands. Use argument arrays with `spawn`.
- Record enough execution metadata to debug judge failures without exposing hidden outputs to contestants.

## 6. Implementation Tasks

### Task 1: Scaffold Monorepo

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `turbo.json`
- Create: `packages/tsconfig/base.json`
- Create: `packages/eslint-config/base.js`
- Create: `.gitignore`

**Step 1: Write the workspace files**

```json
{
  "name": "oj3",
  "private": true,
  "packageManager": "pnpm@9.12.0",
  "scripts": {
    "dev": "turbo dev",
    "build": "turbo build",
    "lint": "turbo lint",
    "typecheck": "turbo typecheck",
    "test": "turbo test"
  },
  "devDependencies": {
    "turbo": "^2.1.0",
    "typescript": "^5.6.0",
    "eslint": "^9.12.0",
    "prettier": "^3.3.0"
  }
}
```

**Step 2: Run install**

Run: `pnpm install`

Expected: lockfile is created and workspace dependencies install successfully.

**Step 3: Commit**

```bash
git add package.json pnpm-workspace.yaml turbo.json packages/tsconfig/base.json packages/eslint-config/base.js .gitignore pnpm-lock.yaml
git commit -m "chore: scaffold monorepo"
```

### Task 2: Add Database Package

**Files:**
- Create: `packages/database/package.json`
- Create: `packages/database/prisma/schema.prisma`
- Create: `packages/database/src/client.ts`
- Create: `packages/database/src/index.ts`
- Test: `packages/database/prisma/seed.ts`

**Step 1: Write Prisma schema**

Use the schema from section 4 and add provider config:

```prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}
```

**Step 2: Generate migration**

Run: `pnpm --filter @oj/database prisma migrate dev --name init`

Expected: migration is created and Prisma Client is generated.

**Step 3: Add seed data**

Create one admin user and one sample problem with two sample test cases.

**Step 4: Verify seed**

Run: `pnpm --filter @oj/database prisma db seed`

Expected: seed completes without unique constraint errors.

**Step 5: Commit**

```bash
git add packages/database
git commit -m "feat: add initial database schema"
```

### Task 3: Add Shared Contracts

**Files:**
- Create: `packages/shared/package.json`
- Create: `packages/shared/src/enums.ts`
- Create: `packages/shared/src/problem.schemas.ts`
- Create: `packages/shared/src/submission.schemas.ts`
- Test: `packages/shared/src/submission.schemas.test.ts`

**Step 1: Write failing schema test**

```ts
import { describe, expect, it } from "vitest";
import { createSubmissionSchema } from "./submission.schemas";

describe("createSubmissionSchema", () => {
  it("rejects unsupported languages", () => {
    expect(() =>
      createSubmissionSchema.parse({
        problemId: "clx123",
        language: "RUBY",
        source: "puts 1",
      }),
    ).toThrow();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @oj/shared test`

Expected: FAIL because `createSubmissionSchema` does not exist.

**Step 3: Implement schema**

```ts
import { z } from "zod";

export const languageSchema = z.enum(["CPP17", "CPP20", "JAVA17", "PYTHON311"]);

export const createSubmissionSchema = z.object({
  problemId: z.string().min(1),
  language: languageSchema,
  source: z.string().min(1).max(100_000),
});
```

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @oj/shared test`

Expected: PASS.

**Step 5: Commit**

```bash
git add packages/shared
git commit -m "feat: add shared validation contracts"
```

### Task 4: Build API Foundation

**Files:**
- Create: `apps/api/package.json`
- Create: `apps/api/src/main.ts`
- Create: `apps/api/src/app.module.ts`
- Create: `apps/api/src/health/health.controller.ts`
- Test: `apps/api/src/health/health.controller.spec.ts`

**Step 1: Write failing health test**

```ts
import { Test } from "@nestjs/testing";
import { HealthController } from "./health.controller";

describe("HealthController", () => {
  it("returns ok", () => {
    const controller = new HealthController();
    expect(controller.getHealth()).toEqual({ status: "ok" });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @oj/api test health.controller`

Expected: FAIL because the controller does not exist.

**Step 3: Implement controller**

```ts
import { Controller, Get } from "@nestjs/common";

@Controller("health")
export class HealthController {
  @Get()
  getHealth() {
    return { status: "ok" };
  }
}
```

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @oj/api test health.controller`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/api
git commit -m "feat: add api health endpoint"
```

### Task 5: Implement Submission Queue API

**Files:**
- Create: `apps/api/src/submissions/submissions.module.ts`
- Create: `apps/api/src/submissions/submissions.controller.ts`
- Create: `apps/api/src/submissions/submissions.service.ts`
- Test: `apps/api/src/submissions/submissions.service.spec.ts`

**Step 1: Write failing service test**

```ts
it("creates a pending submission and enqueues a judge job", async () => {
  const prisma = createMockPrisma();
  const queue = createMockQueue();
  const service = new SubmissionsService(prisma, queue);

  await service.create({
    userId: "user_1",
    problemId: "problem_1",
    language: "CPP17",
    source: "int main(){return 0;}",
  });

  expect(prisma.submission.create).toHaveBeenCalledWith(
    expect.objectContaining({
      data: expect.objectContaining({ status: "PENDING" }),
    }),
  );
  expect(queue.add).toHaveBeenCalledWith("judge-submission", expect.any(Object));
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @oj/api test submissions.service`

Expected: FAIL because `SubmissionsService` does not exist.

**Step 3: Implement minimal service**

Create the submission, store source in object storage or a local abstraction, then enqueue `judge-submission` with `submissionId`.

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @oj/api test submissions.service`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/api/src/submissions
git commit -m "feat: enqueue submissions for judging"
```

### Task 6: Implement Judge Worker Skeleton

**Files:**
- Create: `apps/judge-worker/package.json`
- Create: `apps/judge-worker/src/main.ts`
- Create: `apps/judge-worker/src/judge/judge.processor.ts`
- Create: `apps/judge-worker/src/sandbox/sandbox-runner.ts`
- Test: `apps/judge-worker/src/sandbox/sandbox-runner.spec.ts`

**Step 1: Write failing sandbox command test**

```ts
it("builds nsjail args without shell interpolation", () => {
  const runner = new SandboxRunner();
  const args = runner.buildArgs({
    command: "/usr/bin/python3",
    commandArgs: ["main.py"],
    timeLimitMs: 1000,
    memoryLimitKb: 262144,
    workdir: "/tmp/oj-run-1",
  });

  expect(args).toContain("--disable_clone_newnet");
  expect(args).toContain("--time_limit");
  expect(args).toContain("1");
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @oj/judge-worker test sandbox-runner`

Expected: FAIL because `SandboxRunner` does not exist.

**Step 3: Implement argument builder**

Use `spawn(binary, args, { shell: false })`. Do not concatenate a command string.

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @oj/judge-worker test sandbox-runner`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/judge-worker
git commit -m "feat: add judge worker sandbox skeleton"
```

### Task 7: Build Web App Foundation

**Files:**
- Create: `apps/web/package.json`
- Create: `apps/web/app/layout.tsx`
- Create: `apps/web/app/page.tsx`
- Create: `apps/web/app/problems/page.tsx`
- Create: `apps/web/components/problem-list.tsx`
- Test: `apps/web/tests/problems.spec.ts`

**Step 1: Write failing Playwright smoke test**

```ts
import { expect, test } from "@playwright/test";

test("problem list renders", async ({ page }) => {
  await page.goto("/problems");
  await expect(page.getByRole("heading", { name: "Problems" })).toBeVisible();
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @oj/web test:e2e`

Expected: FAIL because `/problems` does not exist.

**Step 3: Implement problems page**

Render a dense table with title, difficulty, tags, accepted rate, and action link. Avoid a marketing landing page; the first screen should be the usable problem browser.

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @oj/web test:e2e`

Expected: PASS.

**Step 5: Commit**

```bash
git add apps/web
git commit -m "feat: add web problem browser"
```

### Task 8: Add Local Infrastructure

**Files:**
- Create: `infra/compose.local.yml`
- Create: `infra/docker/api.Dockerfile`
- Create: `infra/docker/web.Dockerfile`
- Create: `infra/docker/judge-worker.Dockerfile`
- Create: `.env.example`

**Step 1: Write Compose services**

Include:

- `postgres`
- `redis`
- `minio`
- `api`
- `web`
- `judge-worker`

**Step 2: Boot local dependencies**

Run: `docker compose -f infra/compose.local.yml up postgres redis minio`

Expected: services are healthy.

**Step 3: Run migrations**

Run: `pnpm --filter @oj/database prisma migrate dev`

Expected: migrations apply to local PostgreSQL.

**Step 4: Boot app**

Run: `pnpm dev`

Expected: web, API, and worker start.

**Step 5: Commit**

```bash
git add infra .env.example
git commit -m "chore: add local infrastructure"
```

### Task 9: Add CI

**Files:**
- Create: `.github/workflows/ci.yml`

**Step 1: Write CI workflow**

```yaml
name: CI

on:
  pull_request:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9.12.0
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint
      - run: pnpm typecheck
      - run: pnpm test
      - run: pnpm build
```

**Step 2: Validate locally**

Run: `pnpm lint && pnpm typecheck && pnpm test && pnpm build`

Expected: all commands pass.

**Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add validation workflow"
```

## 7. API Surface V1

```text
GET    /health
POST   /auth/register
POST   /auth/login
POST   /auth/logout
GET    /me
GET    /problems
POST   /problems                 # problem setter/admin
GET    /problems/:slug
PATCH  /problems/:id             # problem setter/admin
POST   /problems/:id/test-cases  # problem setter/admin
POST   /submissions
GET    /submissions/:id
GET    /users/:username/submissions
GET    /contests
POST   /contests                 # admin
GET    /contests/:id
GET    /contests/:id/standings
```

## 8. Verification Checklist

- `pnpm lint` passes.
- `pnpm typecheck` passes.
- `pnpm test` passes.
- `pnpm build` passes.
- `docker compose -f infra/compose.local.yml up` boots local services.
- A user can register, log in, open `/problems`, submit code, and see the result update asynchronously.
- Hidden test cases are never returned by public APIs.
- Judge worker cannot access the network from submitted code.
- Time limit, memory limit, compile error, runtime error, and wrong answer are each covered by tests.

## 9. Key Engineering Rules

- Keep the first release modular, not distributed.
- Treat judge execution as hostile input.
- Make queue jobs idempotent; retrying a judge job must not corrupt final results.
- Keep public DTOs separate from database entities.
- Prefer explicit status transitions over scattered updates.
- Store large source/test artifacts in object storage, not directly in PostgreSQL.
- Do not build contest features until normal problem submission is stable.

## 10. Execution Handoff

Plan complete and saved to `docs/plans/2026-05-11-online-judge-tech-stack.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
