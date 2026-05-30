# Contributing / 提交流程

## Required local gate / 本地强制门禁
- Run `pnpm ci:check` before opening or updating a PR.
- 在创建或更新 PR 前，必须先执行 `pnpm ci:check`。

This includes:
- repo hygiene check (`._*` / `.DS_Store` must not be tracked)
- `@oj/shared` tests
- `@oj/web` tests + build
- `@oj/admin-web` tests + build (auto-skipped when `apps/admin-web/package.json` does not exist)
- go tests (`./apps/api/... ./apps/judge-worker/...`)

## Optional local push guard / 可选本地 push 防线
Enable repo hooks once:

```bash
git config core.hooksPath .githooks
```

After enabled, every `git push` will run `pnpm ci:check`.

## Pull Request checklist / PR 检查清单
Use the PR template and ensure:
- summary and risk are written in both Chinese and English
- `pnpm ci:check` passed locally
- docs are updated when behavior changes
