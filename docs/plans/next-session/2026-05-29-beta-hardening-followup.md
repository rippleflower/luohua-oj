# Beta Hardening Follow-up

## Summary

- 当前基线已经不是“缺题目/比赛后台写链路”的阶段。
- 当前重点是把现有能力继续收口成更稳的 beta 基线：
  - live smoke
  - infra/documentation cleanup
  - broader regression coverage

## Current State

- 已完成：
  - 用户站默认关闭静默 demo fallback
  - 题目、比赛、补题、提交链路在非 demo mode 下会暴露真实错误态
  - 后台题目详情与题面内容编辑已接通
  - 公共题目详情样例已从 object key 刷新为真实文本
  - 密码哈希已迁到 `argon2id`，并兼容旧 `sha256$...`
  - root script 与 CI 已对齐 shared/web/admin/go 质量门槛
- 仍待补强：
  - live end-to-end smoke 还没有自动化
  - infra 配置里还有少量旧 local-storage/minio 叙述混杂
  - 回归覆盖仍偏向 targeted tests，缺更长链路的自动验证

## Next Steps

1. 增加 live smoke
   - 目标：
     - 启动依赖服务
     - 完成登录
     - 创建并发布题目
     - 创建提交并触发重判
     - 冻结比赛并读取公开详情
   - 优先选择 API 级 smoke，再决定是否上浏览器级检查

2. 清理 infra 说明
   - 明确 `SOURCE_ROOT` 是当前默认对象存储路径
   - 区分“仍被 compose 保留的历史组件”和“当前运行真正依赖的组件”

3. 扩大回归覆盖
   - 增加更多 problem/contest/rejudge 的 integration coverage
   - 让 public read model 刷新和 admin write path 的断言更系统化

## Verification

- `pnpm ci:check`
- `pnpm dev:stack`
- `pnpm dev:check`
- 选取一条完整路径做人工 smoke：
  - 登录
  - 打开 `/problems/two-sum`
  - 提交一次代码
  - 后台触发重判
  - 打开 `/contests/<slug>`
