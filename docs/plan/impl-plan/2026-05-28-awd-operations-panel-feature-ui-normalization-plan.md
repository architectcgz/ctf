> 状态：Current
> 事实源：`AWDOperationsPanel.vue` 当前 owner、`contest-awd-admin` 既有 model owner 与 allowlist 现状
> 替代：无

# AWD Operations Panel Feature UI Normalization Plan

## 目标

- 把 `AWDOperationsPanel.vue` 迁入 `features/contest-awd-admin/ui`。
- 让 `ContestOperations.vue` 改为通过 `contest-awd-admin` public API 组合 panel。
- 收掉 `architectureAllowlist.ts` 中这条 `AWDOperationsPanel.vue -> @/features/contest-awd-admin` 历史例外。

## 非目标

- 本轮不迁 `AWDRoundInspector.vue`、`AWDReadinessSummary.vue`、`AWDServiceCheckDialog.vue` 等被 panel 组合的子件。
- 本轮不改 `usePlatformContestAwd()` 的数据、轮次、流量、实例编排或 readiness owner。
- 本轮不重排 `ContestOperations.vue` 的 route shell 样式和告警插槽行为。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDOperationsPanel.vue`
- `code/frontend/src/views/platform/ContestOperations.vue`
- `code/frontend/src/features/contest-awd-admin/index.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/ContestOperations.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `AWDOperationsPanel.vue` 是 `contest-awd-admin` 的 feature-owned UI，不是 shared platform contest primitive。
- `ContestOperations.vue` 只是 route shell，不应该继续直接引用旧 contest component 路径。
- 最小正确落点是 `features/contest-awd-admin/ui/AWDOperationsPanel.vue`。

## 设计边界

### `features/contest-awd-admin/ui/*` 本轮负责

- AWD operations panel
- contest operations route shell 对 panel 的 public API 组合

### `features/contest-awd-admin/model/*` 本轮继续负责

- `usePlatformContestAwd()` 及其组合的 round/readiness/traffic/instance orchestration owner

### `components/platform/contest/*` 本轮继续负责

- `AWDRoundInspector.vue`
- `AWDReadinessSummary.vue`
- `AWDServiceCheckDialog.vue`
- 其它被 panel 组合的 AWD runtime primitive / dialog / subpanel

## 任务切片

### Slice 1：panel owner 迁位

- 目标：
  - 新增 `features/contest-awd-admin/ui/AWDOperationsPanel.vue`
  - `ContestOperations.vue` 改为从 `@/features/contest-awd-admin` public API 引 panel
  - 新增 `features/contest-awd-admin/ui/index.ts`
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/ContestOperations.test.ts`
- Review focus：
  - panel 是否仍直接消费 `usePlatformContestAwd()`
  - route shell 是否仍只负责 contest 查询和插槽组合

### Slice 2：护栏与 allowlist 同步

- 目标：
  - 更新 `components.d.ts`
  - 删除 `architectureAllowlist.ts` 中的这条历史例外
  - 更新 raw-source / extraction 测试
  - backlog 记录 AWD operations feature UI 收口进展
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/platform/contest/AWDOperationsPanel.vue` 路径是否已经从 touched surface 消失
  - allowlist 是否在 touched surface 内被实际收掉

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只处理 `AWDOperationsPanel.vue` 的 feature owner 迁位，不继续下钻到它组合的子件；如果后续要继续清 AWD runtime cluster，需要按 `readiness / round inspector / instance orchestration` 这类 capability 再分刀。
