# Next Session Planning Notes

这个目录留给下一次计划模式直接接手。

## Recommended Starting Points

- `docs/logs/development-log.md`
- `docs/plans/templates/plan-mode-template.md`
- `docs/plans/next-session/2026-05-18-auth-admin-followup.md`
- `docs/plans/next-session/2026-05-24-pr-hardening-followup.md`

## Current Focus Candidates

1. auth / admin 后续收口
   - 入口：`docs/plans/next-session/2026-05-18-auth-admin-followup.md`
   - 目标：补齐后台题目/比赛/重判写链路与权限测试

2. PR 收口与下一阶段基线
   - 入口：`docs/plans/next-session/2026-05-24-pr-hardening-followup.md`
   - 目标：基于当前 draft PR 收口评审风险，并推进后台关键写链路

3. sqlc 与 role 枚举清理
   - 目标：移除 `PROBLEM_SETTER` 残留，避免后续生成代码失真

4. 密码哈希正式化
   - 目标：从当前 fallback 收口到计划内方案，或明确保留临时实现的边界

## Session Bootstrap Checklist

- 确认本地前端是否在运行：
  - `pnpm dev:web`
- 确认后台前端是否在运行：
  - `pnpm --filter @oj/admin-web dev`
- 如果要验证 submissions 后端：
  - `go test ./apps/api/...`
- 如果要继续 auth / admin：
  - `pnpm test`
  - `pnpm build:web`
  - `pnpm build:admin`
