# Auth / Admin Follow-up Plan

> Historical note: this plan predates the 2026-05-29 baseline. Core admin write paths, rejudge, and Argon2 migration are now in place. Use [2026-05-29-beta-hardening-followup.md](/Volumes/新加卷/project/luohua-oj/docs/plans/next-session/2026-05-29-beta-hardening-followup.md) for current follow-up.

## Summary

- 目标：把本轮已经打通的认证、个人中心、独立后台从“可登录、可读”推进到“关键后台写链路可用、权限更稳、测试更完整”。
- 本轮范围：
  - 完成后台题目/比赛/提交的关键写接口
  - 补足认证与权限的测试覆盖
  - 把密码哈希从当前标准库 fallback 收口到计划内实现
  - 校正共享类型、迁移和实现之间的剩余不一致
- 不做的内容：
  - 不引入 OAuth、MFA、短信、验证码登录
  - 不扩展 CMS、审批流、工单系统
  - 不做新的前端设计方向切换，沿用现有用户站/后台壳层

## Current State

- 已完成：
  - 数据库新增认证与后台基础表，见 [000004_auth_admin.sql](/Volumes/新加卷/project/luohua-oj/packages/database/migrations/000004_auth_admin.sql)
  - API 已有 `/auth/*`、`/me/*`、`/admin/*` 路由骨架
  - 用户站已接 `/login`、`/register`、`/me`、`/settings/*`
  - 独立后台站 `apps/admin-web` 已创建并可构建
  - `pnpm test`、`pnpm build:web`、`pnpm build:admin`、`pnpm test:go` 当前通过
- 已知问题：
  - 密码哈希当前是标准库迭代 `SHA-256`，不是原计划里的 Argon2
  - 后台题目/比赛接口目前只有列表读，没有创建、编辑、发布、冻结写链路
  - `/admin/submissions/:id/rejudge`、`/judge/queue` 还没实现
  - 审计日志已落表，但覆盖面还不够完整
  - `system_settings` 当前只存摘要 JSON，没有更细粒度 schema 约束
  - `apps/api/internal/db/generated` 仍保留旧 `user_role` 枚举定义，当前手写 SQL 没被它挡住，但后续如果继续依赖 sqlc 代码生成会出错
- 外部依赖：
  - 如需切回 Argon2，先确认 Go 依赖拉取稳定
  - 如需回归 sqlc 流程，要同步更新 schema / queries / 生成结果

## Critical Gaps

1. 题目后台写链路未完成
   - 缺 `POST /admin/problems`
   - 缺 `PATCH /admin/problems/:id`
   - 缺 `POST /admin/problems/:id/publish`
   - 缺版本、标签、manifest objectKey 的管理闭环

2. 比赛后台写链路未完成
   - 缺 `POST /admin/contests`
   - 缺 `PATCH /admin/contests/:id`
   - 缺 `POST /admin/contests/:id/freeze`
   - 缺冻结后 `contest_snapshots / contest_problem_snapshots / contest_public_*` 刷新

3. 判题后台操作未完成
   - 缺 `POST /admin/submissions/:id/rejudge`
   - 缺 judge queue 读模型或队列摘要接口
   - 缺相关审计事件

4. 认证与权限测试偏薄
   - 普通用户访问 `/admin/*` 的 401/403 组合还缺系统性测试
   - `ADMIN` 与 `SUPER_ADMIN` 的权限差异缺 API 级测试
   - 改密码后“保留当前会话、吊销其他会话”的行为需要显式测试

## Execution Plan

1. 收口认证与权限基础
   - 为 `auth.Service` 和 HTTP middleware 增加测试
   - 明确并测试：
     - 未登录访问 `/me/*` 返回 401
     - `USER` 访问 `/admin/*` 返回 403
     - `ADMIN` 无权限时返回 403
     - `SUPER_ADMIN` 自动通过全部 `requireAdminPermission`
   - 优先检查：
     - [auth_middleware.go](/Volumes/新加卷/project/luohua-oj/apps/api/internal/http/auth_middleware.go)
     - [service.go](/Volumes/新加卷/project/luohua-oj/apps/api/internal/auth/service.go)
     - [repository.go](/Volumes/新加卷/project/luohua-oj/apps/api/internal/auth/repository.go)

2. 完成题目后台写链路
   - 先只做最小可用版本：
     - 创建题目基础记录
     - 更新标题、slug、难度、时空限制
     - 发布当前版本并刷新 `problem_public_summaries/details`
   - 如果一次做不完版本管理 UI，先保证 API 可用，后台前端可用表单直连
   - 重点文件：
     - `apps/api/internal/problem/*`
     - `apps/api/internal/http/admin_handler.go`
     - `apps/admin-web/src/routes/problems/*`

3. 完成比赛后台写链路
   - 最小可用范围：
     - 创建比赛
     - 更新比赛基础信息
     - 绑定题目列表
     - 执行 freeze，生成 snapshot 和 contest public read models
   - 冻结逻辑完成后补比赛后台页面的 snapshot 状态区
   - 重点文件：
     - `apps/api/internal/contest/*`
     - `apps/api/internal/http/admin_handler.go`
     - `apps/admin-web/src/routes/contests/*`

4. 完成重判与判题后台操作
   - 增加 `POST /admin/submissions/:id/rejudge`
   - 增加 `/admin/judge/queue` 或把 `/judge/queue` 映射成后台只读接口
   - 每次重判都要写 `audit_logs`
   - 重点文件：
     - `apps/api/internal/submission/*`
     - `apps/api/internal/queue/*`
     - `apps/admin-web/src/routes/submissions/*`

5. 清理密码哈希与生成代码问题
   - 如果网络稳定：
     - 切回 Argon2
     - 增加迁移兼容或兼容旧 hash 校验
   - 如果继续保持当前标准库方案：
     - 至少在计划和代码注释里标清这是临时 fallback
   - 然后处理 `sqlc` 旧 `user_role` 生成物：
     - 更新 schema / query / generated
     - 避免后续出现 `PROBLEM_SETTER` 残留

## Verification

- 命令：
  - `pnpm test`
  - `pnpm build:web`
  - `pnpm build:admin`
  - `go test ./apps/api/...`
- 页面路径：
  - 用户站：
    - `/login`
    - `/me`
    - `/settings/security`
  - 后台站：
    - `/login`
    - `/`
    - `/users`
    - `/problems`
    - `/contests`
    - `/submissions`
    - `/system/settings`
    - `/audit`
- 需要人工确认的点：
  - `ADMIN` 权限裁剪后侧边栏与接口是否一致
  - 发布题目后主站 `/problems` 和 `/problems/:slug` 是否立即命中新快照
  - 冻结比赛后主站 `/contests/:slug` 是否稳定读取 snapshot，而不是读题库当前版本

## Risks

- 如果先做 UI 再补 freeze / publish 刷新逻辑，容易形成“后台提交成功但主站读不到”的假完成状态
- 如果不尽快处理旧 `sqlc` 生成代码，后续继续改表结构时会越来越难回收
- 如果直接切回 Argon2，但网络和依赖缓存不稳定，会再次卡住测试链路

## Handoff Notes

- 下一轮开始前先检查：
  - [docs/plans/next-session/README.md](/Volumes/新加卷/project/luohua-oj/docs/plans/next-session/README.md)
  - [2026-05-18-auth-admin-followup.md](/Volumes/新加卷/project/luohua-oj/docs/plans/next-session/2026-05-18-auth-admin-followup.md)
  - `git status --short`
- 如果 dev server 已停，先执行：
  - `pnpm dev:web`
  - `pnpm --filter @oj/admin-web dev`
- 如果要直接从后端开始：
  - `go test ./apps/api/...`
- 建议本轮优先顺序：
  1. 认证/权限测试
  2. 题目后台写链路
  3. 比赛 freeze 写链路
  4. rejudge + audit
