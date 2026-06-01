# 前端路由入口收口与迁移台账更新计划

## Objective

- 把前端迁移记录从过时的 `views/composables` 叙述收口成当前事实。
- 先完成一条最小结构收敛切片：`platform/challenges` 路由不再直接指向 `features/ui`，统一改回 `pages` 入口。
- 补一条机械 guardrail，防止 router 再次直接挂到 `features`。

## Non-goals

- 不在本轮继续拆学生侧通知、竞赛列表等厚路由页。
- 不在本轮清理教师侧 `/teacher/*` 兼容 redirect。
- 不在本轮扩展 `entities` 或 `widgets` 的业务范围。

## Source Inputs

- `AGENTS.md`
- `docs/文档规范.md`
- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/scripts/frontend-architecture-policy.json`

## Task Slices

### Slice 1: 更新迁移台账

- 目标：把 `TODO/frontend-sliced-architecture.md` 改写成当前状态、剩余问题和推荐顺序，不再保留已经失效的 `views/composables` 叙述。
- 变更面：
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 文档如果继续保留旧目录名，会误导后续迁移判断。
- 验证：
  - 人工检查文档内容与当前代码结构一致。

### Slice 2: 收口平台题目管理路由入口

- 目标：新增 `pages` 层路由页，让 `platform/challenges` 不再直接 import `features/platform/challenges/ui/ChallengeManagePage.vue`。
- 变更面：
  - `code/frontend/src/pages/platform/challenges/ChallengeManageRoutePage.vue`
  - `code/frontend/src/router/routes/platformRoutes.ts`
- 风险：
  - 路由入口改动如果页面壳层漏透传，会导致展示或交互回退。
- 验证：
  - 题目管理相关单测。

### Slice 3: 加硬架构边界

- 目标：补测试，限制 router 运行时组件只能从 `pages` 层加载，允许的例外只保留 app shell 布局。
- 变更面：
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 风险：
  - 规则写太死会误伤 app shell 等合法入口。
- 验证：
  - 前端架构边界测试。

## Dependencies

- Slice 2 依赖 Slice 1 不强，但都依赖当前架构判断。
- Slice 3 依赖 Slice 2，先确认目标入口形态，再收紧 guardrail。

## Validation Plan

- `npm run test:run -- src/features/platform/challenges/__tests__/ChallengeManagePage.test.ts src/__tests__/architectureBoundaries.test.ts`

## Review Focus

- 路由入口是否统一经过 `pages` 层。
- 文档是否仍残留 `src/views`、`src/composables` 等已退场事实。
- 新 guardrail 是否能覆盖这次发现的绕行入口，而不误伤 `appShellRoute.ts`。

## Rollback / Recovery

- 若路由页引入后出现题目管理页回归，可先回退 `platformRoutes.ts` 对应入口，再保留文档和测试修订单独评估。
