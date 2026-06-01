> 状态：Current
> 事实源：student dashboard overview owner 收口
> 替代：无

# Student Overview Feature Owner Cleanup Plan

## 目标

- 把 `StudentOverviewStyleEditorial.vue` 与 `overviewProps.ts` 从历史 `components/dashboard/student` 收口到 `features/student-dashboard/ui`。
- 清掉只剩转发用途的 `StudentOverviewVariantSwitcher.vue` 残留。
- 同步 raw-source 测试与自动组件声明，消除旧路径引用。

## 非目标

- 本轮不重做 student dashboard 视觉、交互或业务逻辑。
- 本轮不扩展到 `TrainingTimelinePanel`、difficulty、category、recommendation 子块。
- 本轮不处理通用图表、按钮或 journal 样式 token。

## 输入依据

- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue`
- `code/frontend/src/components/dashboard/student/overviewProps.ts`
- `code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue`
- `code/frontend/src/features/student-dashboard/ui/StudentOverviewPage.vue`
- student dashboard 相关 raw-source 样式测试

## 设计边界

### `features/student-dashboard/ui` 本轮负责

- student overview 展示块
- overview props 契约
- page 到展示块的直接装配

### 本轮继续不动

- dashboard route page 的数据 owner
- timeline / recommendation / category / difficulty 等其他 student dashboard 分块
- 通用 `RadarChart` 与共享样式定义

## 任务切片

### Slice 1：overview 组件 owner 迁移

- 目标：
  - 把 `StudentOverviewStyleEditorial.vue`、`overviewProps.ts` 迁到 `features/student-dashboard/ui`
  - 更新 `StudentOverviewPage.vue` 为 feature 内部依赖
- 验证：
  - `npm run test:run -- src/pages/dashboard/__tests__/DashboardView.test.ts`
- Review focus：
  - `StudentOverviewPage.vue` 是否仍只做轻量转发装配
  - props 契约是否和展示 owner 保持在同一 feature

### Slice 2：残留清理与护栏同步

- 目标：
  - 清理 `StudentOverviewVariantSwitcher.vue`
  - 更新 raw-source 测试与 `components.d.ts`
  - 确认源码里不再引用旧 `components/dashboard/student/*overview*`
- 验证：
  - `npm run test:run -- src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentUserSurfaceAlignment.test.ts src/pages/__tests__/journalUserShellStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/studentJournalSoftStyles.test.ts src/__tests__/studentJournalButtonStyles.test.ts`
  - `rg -n "components/dashboard/student/(StudentOverviewStyleEditorial|StudentOverviewVariantSwitcher|overviewProps)" code/frontend/src`
- Review focus：
  - raw-source 护栏是否继续覆盖新的 feature owner 路径
  - 旧目录是否真的不再承担 runtime owner

## 验证计划

- `bash scripts/check-task-intake.sh --reuse-decision student-overview-feature-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/dashboard/__tests__/DashboardView.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentUserSurfaceAlignment.test.ts src/pages/__tests__/journalUserShellStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/studentJournalSoftStyles.test.ts src/__tests__/studentJournalButtonStyles.test.ts`
- `cd /home/azhi/workspace/projects/.worktrees/ctf-awd-service-feature-owner-cleanup && git diff --check`

## 风险与回退

- 风险：
  - raw-source 测试仍引用旧路径，导致 owner 收口不完整。
  - 删除 `StudentOverviewVariantSwitcher.vue` 前若漏掉隐藏 consumer，会产生编译错误。
- 回退：
  - 若迁移后有路径或声明问题，可直接回退 overview 组件、props 契约与 `StudentOverviewPage.vue` 的 import 调整，不影响其他 dashboard 子块。
