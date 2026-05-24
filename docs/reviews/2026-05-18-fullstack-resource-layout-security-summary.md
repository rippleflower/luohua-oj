# 全栈资源、布局与安全总览

日期：2026-05-18

## 摘要

本报告基于当前仓库的静态检查，对 `apps/web`、`apps/admin-web`、`apps/api/internal/http`、`packages/shared` 和 `infra/env/*.example` 做结构化总结。目标不是渗透测试，而是把当前资源组织、页面布局、接口暴露面和主要 Web 安全边界说清楚，并指出未来最容易引入漏洞的位置。

结论先行：

- 当前前端未发现 `dangerouslySetInnerHTML`、`innerHTML`、`eval`、`new Function` 或类似动态 HTML / 代码执行入口。
- 当前认证与写操作保护的主干是成立的：session cookie 为 `HttpOnly`，写接口依赖 `oj_csrf` + `X-CSRF-Token` 双提交校验，后台写路由同时受权限中间件保护。
- 当前未看到明显高危的前端 XSS、任意脚本执行或“仅靠前端隐藏菜单”的权限绕过设计。
- 主要风险不在现有只读/表单式页面，而在未来若引入富文本题面、公告 HTML、对象存储公开访问、或把底层错误原样透出时，容易从“安全边界清晰”退化为“内容可信度不清晰”。

## 一、资源盘点

### 1. 前端应用

- 用户站：`apps/web`
  - 路由入口在 `apps/web/src/app.tsx`
  - 页面包括 `/`、`/problems`、`/problems/:slug`、`/contests`、`/contests/:slug`、`/submissions`、`/submissions/:id`、`/login`、`/register`、`/me`、`/settings/*`
  - 壳层在 `apps/web/src/components/layout/app-shell.tsx`
- 后台站：`apps/admin-web`
  - 路由入口在 `apps/admin-web/src/app.tsx`
  - 页面包括 `/login`、`/`、`/users`、`/users/:id`、`/users/:id/permissions`、`/problems`、`/contests`、`/submissions`、`/announcements`、`/system/*`、`/audit`
  - 壳层在 `apps/admin-web/src/components/layout/admin-shell.tsx`

### 2. 共享契约

- 共享 schema 位于 `packages/shared/src`
- 认证与会话契约：`auth.schemas.ts`
- 后台概览、用户、题目、审计、系统设置契约：`admin.schemas.ts`
- 这些 schema 是前端数据形状的第一道约束，降低了接口字段漂移直接进入 UI 的概率

### 3. 样式与视觉资源

- 用户站样式入口：`apps/web/src/styles.css`
- 后台样式入口：`apps/admin-web/src/styles.css`
- 当前两站都通过 `@import` 引入 Google Fonts：
  - `IBM Plex Sans`
  - `IBM Plex Mono`
  - `Newsreader`
- 当前视觉语言：
  - 主站：冷暖混合的浅色渐变、内容型信息架构
  - 后台：暖灰底、卡片式操作台、表单密度更高

## 二、布局总结

### 1. 用户站布局

- 顶部壳层强调公开导航和登录入口
- 支持轻量本地语言切换，状态存于 `localStorage`
- 页面头部采用统一标题/副标题区域，方便内容页复用
- 当前路由仍是基于 `window.location.pathname` 的条件分发，不是专门路由库

### 2. 后台布局

- 左侧主导航 + 右侧内容区的控制台式结构
- 导航项按权限过滤显示，但真实授权仍以后端为准
- `/problems` 已进入“列表 + 创建/编辑/发布表单”的操作面板模式
- 后台布局明显偏向运维/运营台，而不是公共内容页

### 3. 当前品牌状态

- 当前实现并未采用 `brand-guidelines` 中的 Anthropic 色板和字体组合
- 现状更接近：
  - 浅底渐变
  - IBM Plex 系正文 / monospace
  - Newsreader 大标题
- 若后续要做品牌化展示，建议只对审查产物做映射，不直接把现站样式解释为品牌规范已对齐

## 三、接口与暴露面

### 1. 公共读接口

- `GET /problems`
- `GET /problems/{slug}`
- `GET /contests`
- `GET /contests/{slug}`
- `POST /submissions`
- `GET /submissions/{submissionID}`
- `GET /users/{username}/submissions`

### 2. 认证与个人中心接口

- `/auth/register`
- `/auth/login`
- `/auth/me`
- `/auth/logout`
- `/auth/password/change`
- `/auth/sessions`
- `/auth/sessions/revoke`
- `/me/summary`
- `/me/settings`
- `/me/profile`
- `/me/preferences`

### 3. 后台接口

- 后台接口统一挂在 `/admin/*`
- 权限维度至少覆盖：
  - dashboard.view
  - users.view / users.edit / users.roles
  - problems.view / problems.edit / problems.publish
  - contests.view / contests.edit / contests.publish
  - submissions.view / submissions.rejudge
  - announcements.view / announcements.edit
  - system.view / system.edit
  - audit.view

这意味着后台暴露面不是“一个后台站”，而是一组细粒度权限控制的受保护接口集合。

## 四、安全审查

### 1. 前端执行面

静态检查范围内，未发现以下高风险入口：

- `dangerouslySetInnerHTML`
- `innerHTML`
- `eval(...)`
- `new Function(...)`
- 显式动态脚本注入

这说明当前页面主要依赖 React 默认转义和结构化数据渲染，XSS 表面积较小。

### 2. Cookie、会话与 CSRF

当前会话模型清晰：

- session cookie：
  - `HttpOnly`
  - `SameSite=Lax`
  - 前端不可直接读取
- CSRF cookie：
  - 名称为 `oj_csrf`
  - 非 `HttpOnly`
  - 前端读取后通过 `X-CSRF-Token` 回传
- 中间件 `requireCSRF` 要求：
  - 存在 `oj_csrf` cookie
  - 存在 `X-CSRF-Token`
  - 两者值一致

这意味着当前写操作保护依赖“双提交 cookie”模式，而不是仅依赖 `SameSite`。

### 3. 权限边界

- 后台导航会按 viewer 权限过滤，但这只是可见性优化
- 真正的授权在 API 中由 `requireAuthenticated` + `requireAdminPermission` 完成
- 这是一条正确的边界线：前端隐藏菜单不等于后端放弃校验

普通用户访问后台的边界可总结为：

- 未登录：应被拦在认证层
- 已登录普通用户：即使能手动构造请求，也要被权限层拒绝
- 管理员：只能访问自己被授予的管理能力
- `SUPER_ADMIN`：作为显式特权角色统一放行

### 4. 错误回显边界

- 用户站 HTTP client 当前主要向 UI 抛出 `request failed: <status>`
- 后台站 HTTP client 会尝试读取 JSON `error` 字段并显示业务错误
- 这意味着：
  - 前台页面信息泄露面较小
  - 后台页面更利于运营使用，但也更依赖后端错误文案是否被良好约束

当前未看到把堆栈直接渲染到前端的设计，但仍需注意：

- 若底层数据库错误、对象存储错误、内部路径错误被直接透传到 `error` 字段，后台 UI 会忠实显示它们
- 因此“前端不暴露内部细节”并不完全取决于前端，而取决于后端是否持续把错误收口成业务级消息

### 5. 本地存储与配置

- `localStorage` 当前用于语言偏好与部分前端本地记录，不承载 session 主凭证
- `document.cookie` 的读取用途集中在 CSRF token，而不是 session token
- `infra/env/*.example` 暴露的是示例配置项，不是实密钥

这说明当前配置样板没有直接把生产秘密提交到仓库，但示例中仍然明确了系统依赖：

- PostgreSQL
- Redis
- 本地对象/源码存储根目录
- Web 端 API Base URL

### 6. 第三方依赖与供应链边界

当前显式第三方前端资源依赖主要是 Google Fonts。

这不是现成漏洞，但属于应记录的边界：

- 可用性风险：外部字体服务不可达时会退回系统字体
- 供应链/隐私风险：页面渲染依赖外域资源
- 品牌一致性风险：如果后续强调离线部署或内网部署，需要改成本地字体托管或系统字体回退

## 五、风险结论

### 1. 已有防线

- React 默认转义渲染，未见危险 HTML 注入点
- session cookie 与 CSRF cookie 职责分离
- 写接口有明确 CSRF 校验
- 后台关键接口有服务端权限校验
- 共享 schema 对前端数据形状有基本约束

### 2. 当前未发现的高危问题

- 未发现明显前端 XSS 注入点
- 未发现直接把 session token 放进 `localStorage` 的设计
- 未发现只靠前端菜单隐藏来控制后台权限
- 未发现仓库中提交真实密钥或生产运行态敏感值

### 3. 仍需注意的边界

- 后台错误文案如果继续直接透传底层错误，可能演变成内部信息泄露
- 第三方字体依赖不是漏洞，但对离线部署和供应链治理不友好
- 目前前端路由仍是 pathname 分发，未来若加深交互，需持续避免绕开统一鉴权/错误处理层

### 4. 最容易在后续扩展中引入漏洞的位置

- 富文本题面、公告内容、运营内容若改为 HTML 渲染，最容易引入 XSS
- 外链、对象存储公开地址、附件下载若缺少白名单与内容类型控制，容易扩大注入或越权面
- 若未来为了调试而在前端展示原始后端错误、对象 key、内部路径，容易扩大信息泄露面
- 若把更多敏感状态转移到 `localStorage`，会明显削弱当前会话边界

## 六、品牌映射建议

本次交付如果需要品牌化展示，建议只应用在“审查摘要”本身：

- 深色：`#141413`
- 浅底：`#faf9f5`
- 中灰：`#b0aea5`
- 浅灰：`#e8e6dc`
- 强调色：
  - 橙 `#d97757`
  - 蓝 `#6a9bcc`
  - 绿 `#788c5d`
- 标题字体：`Poppins, Arial, sans-serif`
- 正文字体：`Lora, Georgia, serif`

建议不要把这些映射误写成“当前站点已采用的规范”，而应写成“用于审查报告展示的品牌建议映射”。

## 结论

如果只看当前静态实现，系统已经具备比较清晰的基础 Web 安全边界：会话、CSRF、权限和结构化渲染的主干都在。当前最大的价值不是“立刻修复高危漏洞”，而是维持这条边界不被未来的富文本、外链、原始错误透传和对象存储暴露面扩张破坏。
