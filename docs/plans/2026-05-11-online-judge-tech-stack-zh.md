# Online Judge 技术栈实施计划

> **给 Claude：** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 设计并实现一个可投入生产使用的 Online Judge 平台，支持题目管理、代码提交、沙箱判题、排行榜和易用的 Web 界面。

**架构：** 第一版采用模块化单体架构，按认证、题目、提交、判题、比赛、后台管理拆分清晰的业务边界。代码执行通过队列交给独立 Judge Worker 异步处理，保证 Web API 响应稳定，并让判题沙箱可以单独扩容。数据库模型保持规范化，并预留事件化扩展空间，后续需要拆服务时不用重写核心领域模型。

**技术栈：** React、TypeScript、Vite、Tailwind CSS、Radix UI、Monaco Editor、TanStack Query、React Hook Form、Zod、Go、chi、PostgreSQL、pgx、sqlc、Redis、Asynq、Docker、nsjail 或 isolate、MinIO/S3、Playwright、Vitest、Go testing、Testcontainers、ESLint、Prettier、GitHub Actions。

---

## 1. 产品范围

第一版需要支持：

- 用户注册、登录、会话管理和基于角色的权限控制。
- 题目浏览，包含标签、难度、题面、样例、约束、公开测试和隐藏测试。
- 支持多种编程语言提交代码。
- 异步判题，状态包括：`PENDING`、`RUNNING`、`ACCEPTED`、`WRONG_ANSWER`、`TIME_LIMIT_EXCEEDED`、`MEMORY_LIMIT_EXCEEDED`、`RUNTIME_ERROR`、`COMPILE_ERROR`、`SYSTEM_ERROR`。
- 提交详情页，展示编译输出、测试点摘要、运行时间、内存占用。
- 管理端题目维护和测试数据上传。
- 基础比赛模型，支持开始时间、结束时间、题目集合、封榜选项和排名。
- 可审计的判题执行日志。

第一版不做：

- 分布式微服务架构。
- 实时协同编辑。
- 完整查重系统。
- 交互题。
- 多租户学校/团队隔离。

## 2. 推荐目录结构

```text
oj3.0/
  apps/
    web/                         # React + Vite 前端
    api/                         # Go HTTP API
    judge-worker/                # Go 队列消费者和沙箱执行器
  packages/
    database/                    # SQL migration、sqlc 查询、种子数据
    shared/                      # 前端共享类型、Zod schema、枚举
    eslint-config/               # 共享 lint 配置
    tsconfig/                    # 共享 TypeScript 配置
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
      2026-05-11-online-judge-tech-stack-zh.md
  tests/
    e2e/
```

## 3. 技术选型

### 前端

- **框架：** React + Vite。
- **语言：** TypeScript。
- **UI：** Tailwind CSS，加一层基于 Radix UI primitives 的轻量组件库。
- **代码编辑器：** Monaco Editor。
- **数据请求：** TanStack Query。
- **表单：** React Hook Form + Zod。
- **测试：** Vitest 做组件和逻辑测试，Playwright 做端到端流程测试。

选择理由：React + Vite 启动快、构建链路简单，适合后台系统和 OJ 这种以交互为主的应用；Monaco Editor 是用户熟悉的在线代码编辑体验；TanStack Query 可以清晰处理提交状态轮询、缓存失效和错误重试。

### 后端 API

- **语言：** Go。
- **HTTP 路由：** chi，保持轻量、显式、易测试。
- **校验：** `go-playground/validator` 做后端请求校验；Zod 只放在前端表单和客户端 DTO 校验层。
- **数据库访问：** pgx + sqlc。用 SQL migration 管 schema，用 sqlc 从 SQL 生成类型安全 Go 代码。
- **数据库：** PostgreSQL 16+。
- **缓存和队列：** Redis + Asynq。
- **认证：** 邮箱密码登录，Argon2 哈希，HTTP-only session cookie，浏览器写操作启用 CSRF 防护。
- **授权：** 基于角色控制，角色包括 `USER`、`PROBLEM_SETTER`、`ADMIN`。
- **测试：** Go 标准库 `testing`、`testify`，必要时使用 Testcontainers。

选择理由：Go 后端适合高并发 API 和判题调度这类 IO 密集场景，部署产物简单，运行成本低。pgx + sqlc 保留 SQL 的可控性，同时提供编译期类型检查，避免在核心业务里引入过重 ORM。

### 判题 Worker

- **运行时：** Go Worker，消费 Asynq 队列。
- **沙箱：** Linux 环境优先使用 `nsjail`；如果部署在传统 OJ 判题机环境，可考虑 `isolate`。
- **容器边界：** Judge Worker 独立跑在受限 Docker 容器内。Docker 不是唯一隔离层，容器内仍然需要沙箱。
- **第一版语言：** C++17、C++20、Java 17、Python 3.11。
- **资源限制：** 每个测试点限制 CPU 时间、墙钟时间、内存、输出大小、进程数和文件系统访问。
- **产物：** 源码、编译输出、详细判题日志按保留策略存入数据库或对象存储。

选择理由：执行用户提交的代码是系统最高风险点。Worker 必须可单独部署、可限制权限、可观测，并且可以在不影响主站的情况下下线。

### 存储

- **关系型数据：** PostgreSQL。
- **对象存储：** 本地 MinIO，生产环境使用 S3 兼容存储。
- **Redis：** 只用于队列状态、限流和短期缓存，不作为事实数据源。

### 可观测性

- **日志：** API 和 Worker 使用 `slog` 或 `zap` 输出 JSON 日志。
- **指标：** 暴露 Prometheus 兼容 metrics endpoint。
- **链路追踪：** 使用 OpenTelemetry 串联 API 请求、队列任务和判题结果。
- **错误追踪：** Sentry 或同类自托管服务。

### 部署

- **本地开发：** Docker Compose 启动 PostgreSQL、Redis、MinIO、API、Web、Judge Worker。
- **生产第一版：** 一个 Web 服务、一个 API 服务、可水平扩容的 Judge Worker 池、托管 PostgreSQL、托管 Redis、S3 兼容对象存储。
- **CI：** GitHub Actions 运行前端 lint、前端测试、前端构建、Go 单元测试、migration checks、Playwright smoke tests。

## 4. 核心数据模型

第一版建议使用这些表。迁移文件放在 `packages/database/migrations`，查询 SQL 放在 `packages/database/query`，由 sqlc 生成 Go 类型和查询方法。

```sql
CREATE TYPE user_role AS ENUM ('USER', 'PROBLEM_SETTER', 'ADMIN');
CREATE TYPE problem_difficulty AS ENUM ('EASY', 'MEDIUM', 'HARD');
CREATE TYPE language AS ENUM ('CPP17', 'CPP20', 'JAVA17', 'PYTHON311');
CREATE TYPE submission_status AS ENUM (
  'PENDING',
  'RUNNING',
  'ACCEPTED',
  'WRONG_ANSWER',
  'TIME_LIMIT_EXCEEDED',
  'MEMORY_LIMIT_EXCEEDED',
  'RUNTIME_ERROR',
  'COMPILE_ERROR',
  'SYSTEM_ERROR'
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role user_role NOT NULL DEFAULT 'USER',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE problems (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  difficulty problem_difficulty NOT NULL,
  statement_md TEXT NOT NULL,
  input_md TEXT NOT NULL,
  output_md TEXT NOT NULL,
  constraints_md TEXT NOT NULL,
  time_limit_ms INTEGER NOT NULL,
  memory_limit_kb INTEGER NOT NULL,
  is_published BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE test_cases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
  input_object TEXT NOT NULL,
  output_object TEXT NOT NULL,
  is_sample BOOLEAN NOT NULL DEFAULT false,
  weight INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE submissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  problem_id UUID NOT NULL REFERENCES problems(id),
  language language NOT NULL,
  source_object TEXT NOT NULL,
  status submission_status NOT NULL DEFAULT 'PENDING',
  score INTEGER NOT NULL DEFAULT 0,
  compile_output TEXT,
  max_time_ms INTEGER,
  max_memory_kb INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  judged_at TIMESTAMPTZ
);

CREATE TABLE submission_results (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
  test_case_id UUID NOT NULL REFERENCES test_cases(id),
  status submission_status NOT NULL,
  time_ms INTEGER,
  memory_kb INTEGER,
  output_snippet TEXT,
  error_snippet TEXT
);
```
```

## 5. 安全要求

- 所有用户提交代码必须以非 root 用户运行。
- 沙箱内禁用网络访问。
- 每次执行创建独立临时工作目录。
- 强制限制 CPU、墙钟时间、内存、文件大小、进程数量、输出大小。
- 源码和测试数据不能放在 Web 可直接访问的目录。
- 登录和提交接口必须限流。
- 隐藏测试点输入输出不能从前端或普通用户 API 获取。
- 所有 API 入参必须在边界层校验。
- 禁止把用户输入拼接进 shell 命令，使用 `spawn` 的参数数组。
- 记录足够的执行元数据来排查判题问题，但不能泄露隐藏输出。

## 6. 实施任务

### 任务 1：搭建 Monorepo

**文件：**
- 创建：`package.json`
- 创建：`pnpm-workspace.yaml`
- 创建：`packages/tsconfig/base.json`
- 创建：`go.work`
- 创建：`apps/api/go.mod`
- 创建：`apps/judge-worker/go.mod`
- 创建：`.gitignore`

**步骤 1：编写 workspace 配置**

```json
{
  "name": "oj3",
  "private": true,
  "packageManager": "pnpm@9.12.0",
  "scripts": {
    "dev:web": "pnpm --filter @oj/web dev",
    "build:web": "pnpm --filter @oj/web build",
    "lint:web": "pnpm --filter @oj/web lint",
    "test:web": "pnpm --filter @oj/web test",
    "test:go": "go test ./apps/api/... ./apps/judge-worker/..."
  },
  "devDependencies": {
    "typescript": "^5.6.0",
    "eslint": "^9.12.0",
    "prettier": "^3.3.0"
  }
}
```

**步骤 2：安装依赖**

运行：`pnpm install`

预期：生成 lockfile，前端 workspace 依赖安装成功。

**步骤 3：提交**

```bash
git add package.json pnpm-workspace.yaml packages/tsconfig/base.json go.work apps/api/go.mod apps/judge-worker/go.mod .gitignore pnpm-lock.yaml
git commit -m "chore: scaffold monorepo"
```

### 任务 2：添加数据库迁移和 sqlc

**文件：**
- 创建：`packages/database/migrations/000001_init.sql`
- 创建：`packages/database/query/submissions.sql`
- 创建：`apps/api/internal/db/sqlc.yaml`
- 生成：`apps/api/internal/db/generated`
- 测试：`apps/api/internal/db/db_test.go`

**步骤 1：编写 migration**

使用第 4 节 SQL 作为 `000001_init.sql`，并补充必要索引，例如 `submissions(user_id, created_at DESC)`、`submissions(problem_id, created_at DESC)`。

**步骤 2：编写 sqlc 查询**

```sql
-- name: CreateSubmission :one
INSERT INTO submissions (user_id, problem_id, language, source_object, status)
VALUES ($1, $2, $3, $4, 'PENDING')
RETURNING *;

-- name: GetSubmission :one
SELECT * FROM submissions WHERE id = $1;
```

**步骤 3：生成 Go 代码**

运行：`sqlc generate -f apps/api/internal/db/sqlc.yaml`

预期：`apps/api/internal/db/generated` 下生成类型安全查询代码。

**步骤 4：验证 migration**

运行：`go test ./apps/api/internal/db/...`

预期：Testcontainers 启动 PostgreSQL，migration 成功执行，`CreateSubmission` 测试通过。

**步骤 5：提交**

```bash
git add packages/database apps/api/internal/db
git commit -m "feat: add database migrations and sqlc queries"
```

### 任务 3：添加前端共享契约

**文件：**
- 创建：`packages/shared/package.json`
- 创建：`packages/shared/src/enums.ts`
- 创建：`packages/shared/src/problem.schemas.ts`
- 创建：`packages/shared/src/submission.schemas.ts`
- 测试：`packages/shared/src/submission.schemas.test.ts`

**步骤 1：编写失败测试**

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

**步骤 2：运行测试确认失败**

运行：`pnpm --filter @oj/shared test`

预期：失败，因为 `createSubmissionSchema` 尚不存在。

**步骤 3：实现 schema**

```ts
import { z } from "zod";

export const languageSchema = z.enum(["CPP17", "CPP20", "JAVA17", "PYTHON311"]);

export const createSubmissionSchema = z.object({
  problemId: z.string().min(1),
  language: languageSchema,
  source: z.string().min(1).max(100_000),
});
```

**步骤 4：运行测试确认通过**

运行：`pnpm --filter @oj/shared test`

预期：通过。

**步骤 5：提交**

```bash
git add packages/shared
git commit -m "feat: add frontend validation contracts"
```

### 任务 4：搭建 Go API 基础

**文件：**
- 创建：`apps/api/cmd/api/main.go`
- 创建：`apps/api/internal/http/router.go`
- 创建：`apps/api/internal/http/health_handler.go`
- 测试：`apps/api/internal/http/health_handler_test.go`

**步骤 1：编写健康检查失败测试**

```go
package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apphttp "github.com/example/oj3/apps/api/internal/http"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	apphttp.NewRouter().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}
```

**步骤 2：运行测试确认失败**

运行：`go test ./apps/api/internal/http -run TestHealth`

预期：失败，因为 router 和 handler 不存在。

**步骤 3：实现 handler**

```go
package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", healthHandler)
	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

**步骤 4：运行测试确认通过**

运行：`go test ./apps/api/internal/http -run TestHealth`

预期：通过。

**步骤 5：提交**

```bash
git add apps/api
git commit -m "feat: add api health endpoint"
```

### 任务 5：实现提交队列 API

**文件：**
- 创建：`apps/api/internal/submission/service.go`
- 创建：`apps/api/internal/submission/handler.go`
- 创建：`apps/api/internal/queue/client.go`
- 测试：`apps/api/internal/submission/service_test.go`

**步骤 1：编写失败测试**

```go
func TestCreateSubmissionEnqueuesJudgeJob(t *testing.T) {
	repo := NewMockSubmissionRepo(t)
	queue := NewMockJudgeQueue(t)
	service := submission.NewService(repo, queue)

	repo.EXPECT().CreateSubmission(mock.Anything, mock.Anything).Return(db.Submission{
		ID:     uuid.New(),
		Status: "PENDING",
	}, nil)
	queue.EXPECT().EnqueueJudgeSubmission(mock.Anything, mock.Anything).Return(nil)

	created, err := service.Create(context.Background(), submission.CreateInput{
		UserID: uuid.New(),
		ProblemID: uuid.New(),
		Language: "CPP17",
		Source: "int main(){return 0;}",
	})

	require.NoError(t, err)
	require.Equal(t, "PENDING", created.Status)
}
```

**步骤 2：运行测试确认失败**

运行：`go test ./apps/api/internal/submission -run TestCreateSubmissionEnqueuesJudgeJob`

预期：失败，因为 submission service 不存在。

**步骤 3：实现最小服务**

创建提交记录，将源码写入对象存储或本地存储抽象，然后使用 Asynq 写入 `judge:submission` 任务，任务 payload 包含 `submissionId`。

**步骤 4：运行测试确认通过**

运行：`go test ./apps/api/internal/submission -run TestCreateSubmissionEnqueuesJudgeJob`

预期：通过。

**步骤 5：提交**

```bash
git add apps/api/internal/submission apps/api/internal/queue
git commit -m "feat: enqueue submissions for judging"
```

### 任务 6：实现 Go Judge Worker 骨架

**文件：**
- 创建：`apps/judge-worker/cmd/worker/main.go`
- 创建：`apps/judge-worker/internal/judge/processor.go`
- 创建：`apps/judge-worker/internal/sandbox/runner.go`
- 测试：`apps/judge-worker/internal/sandbox/runner_test.go`

**步骤 1：编写沙箱命令失败测试**

```go
func TestBuildNSJailArgs(t *testing.T) {
	runner := sandbox.Runner{}
	args := runner.BuildArgs(sandbox.RunSpec{
		Command: "/usr/bin/python3",
		Args: []string{"main.py"},
		TimeLimitMs: 1000,
		MemoryLimitKB: 262144,
		Workdir: "/tmp/oj-run-1",
	})

	require.Contains(t, args, "--disable_clone_newnet")
	require.Contains(t, args, "--time_limit")
	require.Contains(t, args, "1")
}
```

**步骤 2：运行测试确认失败**

运行：`go test ./apps/judge-worker/internal/sandbox -run TestBuildNSJailArgs`

预期：失败，因为 `sandbox.Runner` 不存在。

**步骤 3：实现参数构造器**

使用 `exec.CommandContext(binary, args...)`。禁止通过 shell 拼接命令字符串。

**步骤 4：运行测试确认通过**

运行：`go test ./apps/judge-worker/internal/sandbox -run TestBuildNSJailArgs`

预期：通过。

**步骤 5：提交**

```bash
git add apps/judge-worker
git commit -m "feat: add judge worker sandbox skeleton"
```

### 任务 7：搭建 Web 前端

**文件：**
- 创建：`apps/web/package.json`
- 创建：`apps/web/index.html`
- 创建：`apps/web/src/main.tsx`
- 创建：`apps/web/src/app.tsx`
- 创建：`apps/web/src/pages/problems.tsx`
- 创建：`apps/web/src/components/problem-list.tsx`
- 测试：`apps/web/tests/problems.spec.ts`

**步骤 1：编写 Playwright 失败测试**

```ts
import { expect, test } from "@playwright/test";

test("problem list renders", async ({ page }) => {
  await page.goto("/problems");
  await expect(page.getByRole("heading", { name: "Problems" })).toBeVisible();
});
```

**步骤 2：运行测试确认失败**

运行：`pnpm --filter @oj/web test:e2e`

预期：失败，因为 `/problems` 路由不存在。

**步骤 3：实现题目列表页**

渲染高信息密度表格，包含题名、难度、标签、通过率和操作入口。不要做营销式落地页；第一屏就是可用的题目浏览器。

**步骤 4：运行测试确认通过**

运行：`pnpm --filter @oj/web test:e2e`

预期：通过。

**步骤 5：提交**

```bash
git add apps/web
git commit -m "feat: add web problem browser"
```

### 任务 8：添加本地基础设施

**文件：**
- 创建：`infra/compose.local.yml`
- 创建：`infra/docker/api.Dockerfile`
- 创建：`infra/docker/web.Dockerfile`
- 创建：`infra/docker/judge-worker.Dockerfile`
- 创建：`.env.example`

**步骤 1：编写 Compose 服务**

包含：

- `postgres`
- `redis`
- `minio`
- `api`
- `web`
- `judge-worker`

**步骤 2：启动本地依赖**

运行：`docker compose -f infra/compose.local.yml up postgres redis minio`

预期：服务健康。

**步骤 3：运行数据库迁移**

运行：`goose -dir packages/database/migrations postgres "$DATABASE_URL" up`

预期：迁移成功应用到本地 PostgreSQL。

**步骤 4：启动应用**

运行：`pnpm dev:web`，另开终端运行 `go run ./apps/api/cmd/api` 和 `go run ./apps/judge-worker/cmd/worker`

预期：Web、API、Worker 都能启动。

**步骤 5：提交**

```bash
git add infra .env.example
git commit -m "chore: add local infrastructure"
```

### 任务 9：添加 CI

**文件：**
- 创建：`.github/workflows/ci.yml`

**步骤 1：编写 CI workflow**

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
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint:web
      - run: pnpm test:web
      - run: pnpm build:web
      - run: go test ./apps/api/... ./apps/judge-worker/...
```

**步骤 2：本地验证**

运行：`pnpm lint:web && pnpm test:web && pnpm build:web && go test ./apps/api/... ./apps/judge-worker/...`

预期：全部通过。

**步骤 3：提交**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add validation workflow"
```

## 7. API 设计 V1

```text
GET    /health
POST   /auth/register
POST   /auth/login
POST   /auth/logout
GET    /me
GET    /problems
POST   /problems                 # 出题人/管理员
GET    /problems/:slug
PATCH  /problems/:id             # 出题人/管理员
POST   /problems/:id/test-cases  # 出题人/管理员
POST   /submissions
GET    /submissions/:id
GET    /users/:username/submissions
GET    /contests
POST   /contests                 # 管理员
GET    /contests/:id
GET    /contests/:id/standings
```

## 8. 验收清单

- `pnpm lint:web` 通过。
- `pnpm test:web` 通过。
- `pnpm build:web` 通过。
- `go test ./apps/api/... ./apps/judge-worker/...` 通过。
- `docker compose -f infra/compose.local.yml up` 能启动本地服务。
- 用户可以注册、登录、打开 `/problems`、提交代码，并异步看到判题结果。
- 隐藏测试点不会从公开 API 泄露。
- 用户提交代码在沙箱中无法访问网络。
- 时间超限、内存超限、编译错误、运行错误、答案错误都覆盖测试。

## 9. 关键工程原则

- 第一版保持模块化单体，不要过早拆微服务。
- 把判题执行视为恶意输入。
- 队列任务必须幂等；重复执行判题任务不能破坏最终结果。
- 公开 DTO 与数据库实体分离。
- 状态流转要显式，不要在各处随意更新。
- 大源码和测试数据放对象存储，不直接塞进 PostgreSQL。
- 普通题目提交稳定前，不扩展复杂比赛功能。

## 10. 执行交接

计划已保存到 `docs/plans/2026-05-11-online-judge-tech-stack-zh.md`。两个执行选项：

**1. Subagent-Driven（当前会话）** - 每个任务派发一个新 subagent，任务间 review，快速迭代。

**2. Parallel Session（单独会话）** - 在新会话中使用 executing-plans，按检查点批量执行。

请选择执行方式。
