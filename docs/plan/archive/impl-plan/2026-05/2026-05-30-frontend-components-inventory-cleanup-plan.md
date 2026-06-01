> 状态：Current
> 事实源：当前 `components/*` 存量分类清单整理
> 替代：无

# Frontend Components Inventory Cleanup Plan

## 目标

- 基于当前代码事实整理 `components/*` 存量清理清单
- 把 `components` 存量按 `保留在 components` / `迁入 features` / `收进 widgets` 分类
- 为后续“彻底迁移”提供明确入口，而不是继续逐页碰运气
- 明确 route entry 最终统一收敛到 `code/frontend/src/pages/`，不再让 `features/*RoutePage.vue` 或 `widgets/*RoutePage.vue` 兼任页面层语义

## 非目标

- 本轮不重写与第 1 批无关的 `views` / `features` / `widgets` 现有实现

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `code/frontend/src/components/`
- `code/frontend/src/features/`
- `code/frontend/src/widgets/`
- `code/frontend/src/views/`

## 当前结论

- 前端已经不是“中间迁移态”，但仍保留大批历史业务组件目录。
- 当前更需要的是显式存量清单，而不是继续零碎推进单个页面。
- 真正的“彻底迁移”目标，应该是逐步清空 `components/platform|teacher|contests|dashboard|challenge|scoreboard|notifications` 这类历史业务目录中的单一 feature UI，只留下 `components/common` / `components/layout` 和少数跨 feature 中立展示块。
- 第 1 批平台 7 组组件已经可以直接迁移，不需要保留任何桥接状态。
- 管理员端平台能力型 feature 已统一落到 `features/platform/*` 命名空间，后续不再新增或保留 `platform-*` 扁平 feature 目录。
- 运行时 `views/*.vue` 不再保留。
- route entry 的最终物理层是 `code/frontend/src/pages/`：
  - `pages/**` 只负责路由入口、参数接线、redirect 组合和页面编排。
  - `features/**` 继续持有单一能力 owner。
  - `widgets/**` 继续持有跨 feature 页面区块组合，不再兼任 route page 语义。
- 本轮开始把现有 `features/**/ui/*RoutePage.vue` 与 `widgets/**/**/*RoutePage.vue` 迁到 `pages/**`，并通过前端架构守卫禁止回流。

## 输出

- 在现有前端 backlog 中新增一节 `components/*` 存量清理清单
- 在前端架构事实与守卫中显式新增 `pages/` 层，并禁止新增非 `pages/**` 的 `*RoutePage.vue`
- 完成第 1 批平台 7 组历史业务组件迁移：
  - `platform/dashboard/*` -> `features/platform/overview/ui`
  - `platform/user/*` -> `features/platform/user-management/ui`
  - `platform/class/*` -> `features/platform/class-management/ui`
  - `platform/student/*` -> `features/platform/student-management/ui`
  - `platform/instance/*` -> `features/platform/instance-management/ui`
  - `platform/audit/*` -> `features/audit-log/ui`
  - `platform/images/*` -> `features/image-management/ui`
- 补齐平台管理员端 feature 命名空间迁移：
  - `features/platform-overview` -> `features/platform/overview`
  - `features/platform-user-management` -> `features/platform/user-management`
  - `features/platform-class-management` -> `features/platform/class-management`
  - `features/platform-student-management` -> `features/platform/student-management`
  - `features/platform-instance-management` -> `features/platform/instance-management`
  - `features/platform-challenges` -> `features/platform/challenges`
  - `features/platform-challenge-detail` -> `features/platform/challenge-detail`
  - `features/platform-contests` -> `features/platform/contests`
  - `features/platform-awd-challenges` -> `features/platform/awd-challenges`
- 补齐 route page 彻底迁移：
  - 历史 `views/*.vue` 入口先迁出运行时路径
  - 现有 `features/**/ui/*RoutePage.vue`、`widgets/**/**/*RoutePage.vue` 再统一迁到 `pages/**`
  - `router/routes/*.ts` 只再指向 `pages/**`
- 第 1 批 pages 迁移面：
  - `TeacherDashboardRoutePage`
  - `PlatformOverviewRoutePage`
  - `StudentReviewArchiveRoutePage`
  - `PlatformAwdReviewIndexRoutePage`
  - `TeacherAwdReviewIndexRoutePage`
  - `ChallengeImportPreviewRoutePage`
  - `ChallengeTopologyStudioRoutePage`
  - `AwdChallengeImportRoutePage`
  - `ChallengeWriteupRoutePage`
  - `ChallengeWriteupViewRoutePage`
- 同步更新 route view、测试和 feature public API，确保不保留中间桥。

## 验证

- `bash scripts/check-task-intake.sh --reuse-decision frontend-components-inventory-cleanup`
- `git diff --check`
- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/ImageManage.test.ts`
- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`

## 残余风险

- 该清单是当前事实快照；后续如果 `components/*` 继续变化，需要同步更新。
