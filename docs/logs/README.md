# Logs

本目录记录三类与日志相关的内容：

1. 运行日志：API 和 Judge Worker 输出到 `tmp/logs/*.log`
2. 前端调试日志：浏览器 console 中的结构化 JSON 事件
3. 工程记录文档：本目录中的规范、决策和阶段记录

## Services

- API
  - stdout: JSON
  - file: `tmp/logs/api.log`
- Judge Worker
  - stdout: JSON
  - file: `tmp/logs/judge-worker.log`
- Web
  - browser console: JSON
  - 本次不接远端日志采集

## Key Fields

后端统一字段：

- `service`
- `component`
- `event`
- `submissionId`
- `requestId` 或 `taskId`
- `error`
- `durationMs`

前端 `submissions` 统一字段：

- `event`
- `feature`
- `route`
- `requestId`
- `apiMode`
- `language`
- `userIdPresent`
- `problemIdPresent`
- `errorMessage`
- `httpStatus`

## Sensitive Data Rules

- 不记录源码原文
- 不记录完整表单快照
- `userId` 和 `problemId` 只记录是否提供，必要时只记录截断预览
- 错误日志允许记录错误摘要，不允许回写源码正文

## Local Usage

查看 API 日志：

```bash
tail -f tmp/logs/api.log
```

查看 Worker 日志：

```bash
tail -f tmp/logs/judge-worker.log
```

按事件检索：

```bash
rg 'submission.request.succeeded|submission.enqueue.succeeded' tmp/logs/api.log
```

按 `submissionId` 检索：

```bash
rg 'submissionId' tmp/logs/api.log tmp/logs/judge-worker.log
```
