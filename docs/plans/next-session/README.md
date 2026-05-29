# Next Session Planning Notes

这个目录留给下一次计划模式直接接手。

## Recommended Starting Points

- `docs/logs/development-log.md`
- `docs/plans/templates/plan-mode-template.md`
- `docs/plans/next-session/2026-05-29-beta-hardening-followup.md`

## Current Focus Candidates

1. beta hardening
   - 入口：`docs/plans/next-session/2026-05-29-beta-hardening-followup.md`
   - 目标：补 live smoke、继续扩大回归覆盖、收口 infra/documentation

2. historical context
   - 入口：`docs/plans/next-session/2026-05-18-auth-admin-followup.md`
   - 目标：查看这轮收口前的缺口来源，不把它当成当前事实

3. sqlc 与 role 枚举清理
   - 目标：移除 `PROBLEM_SETTER` 残留，避免后续生成代码失真

4. deployment-oriented automation
   - 目标：把当前 API 级、worker 级、前后端级验证进一步串成可重复 smoke

## Session Bootstrap Checklist

- 确认本地前端是否在运行：
  - `pnpm dev:web`
- 确认后台前端是否在运行：
  - `pnpm --filter @oj/admin-web dev`
- 如果要验证 submissions 后端：
  - `go test ./apps/api/...`
- 如果要继续 auth / admin：
  - `pnpm ci:check`
