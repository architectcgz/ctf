> 状态：Current
> 事实源：`PlatformContestTable.vue` 当前 owner、`platform-contests` 既有 feature-owned UI 收口模式
> 替代：无

# Platform Contest Table Feature UI Normalization Plan

## 目标

- 把 `PlatformContestTable.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestOrchestrationPage.vue` 改为 feature 内部相对 import。
- 同步更新 feature UI 聚合导出、组件声明、raw-source 测试和 backlog 记录。

## 非目标

- 本轮不拆 `PlatformContestTable.vue` 内部列定义、状态 badge 或菜单逻辑。
- 本轮不顺手处理不存在的 `PlatformContestSummaryStrip.vue` 候选。
- 本轮不改 contest list 的分页、筛选、路由和数据 owner。

## 输入依据

- `code/frontend/src/components/platform/contest/PlatformContestTable.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform-contests/ui/index.ts`
- `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `PlatformContestTable.vue` 只服务 `ContestOrchestrationPage.vue` 这条 feature 目录页，不是 shared platform table。
- 最小正确落点是 `features/platform-contests/ui/PlatformContestTable.vue`。
- 这轮主要是 owner 迁移和引用面同步，不涉及 table workflow owner 变化。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- contest list table
- contest orchestration page 对 table 的 feature 内部组合

### `features/platform-contests/model/*` 本轮继续负责

- contest list 数据获取、分页、筛选、创建 / 编辑 / 通知 / 跳转 action owner

### `views/platform/ContestManage.vue` 本轮不负责

- 这轮不直接触达 route shell；list surface owner 继续由 `ContestOrchestrationPage.vue` 组合

## 任务切片

### Slice 1：table owner 迁位

- 目标：
  - 新增 `features/platform-contests/ui/PlatformContestTable.vue`
  - `ContestOrchestrationPage.vue` 改为 feature 内部相对 import
  - `features/platform-contests/ui/index.ts` 补充 table export
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/PlatformContestTable.test.ts`
- Review focus：
  - table props / emits contract 是否保持不变
  - page owner 是否仍只保留事件桥接和目录组合

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - 更新 raw-source / typography / surface alignment 测试引用路径
  - backlog 记录这条 `platform-contests` 残余 surface 的继续收口进展
- 验证：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/platform/contest/PlatformContestTable.vue` 路径是否已经从 touched surface 消失
  - feature UI 聚合出口是否继续完整

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/PlatformContestTable.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `PlatformContestTable.vue` 迁位后仍然是较大的目录表格组件；如果后续继续叠加更多筛选、批量操作或多态列渲染，应按目录页 capability 拆成更细的 table section 或 row action owner，而不是重新回到 `components/`。
