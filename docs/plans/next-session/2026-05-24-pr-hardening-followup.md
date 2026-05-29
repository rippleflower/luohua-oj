# PR Hardening Follow-up Plan

> Historical note: this plan predates the 2026-05-29 baseline. The problem/contest admin write path and default demo fallback cleanup have since moved forward. Use [2026-05-29-beta-hardening-followup.md](/Volumes/新加卷/project/luohua-oj/docs/plans/next-session/2026-05-29-beta-hardening-followup.md) for current follow-up.

## Summary

- 目标：
  - 把当前 draft PR 从“功能已接通”推进到“更容易评审和继续开发的稳定基线”。
- 本轮范围：
  - 收口 PR 自检中剩余的高风险点
  - 明确并推进下一阶段优先级最高的后台写链路
  - 保持 `codex/auth-admin-submission-foundation-split` 作为唯一继续开发基线
- 不做的内容：
  - 不再重组一次历史
  - 不做新的大范围 UI 改版
  - 不引入新的路由框架或状态管理框架

## Current State

- 已完成：
  - draft PR 已创建：
    - [PR #1](https://github.com/rippleflower/luohua-oj/pull/1)
  - 分支历史已整理为更小的主题提交
  - 用户端 / 管理端直达入口已接通
  - `/problems` 已支持前端分类筛选
  - `/problems/:slug` 已接入 Markdown 题面 + Monaco 编辑器工作区
  - `AppShell` 重复 action 已修复
  - `apps/web` 已做按路由懒加载拆包，构建不再出现 `>500 kB` warning
- 已知问题：
  - PR 仍然跨 schema、API、worker、web、admin-web 多层
  - 后台核心写链路仍然是下一阶段最大空缺：
    - 题目创建/编辑/发布
    - 比赛创建/编辑/freeze
    - rejudge 管理动作
  - problem workspace 仍保留 `compatUserId` 兼容输入
  - judge / submission 闭环虽然已能运行，但还缺更强的回归和运维视角验证
- 外部依赖：
  - 如需做页面验证，先确认本地 dev server 存活
  - 如需继续推 PR 状态，优先沿用当前 GitHub CLI / API 路径

## Execution Plan

1. 先做 PR review-hardening 自检
   - 重新浏览以下关键路径，按 reviewer 视角找遗漏：
     - 用户端：
       - `/problems`
       - `/problems/:slug`
       - `/submissions`
     - 管理端：
       - `/`
       - `/problems`
       - `/contests`
       - `/submissions`
   - 检查点：
     - 权限入口是否一致
     - 题目页提交链路是否清晰
     - 后台跳前台、前台跳后台的 URL 回退是否正确

2. 优先补后台题目写链路
   - 目标是最小可用后台写路径：
     - `POST /admin/problems`
     - `PATCH /admin/problems/:id`
     - `POST /admin/problems/:id/publish`
   - 然后让 `apps/admin-web/src/routes/problems/*` 有可用表单入口
   - 完成后必须确认主站 `/problems` 与 `/problems/:slug` 命中新发布数据

3. 再补比赛后台写链路
   - 范围：
     - 创建比赛
     - 更新比赛基础信息
     - 绑定题目
     - freeze 生成 snapshot
   - 重点是避免“后台提交成功，但前台仍读旧版本”的假完成状态

4. 收口 rejudge 管理动作
   - 完成后台重判接口与页面入口
   - 确认 judge queue / task summary 的可读性足以支持人工验证
   - 重判流程要显式确认旧结果清空与终态回写

5. 视时间决定是否处理 `compatUserId`
   - 如果登录用户身份链路已足够稳定：
     - 去掉题目页兼容输入
   - 如果暂时不能去掉：
     - 至少补注释和测试，明确它只是临时 fallback

## Verification

- 命令：
  - `pnpm test`
  - `pnpm --filter @oj/web build`
  - `pnpm --filter @oj/admin-web build`
  - `go test ./apps/api/...`
- 页面路径：
  - 用户端：
    - `/login`
    - `/problems`
    - `/problems/two-sum`
    - `/submissions`
  - 管理端：
    - `/`
    - `/problems`
    - `/contests`
    - `/submissions`
- 需要人工确认的点：
  - 管理员是否只在应出现的位置看到“管理端”入口
  - 题目页切语言后默认模板与 Monaco mode 是否同步
  - 提交后结果面板是否稳定反馈
  - 发布/冻结后前台是否稳定命中新快照

## Risks

- 如果继续往这条 PR 上堆大功能，review 可读性会再次下降
- 如果先补 UI 而不补后台 publish/freeze/rejudge 写链路，会产生更多“可看但不可用”的管理界面
- 如果过早去掉 `compatUserId`，但真实用户身份链路仍有边界缺失，会直接打断当前题目提交流程

## Handoff Notes

- 下一轮开始前先检查：
  - [development-log.md](/Volumes/新加卷/project/luohua-oj/docs/logs/development-log.md)
  - [2026-05-24-pr-hardening-followup.md](/Volumes/新加卷/project/luohua-oj/docs/plans/next-session/2026-05-24-pr-hardening-followup.md)
  - `git status --short`
- 如果 dev server 已停，先执行：
  - `pnpm dev:web`
  - `pnpm --filter @oj/admin-web dev`
- 如果从评审准备开始：
  - `gh pr view 1`
  - `pnpm test`
- 建议本轮优先顺序：
  1. 后台题目写链路
  2. 比赛 freeze 写链路
  3. rejudge 管理动作
  4. `compatUserId` 清理或边界固化
