# 前端路由设计

> 状态：Current
> 事实源：`code/frontend/src/router/`、`code/frontend/src/config/backofficeNavigation.ts`、`code/frontend/src/utils/roleRoutes.ts`
> 替代：无

## 定位

本文档只说明前端页面路由的注册方式、命名空间、权限守卫、错误页回退和导航匹配规则。

- 覆盖：`Vue Router` 路由树、`/academy/*` 与 `/platform/*` 命名空间、登录态恢复、角色校验、错误状态页和默认首页映射。
- 不覆盖：页面内部的 tab/query 状态、列表筛选、异步数据加载和业务流程编排；这些能力由 `code/frontend/src/features/**/model` 或 `code/frontend/src/composables/use*.ts` 负责。

## 当前设计

- `code/frontend/src/router/index.ts`、`code/frontend/src/router/routes/appShellRoute.ts`
  - 负责：组装认证路由、主应用壳、错误页和工具页；在 `/` 下挂载学生端、教师端和平台端子路由，并把根路径重定向到 `/student/dashboard`
  - 不负责：承载页面内的数据请求、筛选状态或业务动作

- `code/frontend/src/router/routes/studentRoutes.ts`、`teacherRoutes.ts`、`platformRoutes.ts`
  - 负责：声明各工作区的正式 URL、组件入口和 `meta.requiresAuth / meta.roles / meta.title / meta.icon`
  - 不负责：根据角色动态拼接第二套路由树，也不在页面组件里复制权限判断

- `code/frontend/src/router/guards.ts`
  - 负责：公开页放行、登录态恢复、未登录跳转 `/login?redirect=...`、角色不匹配跳 `/403`、异常时登出并回到登录页、`afterEach` 更新标题、`router.onError` 跳 `/500`
  - 不负责：显示复杂业务降级 UI；守卫只做导航级回退

- `code/frontend/src/config/backofficeNavigation.ts`、`code/frontend/src/utils/roleRoutes.ts`、`code/frontend/src/utils/routeTitle.ts`
  - 负责：教师/管理员导航高亮、详情页回归所属目录项、角色默认首页和页面标题解析
  - 不负责：替代路由注册本身，也不决定页面内的 tab 或二级状态

## 1. 运行入口

路由注册从 `code/frontend/src/router/index.ts` 开始：

1. 挂载 `authRoutes`
2. 挂载 `appShellRoute`
3. 挂载 `errorRoutes`
4. 挂载 `utilityRoutes`
5. 在 `setupRouterGuards(router)` 中接入全局守卫

`appShellRoute` 当前是整个受保护应用的唯一壳：

- 路径：`/`
- 组件：`@/components/layout/AppLayout.vue`
- 默认跳转：`/student/dashboard`
- 子路由来源：
  - `studentRoutes`
  - `teacherRoutes`
  - `platformRoutes`

## 2. 路由命名空间与分组

### 2.1 公开页与错误页

| 分组 | 当前入口 | 说明 |
| --- | --- | --- |
| 认证页 | `/login`、`/register` | 来自 `authRoutes.ts`，不要求登录 |
| 错误页 | `/401`、`/403`、`/404`、`/429`、`/500`、`/502`、`/503`、`/504` | 来自 `errorRoutes.ts`，由守卫和请求层回退 |
| 工具页 | `/ui-lab` | 仅 `admin` 可访问，不属于正式业务导航入口 |
| 兜底 | `/:pathMatch(.*)* -> /404` | 来自 `utilityRoutes.ts` |

### 2.2 学生端与共享题目页

学生工作区当前不是“所有页面都带 `/student/*` 前缀”的模型，而是混合命名：

| 路由组 | 当前正式路径 | 说明 |
| --- | --- | --- |
| 学生首页 | `/student/dashboard` | `/dashboard` 只是兼容 redirect |
| 学生实例 | `/student/instances` | `/instances` 只是兼容 redirect |
| 能力画像 | `/student/skill-profile` | `/skill-profile` 只是兼容 redirect |
| 题目与竞赛共享页 | `/challenges`、`/challenges/:id`、`/contests`、`/contests/:id`、`/scoreboard`、`/scoreboard/:contestId` | 面向学生和教师共用 |
| 通知与个人资料 | `/notifications`、`/notifications/:id`、`/profile`、`/settings/security` | 通过 `meta.roles` 或默认登录态控制 |

约束：

- 不能把现有学生路由误写成“统一的 `/student/*` 命名空间”。
- 页面内部 tab 状态不进入顶层路由表，继续放在 feature model 里处理，例如 `useContestDetailRoutePage`、`useScoreboardRoutePage`、`useStudentDashboardPage`。

### 2.3 教师工作区

`teacherRoutes.ts` 当前正式入口是 `/academy/*`：

| 路由组 | 当前正式路径 | 兼容入口 |
| --- | --- | --- |
| 教学概览 | `/academy/overview` | `/teacher/dashboard` |
| 班级与学生 | `/academy/classes`、`/academy/students`、`/academy/classes/:className/**` | 对应 `/teacher/classes...`、`/teacher/students...` |
| AWD 复盘 | `/academy/awd-reviews`、`/academy/awd-reviews/:contestId` | `/teacher/awd-reviews...` |
| 实例管理 | `/academy/instances` | `/teacher/instances` |

说明：

- `/teacher/*` 仍存在于当前代码里，但只做 redirect 兼容，不是新的活动命名空间。
- 教师端多数页面同时允许 `teacher`、`admin` 访问，最终权限仍以 `meta.roles` 判定。

### 2.4 平台工作区

`platformRoutes.ts` 当前正式入口是 `/platform/*`：

| 路由组 | 当前正式路径 | 权限说明 |
| --- | --- | --- |
| 平台总览与教学目录 | `/platform/overview`、`/platform/classes`、`/platform/students`、`/platform/classes/:className/**`、`/platform/awd-reviews/**`、`/platform/instances` | 主要面向 `admin` |
| 题目创作与题库管理 | `/platform/challenges/**`、`/platform/awd-challenges/**` | 当前允许 `teacher`、`admin` |
| 赛事运维与大屏 | `/platform/contest-ops/**`、`/platform/contests/:contestId/**` | 当前允许 `admin`，导航匹配由 `backofficeNavigation.ts` 维护 |
| 治理类页面 | 用户、通知、镜像、审计等 `platform` 目录页面 | 当前由 `admin` 使用 |

说明：

- 当前前端页面路由树里没有活动的 `/admin/*` 页面命名空间。
- `/platform/*` 表示平台工作区 owner，而不是“所有页面都只能由 admin 访问”；共享创作页仍以 `meta.roles` 决定访问角色。

## 3. 路由守卫与回退

`code/frontend/src/router/guards.ts` 当前守卫顺序如下：

1. 命中公开页 `/login` 或 `/register` 时直接放行
2. 对公开页也会尝试 `ensureSessionRestored()`，已登录用户访问登录/注册页时，按 `redirect` 参数或 `getRoleDashboardPath()` 跳转
3. 受保护页面在首次进入时调用 `authStore.restore()`
4. 恢复后仍未登录时，跳转到 `/login?redirect=${to.fullPath}`
5. `meta.roles` 与当前用户角色不匹配时，提示“无权限访问该页面”并跳 `/403`
6. 守卫内部出现异常时，执行 `authStore.logout()` 并跳回登录页
7. `afterEach` 调用 `resolveRouteTitle()`，按 `APP_TITLE_PREFIX` 更新页面标题
8. `router.onError` 统一跳到 `/500`

相关路径：

- 登录态恢复：`code/frontend/src/stores/auth.ts`
- 默认首页映射：`code/frontend/src/utils/roleRoutes.ts`
- 标题解析：`code/frontend/src/utils/routeTitle.ts`

## 4. 导航匹配与边界

当前导航匹配不依赖“页面自己知道应该高亮哪个菜单”，而是统一交给 `code/frontend/src/config/backofficeNavigation.ts`。

- 教师与管理员详情页会回映到所属目录项，而不是把每个详情页单独做成导航入口。
- `getBackofficeModuleByPath()` 负责把 `/academy/classes/:className/students/:studentId/review-archive` 这类详情页映射回 `operations` 模块。
- `getVisibleBackofficeSecondaryItems()` 负责根据角色裁剪可见目录项，并保持当前项的 active 状态。

这层规则的直接目的，是避免在 `.vue` 页面里重复写导航归属判断。

## 5. 兼容与历史例外

- `/teacher/*` 当前仍保留 redirect，属于旧命名空间迁移兼容；正式事实源只认 `/academy/*`。
- `/dashboard`、`/instances`、`/skill-profile` 当前仍保留 redirect，属于学生端早期路径兼容。
- `resolveRouteTitle()` 对 `/dashboard` 和 `/student/dashboard` 做了特例处理，允许通过 query/变体路由生成不同标题。

## 6. Guardrail

- 路由 view 不能直接持有路由状态或 query-tab 逻辑：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 教师/管理员导航映射和命名空间匹配：`code/frontend/src/config/__tests__/backofficeNavigation.test.ts`
- 默认首页与角色跳转：`code/frontend/src/utils/roleRoutes.ts`
