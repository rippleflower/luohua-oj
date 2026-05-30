# Admin Web 本地联调前置检查

## 1) 基础进程
- API: `http://localhost:8080`
- Admin Web: `http://localhost:5173`
- Web: `http://localhost:5174`
- Judge Worker: 运行中（用于重判/队列联调）

## 2) 数据库迁移
- 目标：`packages/database/migrations` 至少包含并执行 `000004_auth_admin.sql` 的 `Up`。
- 缺失该迁移时，`/auth/login` 可能返回：
  - `relation "admin_permission_grants" does not exist`

## 3) 本地管理员账号
- 确保有可登录管理员账号（建议 `SUPER_ADMIN`）用于完整后台联调。
- 若只有 `USER` 角色，会在后台写接口遇到 `403 forbidden`。

## 4) Admin Web 环境变量
- `VITE_API_BASE_URL`：默认可空（走 Vite `/auth|/admin|/me` 代理）或显式指定 API 地址。
- `VITE_WEB_BASE_URL`：前台预览跳转地址，默认 `http://localhost:5174`。
- 示例文件：`infra/env/admin-web.env.example`。

## 5) 联调验收建议顺序
1. 登录：`/login` 成功后跳转 `/`。
2. 读链路：`/users`、`/problems`、`/contests`、`/submissions`、`/system/settings`、`/announcements`、`/audit`。
3. 写链路：题目创建/更新/发布、比赛创建/更新/编排/冻结、重判、系统配置更新、公告创建。
4. 错误态：401/403/400 页面反馈可见，不静默失败。
