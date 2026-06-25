# 前端路由设计

> 状态：Current
> 事实源：`code/frontend/src/router/`、`code/frontend/src/config/backofficeNavigation.ts`、`code/frontend/src/utils/roleRoutes.ts`
> 替代：无

## 定位

本文档只说明前端页面路由的注册方式、命名空间、权限守卫、错误页回退和导航匹配规则。

- 覆盖：`Vue Router` 路由树、`/academy/*` 与 `/platform/*` 命名空间、登录态恢复、角色校验、错误状态页和默认首页映射。
- 不覆盖：页面内部的 tab/query 状态、列表筛选、异步数据加载和业务流程编排；这些能力由 `code/frontend/src/features/**/model` 或 `code/frontend/src/shared/model/navigation/*` 等共享 model owner 负责。

## 当前设计

- `code/frontend/src/router/index.ts`、`code/frontend/src/router/routes/appShellRoute.ts`
  - 负责：组装认证路由、主应用壳、错误页和工具页；在 `/` 下挂载学生端、教师端和平台端子路由，并把根路径重定向到 `/student/dashboard`
  - 不负责：承载页面内的数据请求、筛选状态或业务动作

- `code/frontend/src/router/routes/studentRoutes.ts`、`teacherRoutes.ts`、`platformRoutes.ts`
  - 负责：声明各工作区的正式 URL、`pages/**` 组件入口，以及 `meta.requiresAuth / meta.roles / meta.title / meta.icon`
  - 不负责：根据角色动态拼接第二套路由树，也不在页面组件里复制权限判断

- `code/frontend/src/router/guards.ts`
  - 负责：公开页放行、登录态恢复、未登录跳转 `/login?redirect=...`、角色不匹配跳 `/403`、异常时登出并回到登录页、`afterEach` 更新标题、`router.onError` 跳 `/500`
  - 不负责：显示复杂业务降级 UI；守卫只做导航级回退

- `code/frontend/src/config/backofficeNavigation.ts`、`code/frontend/src/utils/roleRoutes.ts`、`code/frontend/src/utils/routeTitle.ts`
  - 负责：教师/管理员导航高亮、详情页回归所属目录项、角色默认首页和页面标题解析
  - 不负责：替代路由注册本身，也不决定页面内的 tab 或二级状态

- `code/frontend/src/shared/lib/navigation/routeTarget.ts`
  - 负责：共享导航表面对 `vue-router` `RouteLocationRaw` 的薄契约别名，给 `AppRouteLink`、`AppRouteRedirect` 和 feature/page props 提供统一 target 类型
  - 不负责：持有 `push / replace / useRoute / useRouter` 等运行时路由 owner，也不承载业务跳转策略

## 1. 运行入口

路由注册从 `code/frontend/src/router/index.ts` 开始：

1. 挂载 `authRoutes`
2. 挂载 `appShellRoute`
3. 挂载 `errorRoutes`
4. 挂载 `utilityRoutes`
5. 在 `setupRouterGuards(router)` 中接入全局守卫

`appShellRoute` 当前是整个受保护应用的唯一壳：

- 路径：`/`
- 组件：`@/shared/ui/layout/AppLayout.vue`
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

### 3.1 守卫执行顺序

`code/frontend/src/router/guards.ts` 当前守卫顺序如下：

1. 命中公开页 `/login` 或 `/register` 时直接放行
2. 对公开页也会尝试 `ensureSessionRestored()`，已登录用户访问登录/注册页时，先通过 `sanitizeRedirectPath()` 做 open redirect 防御和 legacy `/teacher/* -> /academy/*` 归一化，再按 `redirect` 参数或 `getRoleDashboardPath()` 跳转
3. 受保护页面在首次进入时调用 `authStore.restore()`
4. 恢复后仍未登录时，跳转到 `/login?redirect=${to.fullPath}`
5. `meta.roles` 与当前用户角色不匹配时，提示”无权限访问该页面”并跳 `/403`
6. 守卫内部出现异常时，执行 `authStore.logout()` 并跳回登录页
7. `afterEach` 调用 `resolveRouteTitle()`，按 `APP_TITLE_PREFIX` 更新页面标题
8. `router.onError` 统一跳到 `/500`

**代码位置**：
- `code/frontend/src/router/guards.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`

### 3.2 Session 恢复时机

**恢复策略**：
- 公开页（`/login`、`/register`）也会执行 session 恢复
- 恢复通过 `ensureSessionRestored()` 函数统一处理
- 恢复只执行一次，通过 `authStore.sessionRestored` 标志位避免重复调用

**实现细节**：
```typescript
async function ensureSessionRestored(): Promise<void> {
  const authStore = useAuthStore()
  if (authStore.user || authStore.sessionRestored) return
  await authStore.restore()
}
```

**恢复时机**：
- 首次访问任何页面时（包括公开页和受保护页）
- 刷新页面后重新进入应用时
- 从外部链接或书签直接跳转时

**注意事项**：
- 恢复失败不阻塞公开页访问
- 恢复成功后，已登录用户访问 `/login` 会自动跳转到 `redirect` 参数或角色默认首页

### 3.3 Redirect 路径清洗

**功能定位**：防止 open redirect 攻击和兼容 legacy 路径。

**实现函数**：`sanitizeRedirectPath(redirectParam: string | string[] | undefined): string`

**代码位置**：`code/frontend/src/utils/redirectPath.ts`

**清洗规则**：
1. **参数规范化**：
   - `redirectParam` 为数组时取第一个元素
   - `redirectParam` 为空或非字符串时返回 `/`

2. **外部 URL 拦截**：
   - 包含 `://` 或 `//` 的 URL 视为外部链接，返回 `/`
   - 防止 `?redirect=https://evil.com` 跳转到外部站点

3. **Legacy 路径归一化**：
   - `/teacher/*` 路径统一返回 `/`（不再支持教师端旧路径）
   - 由登录后逻辑重新导向角色默认首页

4. **合法路径透传**：
   - 其他以 `/` 开头的路径直接返回
   - 包括 `/academy/*`、`/platform/*`、`/student/*` 等正式路径

**示例**：
```typescript
sanitizeRedirectPath('https://evil.com')       // → '/'
sanitizeRedirectPath('//evil.com')             // → '/'
sanitizeRedirectPath('/teacher/classes')       // → '/'
sanitizeRedirectPath('/academy/overview')      // → '/academy/overview'
sanitizeRedirectPath(['/contests', '/other'])  // → '/contests'
sanitizeRedirectPath(undefined)                // → '/'
```

### 3.4 403 跳转策略

**触发条件**：
- 用户已登录，但 `meta.roles` 不包含当前用户角色
- 例如：学生访问 `/platform/challenges`（仅 admin/teacher 可访问）

**执行流程**：
1. `hasRequiredRole()` 函数判断角色匹配
2. 不匹配时调用 `toast.warning('无权限访问该页面')`
3. 跳转到 `/403` 错误页

**实现细节**：
```typescript
export function hasRequiredRole(
  requiredRoles: RouteLocationNormalized['meta']['roles'],
  currentRole: UserRole | undefined
): boolean {
  if (!requiredRoles || requiredRoles.length === 0) return true
  if (!currentRole) return false
  return requiredRoles.includes(currentRole)
}
```

**注意事项**：
- `/403` 页面本身不要求登录
- 用户可从 `/403` 页面返回上一页或回到首页
- 权限校验只在路由层执行一次，页面内部不重复判断

### 3.5 异常处理与回退

**守卫异常**：
- `beforeEach` 守卫内部捕获所有异常
- 异常时执行 `authStore.logout()` 清除登录态
- 跳转到 `/login` 页面

**路由错误**：
- `router.onError()` 统一捕获路由加载失败
- 跳转到 `/500` 错误页

**测试覆盖**：
- `code/frontend/src/router/__tests__/guards.test.ts` 覆盖：
  - 公开页放行
  - Session 恢复时机
  - Redirect 参数处理
  - 角色权限校验
  - 异常回退流程

相关路径：

- 登录态恢复：`code/frontend/src/stores/auth.ts`
- 默认首页映射：`code/frontend/src/utils/roleRoutes.ts`
- 标题解析：`code/frontend/src/utils/routeTitle.ts`
- Redirect 清洗：`code/frontend/src/utils/redirectPath.ts`

## 4. 导航匹配与边界

当前导航匹配不依赖“页面自己知道应该高亮哪个菜单”，而是统一交给 `code/frontend/src/config/backofficeNavigation.ts`。

- 教师与管理员详情页会回映到所属目录项，而不是把每个详情页单独做成导航入口。
- `getBackofficeModuleByPath()` 负责把 `/academy/classes/:className/students/:studentId/review-archive` 这类详情页映射回 `operations` 模块。
- `getVisibleBackofficeSecondaryItems()` 负责根据角色裁剪可见目录项，并保持当前项的 active 状态。

这层规则的直接目的，是避免在 `.vue` 页面里重复写导航归属判断。

## 5. 兼容与历史例外

- `/teacher/*` 已不再注册为 router runtime page route，登录 redirect 参数也不再对它做 canonicalize；正式事实源只认 `/academy/*`。
- 遇到旧教师端页面路径 redirect 参数时，`sanitizeRedirectPath()` 会直接回退到 `/`，再由登录 redirect fallback 导向角色默认首页。
- `/dashboard`、`/instances`、`/skill-profile` 当前仍保留 redirect，属于学生端早期路径兼容。
- `resolveRouteTitle()` 对 `/dashboard` 和 `/student/dashboard` 做了特例处理，允许通过 query/变体路由生成不同标题。

## 6. Guardrail

- route page 入口只能落在 `pages/**`，且不能直接持有业务 API、路由状态或 query-tab 逻辑：`code/frontend/src/__tests__/architectureBoundaries.test.ts`、`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- legacy 教师端前端页面路径不再允许出现在活跃前端源码里，且 router runtime 不再注册 `/teacher/*` 页面 route：`code/frontend/src/__tests__/architectureBoundaries.test.ts`、`code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- 教师/管理员导航映射和命名空间匹配：`code/frontend/src/config/__tests__/backofficeNavigation.test.ts`
- 默认首页与角色跳转：`code/frontend/src/utils/roleRoutes.ts`
