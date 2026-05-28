# AWD Inspector UI Cluster Feature UI Normalization Plan

## 目标

- 把 `AWDRoundInspector` 及其直接依赖的 inspector UI cluster 迁入 `features/awd-inspector/ui`。
- 让 `AWDOperationsPanel.vue` 与 `ContestOperations.vue` 改为通过 `@/features/awd-inspector` public API 组合 inspector UI。
- 收掉 `architectureAllowlist.ts` 中与 `awd-inspector` 对应的历史例外。

## 非目标

- 本轮不迁 `AWDReadiness*`、`AWDInstanceOrchestrationPanel.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`。
- 本轮不改 `features/awd-inspector/model/*` 的导出、筛选、格式化、traffic summary 或下载逻辑。
- 本轮不重排 `ContestOperations.vue` 与 `AWDOperationsPanel.vue` 的 workflow owner。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`
- `code/frontend/src/components/platform/contest/AWDTrafficPanel.vue`
- `code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue`
- `code/frontend/src/components/platform/contest/AWDScoreboardSummaryPanel.vue`
- `code/frontend/src/components/platform/contest/AWDRoundHeaderPanel.vue`
- `code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue`
- `code/frontend/src/components/platform/contest/AWDServiceAlertBanner.vue`
- `code/frontend/src/components/platform/contest/awdInspector.types.ts`
- `code/frontend/src/features/awd-inspector/index.ts`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/views/platform/ContestOperations.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `features/awd-inspector/model/*` 已经是 inspector 的真实 owner。
- 目前 UI cluster 仍滞留在 `components/platform/contest/*`，并依赖 allowlist 才能访问 `awd-inspector` model。
- 最小正确落点是新增 `features/awd-inspector/ui/*` 并通过 feature public API 暴露。

## 设计边界

### `features/awd-inspector/ui/*` 本轮负责

- `AWDRoundInspector.vue`
- `AWDTrafficPanel.vue`
- `AWDServiceStatusPanel.vue`
- `AWDScoreboardSummaryPanel.vue`
- `AWDRoundHeaderPanel.vue`
- `AWDAttackLogPanel.vue`
- `AWDServiceAlertBanner.vue`
- `awdInspector.types.ts`
- `AWDOperationsPanel.vue` / `ContestOperations.vue` 对上述 cluster 的 public API 组合

### `features/awd-inspector/model/*` 本轮继续负责

- inspector 过滤、导出、派生、格式化、traffic 面板逻辑

### `features/contest-awd-admin/*` 本轮继续负责

- 运维主面板 owner
- readiness、instance orchestration、round create、service check、attack log dialog 等 admin runtime workflow

### `components/platform/contest/*` 本轮不再负责

- `AWDRoundInspector.vue`
- `AWDTrafficPanel.vue`
- `AWDServiceStatusPanel.vue`
- `AWDScoreboardSummaryPanel.vue`
- `AWDRoundHeaderPanel.vue`
- `AWDAttackLogPanel.vue`
- `AWDServiceAlertBanner.vue`
- `awdInspector.types.ts`

## 任务切片

### Slice 1：feature UI owner 迁位

- 目标：
  - 新增 `features/awd-inspector/ui/index.ts`
  - 迁入 inspector UI cluster
  - `features/awd-inspector/index.ts` 暴露 UI public API
  - `AWDOperationsPanel.vue`、`ContestOperations.vue` 改为从 `@/features/awd-inspector` 引用 inspector UI
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/views/platform/__tests__/ContestOperations.test.ts`
- Review focus：
  - `contest-awd-admin` 是否继续只组合 inspector panel，而不重新吸入 inspector 逻辑 owner
  - `ContestOperations.vue` 是否继续只保留 route shell 与插槽装配

### Slice 2：护栏与 allowlist 同步

- 目标：
  - 更新 `components.d.ts`
  - 删除 `architectureAllowlist.ts` 中 3 条 `awd-inspector` 历史例外
  - 更新 raw-source / extraction / theme / surface 测试
  - backlog 记录 awd inspector feature UI 收口进展
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - touched surface 是否已不再依赖旧 `components/platform/contest/AWDRoundInspector*` 路径
  - allowlist 是否在 touched surface 内被实际收掉

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只处理 inspector UI cluster owner，不继续下钻 readiness / instance orchestration / dialogs 的 owner；后续如果继续清 AWD runtime cluster，需要再按 capability 单独切片。
