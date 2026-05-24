# Plans

本目录存放三类计划文件：

1. 长期方案：跨模块、跨阶段的架构或产品实施计划
2. 会话计划：某一次协作要执行的明确任务拆解
3. 模板与草稿：给后续计划模式直接复用的骨架文件

## Suggested Layout

```text
docs/plans/
  README.md
  templates/
    plan-mode-template.md
  next-session/
    README.md
```

## Usage

- 长期计划：直接以日期开头命名，例如 `2026-05-16-submissions-followup.md`
- 临时计划：先放到 `next-session/`，确认后再升级成正式日期文件
- 进入计划模式时，优先读取：
  - `docs/plans/next-session/README.md`
  - `docs/plans/templates/plan-mode-template.md`

## Naming

- 推荐格式：`YYYY-MM-DD-topic.md`
- 中文文件名可以用，但英文 topic 更适合后续检索和引用
