> 状态：Current
> 事实源：`architectureAllowlist.ts`、对应 feature index / ui / model、相关 view 入口
> 替代：无

# Allowlist A Feature UI Batch One Plan

## 目标

- 收口 `A 档` 第一批 6 条 `componentFeatureImportAllowlist`，把明显属于单一 feature owner 的 UI 从 `components/*` 迁入对应 feature 的 `ui` 目录。
- 保持现有页面行为、feature model、路由、API 调用和测试语义不变，只调整 UI owner 落点与 public API 出口。

## 非目标

- 本轮不处理 `challenge-detail` 与 `contest-awd-workspace` 的其余 `A 档` 条目；仅在 `ScoreboardRealtimeBridge` 迁移后，为消除新的 component→feature 违规，顺带迁移 `ContestAWDWorkspacePanel.vue` 这个唯一 legacy component consumer。
- 本轮不处理 topology editor、layout shell 这些需要先重定边界的 `B 档` 条目。
- 本轮不改功能逻辑、路由结构、权限语义或 API contract。

## 输入依据

- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `code/frontend/src/components/scoreboard/ScoreboardRealtimeBridge.vue`
- `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`
- `code/frontend/src/components/platform/user/PlatformUserFormDialog.vue`
- `code/frontend/src/components/platform/challenge/ChallengeManageDirectoryPanel.vue`
- `code/frontend/src/components/teacher/InterventionPanel.vue`
- `code/frontend/src/components/teacher/reports/ClassReportExportDialog.vue`
- 对应 feature 目录：
  - `features/scoreboard`
  - `features/admin-notification-publisher`
  - `features/platform-user-management`
  - `features/platform-challenges`
  - `features/teacher-student-analysis`
  - `features/teacher-class-report-export`

## 当前结论

- 这 6 个组件都直接消费单一 feature model / workflow，没有 layout infra 或 topology editor 那种多边界混合问题。
- 如果继续把它们留在 `components/*`，`componentFeatureImportAllowlist` 只是把“feature-owned UI 还没迁完”冻结成长期例外。
- 最小正确改动是迁入 feature `ui`，再让 view / page / barrel 从 feature public API 组合，不去再包一层历史路径兼容壳。
- `ScoreboardRealtimeBridge` 原先还有一个 legacy component consumer `ContestAWDWorkspacePanel.vue`；若只迁 bridge，不迁这个 panel，就会立刻新增一条 `components/** -> @/features/scoreboard` 违规，因此本轮把它视为伴随收口面。

## 设计边界

### 本轮负责

- 新建或补齐对应 feature 的 `ui/index.ts` 与 public API
- 把 6 个组件迁入 feature `ui`
- 更新直接 import 这些组件的 view / feature / test / barrel
- 收掉对应 `componentFeatureImportAllowlist` 条目
- 更新 backlog 与 review 证据

### 本轮不动

- 组件内部的业务交互逻辑与样式策略
- 组件依赖的 feature model
- route view 结构
- 其它 allowlist 类型

## 任务切片

### Slice 1：bridge / drawer / dialog 三件套迁入 feature ui

- 目标：
  - 迁 `ScoreboardRealtimeBridge`、`AdminNotificationPublishDrawer`、`PlatformUserFormDialog`
- 预期改动：
  - `features/scoreboard/index.ts`
  - `features/scoreboard/ui/index.ts`
  - `features/scoreboard/ui/ScoreboardRealtimeBridge.vue`
  - `features/admin-notification-publisher/index.ts`
  - `features/admin-notification-publisher/ui/index.ts`
  - `features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue`
  - `features/platform-user-management/index.ts`
  - `features/platform-user-management/ui/index.ts`
  - `features/platform-user-management/ui/PlatformUserFormDialog.vue`
  - 对应 view / test import
- review focus：
  - public API 是否清楚
  - 是否引入新的跨层依赖
  - `ScoreboardRealtimeBridge` 迁移后是否还残留 component 层 consumer

### Slice 2：challenge / teacher feature-owned panel 迁入 feature ui

- 目标：
  - 迁 `ChallengeManageDirectoryPanel`、`InterventionPanel`、`ClassReportExportDialog`
- 预期改动：
  - `features/platform-challenges/index.ts`
  - `features/platform-challenges/ui/index.ts`
  - `features/platform-challenges/ui/ChallengeManageDirectoryPanel.vue`
  - `features/teacher-student-analysis/index.ts`
  - `features/teacher-student-analysis/ui/index.ts`
  - `features/teacher-student-analysis/ui/InterventionPanel.vue`
  - `features/teacher-class-report-export/index.ts`
  - `features/teacher-class-report-export/ui/index.ts`
  - `features/teacher-class-report-export/ui/ClassReportExportDialog.vue`
  - `components/reports/index.ts`
  - `components/teacher/reports/index.ts`
  - 对应 page / test import
- review focus：
  - teacher / platform 共享入口是否仍稳定
  - `components/reports` 公共出口是否只保留 barrel 责任

### Slice 3：allowlist / 测试 / backlog 收尾

- 目标：
  - 删除这 6 条 allowlist，更新原始源码测试与 backlog 记录
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - 受影响 raw-source / extraction / surface 测试
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-allowlist-a-feature-ui-batch-one-review.md`
- review focus：
  - allowlist 是否真的减少
  - 没有残留旧路径 import

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/notifications/__tests__/*.test.ts src/views/platform/__tests__/UserManage.test.ts src/views/platform/__tests__/ChallengeManage.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/components/notifications/__tests__/AdminNotificationPublishDrawer.test.ts src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `challenge-detail`、`contest-awd-workspace` 剩余 6 条、`topology`、`layout` 仍在后续批次，不在本轮收口。
