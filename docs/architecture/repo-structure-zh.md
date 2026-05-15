# Online Judge 仓库目录结构建议

## 目标

这个目录结构的目标不是“看起来完整”，而是让前端、后端 API、判题 Worker、数据库、基础设施分层明确，后续迭代时不互相污染。

当前项目已经有一个不错的起点：`apps/web`、`apps/api`、`apps/judge-worker`、`packages/database`、`packages/shared`。下一步建议是在这个基础上细化，而不是推翻重来。

## 设计原则

- 前端、后端、判题端必须物理分开。
- HTTP API 和判题 Worker 分开部署，不能共享一套“万能 service”目录。
- 数据库 schema 和查询语句统一放在 `packages/database`。
- 前端共享类型只服务前端，不强行让 Go 直接复用 TypeScript 类型。
- 目录先为“已有职责”服务，不为“未来可能用到”预留太多空目录。

## 推荐结构

```text
oj3.0/
  apps/
    web/
      index.html
      package.json
      tsconfig.json
      vite.config.ts
      playwright.config.ts
      postcss.config.js
      tailwind.config.ts
      src/
        main.tsx
        app.tsx
        styles.css
        routes/
          problems/
            index.tsx
            detail.tsx
          submissions/
            index.tsx
            detail.tsx
          admin/
            problems.tsx
            test-cases.tsx
        components/
          layout/
            app-shell.tsx
            top-nav.tsx
            side-nav.tsx
          problem/
            problem-list.tsx
            problem-filter.tsx
            problem-detail.tsx
          submission/
            submission-table.tsx
            submission-status-badge.tsx
          editor/
            code-editor.tsx
            language-select.tsx
        features/
          auth/
            api.ts
            hooks.ts
            schema.ts
          problems/
            api.ts
            hooks.ts
            schema.ts
          submissions/
            api.ts
            hooks.ts
            schema.ts
          contests/
            api.ts
            hooks.ts
            schema.ts
        lib/
          http-client.ts
          query-client.ts
          env.ts
        test/
          setup.ts
        types/
          ui.ts
      tests/
        problems.spec.ts
        submissions.spec.ts

    api/
      go.mod
      cmd/
        api/
          main.go
      internal/
        http/
          router.go
          middleware.go
          health_handler.go
          auth_handler.go
          problem_handler.go
          submission_handler.go
          contest_handler.go
        auth/
          service.go
          repository.go
          password.go
          session.go
        problem/
          service.go
          repository.go
          validator.go
        submission/
          service.go
          repository.go
          validator.go
        contest/
          service.go
          repository.go
        queue/
          client.go
          tasks.go
        source/
          local.go
          minio.go
        db/
          sqlc.yaml
          generated/
        platform/
          config.go
          logger.go
      tests/
        integration/
          submission_test.go

    judge-worker/
      go.mod
      cmd/
        worker/
          main.go
      internal/
        judge/
          processor.go
          repository.go
          executor.go
          aggregator.go
        queue/
          tasks.go
        sandbox/
          runner.go
          nsjail.go
        language/
          cpp.go
          java.go
          python.go
        source/
          loader.go
        result/
          mapper.go

  packages/
    database/
      migrations/
        000001_init.sql
        000002_indexes.sql
      query/
        users.sql
        problems.sql
        submissions.sql
        contests.sql
      seed/
        local.sql
    shared/
      package.json
      src/
        enums.ts
        auth.schemas.ts
        problem.schemas.ts
        submission.schemas.ts
        contest.schemas.ts
        index.ts
    tsconfig/
      base.json

  infra/
    compose.local.yml
    docker/
      api.Dockerfile
      web.Dockerfile
      judge-worker.Dockerfile
    env/
      api.env.example
      worker.env.example
      web.env.example

  docs/
    architecture/
      repo-structure-zh.md
      data-model.md
      judge-sandbox.md
    plans/
      2026-05-11-online-judge-tech-stack-zh.md

  tmp/
    submissions/

  .github/
    workflows/
      ci.yml

  package.json
  pnpm-workspace.yaml
  go.work
```

## 前端目录说明

前端建议按“页面路由”和“业务功能”两层拆分：

- `src/routes` 负责页面入口。
- `src/features` 负责按业务域组织 API 调用、hooks、schema。
- `src/components` 只放可复用 UI 组件，不放接口逻辑。
- `src/lib` 放全局基础设施，如请求封装、QueryClient、环境变量解析。

这样可以避免两个常见问题：

- 所有代码都堆到 `components/`
- 所有业务都堆到 `pages/`

## 后端 API 目录说明

`apps/api/internal` 建议按“传输层 + 领域层 + 基础设施层”分开：

- `http/`：只处理路由、请求解析、响应序列化、中间件。
- `auth/`、`problem/`、`submission/`、`contest/`：领域服务和 repository 适配。
- `queue/`：只负责向 Redis/Asynq 投递任务。
- `source/`：只负责源码存储，后续可从本地切到 MinIO。
- `db/`：只放 sqlc 配置和生成代码。
- `platform/`：配置、日志、通用运行时支撑。

关键点是：`http/` 不应该直接写 SQL，`submission/` 不应该直接知道 chi 路由细节。

## 判题 Worker 目录说明

`judge-worker` 不应该复用 `api/internal/submission` 的业务目录。它有自己独立的职责：

- `judge/`：消费任务、组织判题流程、汇总结果。
- `queue/`：定义消费的任务类型。
- `sandbox/`：隔离执行层。
- `language/`：不同语言的编译/运行配置。
- `source/`：加载源码和测试数据。
- `result/`：把执行结果映射成数据库状态。

这样后续要接 `nsjail`、编译器、测试点循环时，不会把 API 代码拖脏。

## 数据库与共享包说明

`packages/database` 建议继续保留为数据库唯一入口，但查询文件应按领域拆开：

- `users.sql`
- `problems.sql`
- `submissions.sql`
- `contests.sql`

不要把所有 SQL 堆在一个文件里。

`packages/shared` 只保留前端真正需要共享的 schema 和枚举。不要把 Go 后端 DTO 镜像一份 TypeScript 类型，再反向要求两边同步，那样维护成本高。

## 当前项目到目标结构的最小演进

可以按这个顺序逐步整理，不需要一次性重构：

1. 前端把 `src/pages` 改成 `src/routes`，新增 `src/features`。
2. API 把更多 handler 和 service 按领域拆进 `auth/`、`problem/`、`submission/`。
3. `packages/database/query` 按领域拆分 SQL 文件。
4. worker 补齐 `language/`、`result/`、`source/`，但只在真正需要时创建文件。
5. `infra/env` 拆分不同进程的环境变量模板。

## 不建议的结构

下面这些目录组织方式后面基本都会变差：

- 前后端都叫 `src/`，但没有 `apps/web` 和 `apps/api` 这种边界。
- API 和 worker 共用一个 `service/` 目录。
- SQL、迁移、seed、生成代码散落在 `apps/api` 各处。
- 前端所有接口调用都塞进 `components/`。
- 为未来预建大量空目录。

## 结论

如果你的目标是“前后端合理分开”，那当前方向是对的，但还不够细。最推荐的做法不是换架构，而是在现有 `apps/` 和 `packages/` 基础上，把：

- 前端拆成 `routes + features + components + lib`
- 后端拆成 `http + domain + infra`
- Worker 拆成 `judge + sandbox + language + result`

这样后续功能继续长，也不会互相缠在一起。
