#!/usr/bin/env bash
set -euo pipefail

echo "==> repo hygiene"
pnpm check:repo-hygiene

echo "==> shared tests"
pnpm --filter @oj/shared test

echo "==> web tests"
pnpm --filter @oj/web test

echo "==> web build"
pnpm --filter @oj/web build

if [ -f "apps/admin-web/package.json" ]; then
  echo "==> admin-web tests"
  pnpm --filter @oj/admin-web test

  echo "==> admin-web build"
  pnpm --filter @oj/admin-web build
else
  echo "==> admin-web checks skipped (apps/admin-web/package.json not found)"
fi

echo "==> go tests"
go test ./apps/api/... ./apps/judge-worker/...

echo "==> ci:check passed"
