> 状态：Current
> 事实源：`PlatformContestFormPanel.vue` / `PlatformContestFormDialog.vue` 当前 owner、`platform-contests` 既有 feature-owned UI 收口模式
> 替代：无

# Platform Contest Form Feature UI Normalization Plan

## 目标

- 把 `PlatformContestFormPanel.vue` 与 `PlatformContestFormDialog.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestManage.vue`、`ContestOrchestrationPage.vue`、`ContestEditWorkspacePanel.vue` 改为通过 feature 内部 UI 或 feature public API 组合。
- 收掉 `architectureAllowlist.ts` 中与 `PlatformContestForm*.vue` 对应的两条历史例外。

## 非目标

- 本轮不拆 `PlatformContestFormPanel.vue` 的内部 section。
- 本轮不修改 `useContestManagePage.ts`、`useContestEditPage.ts` 的数据/保存 owner。
- 本轮不顺手处理 `PlatformContestTable.vue` 或 `ContestAwdPreflightPanel.vue`。

## 输入依据

- `code/frontend/src/components/platform/contest/PlatformContestFormPanel.vue`
- `code/frontend/src/components/platform/contest/PlatformContestFormDialog.vue`
- `code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform-contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/views/platform/ContestManage.vue`
- `code/frontend/src/features/platform-contests/model/contestFormSupport.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `PlatformContestFormPanel.vue` 是 contest create / edit 的单一 feature 表单 UI，不是通用 backoffice shared form。
- `PlatformContestFormDialog.vue` 只是 `AdminSurfaceModal + PlatformContestFormPanel` 的壳，owner 也应跟随 `platform-contests`。
- 最小正确落点是 `features/platform-contests/ui/PlatformContestFormPanel.vue` 与 `features/platform-contests/ui/PlatformContestFormDialog.vue`。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- contest form panel
- contest form dialog
- contest create / edit 表单 UI 组合

### `features/platform-contests/model/*` 本轮继续负责

- `ContestFormDraft` / `ContestFieldLocks` / status option 等表单 model 类型和辅助逻辑
- contest create / edit / save workflow

### `views/platform/ContestManage.vue` 本轮继续负责

- page owner
- 对话框开关与保存事件桥接

## 任务切片

### Slice 1：form panel / dialog 落位

- 目标：
  - 新增 `features/platform-contests/ui/PlatformContestFormPanel.vue`
  - 新增 `features/platform-contests/ui/PlatformContestFormDialog.vue`
  - `ContestOrchestrationPage.vue`、`ContestEditWorkspacePanel.vue`、`ContestManage.vue` 改为走 feature 内部导入或 feature public API
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- Review focus：
  - dialog / panel 是否继续只保留 props / emits contract
  - route owner / save owner 是否没有回流到 UI 组件

### Slice 2：护栏与 raw-source 同步

- 目标：
  - 更新 `components.d.ts`
  - 删除 `architectureAllowlist.ts` 中的两条旧例外
  - 更新 raw-source / theme token / surface alignment 测试
  - backlog 记录下一批剩余候选
- 验证：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - feature UI 是否不再走 `components/platform/contest/*` 路径
  - allowlist 是否在 touched surface 内被实际收掉

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `PlatformContestFormPanel.vue` 迁位后仍然是较大的单文件表单；如果后续继续在这块表单追加新的运营/赛制编辑分区，再按 section 拆细，而不是重新把它压回 `components/`。
