# Reuse Decision

## Change type
+component / feature / cleanup / docs

## Existing code searched
- `code/frontend/src/components/platform/contest/AWDRoundSelectionPanel.vue`
- `code/frontend/src/components/platform/contest/ContestOperationsTopbarPanel.vue`
- `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- `code/frontend/src/components/platform/contest/awdCheckerPreviewProgress.ts`
- `code/frontend/src/features/contest-awd-config/model/awdCheckerConfigSupport.ts`
- `code/frontend/src/features/contest-awd-config/model/useAwdCheckerPreview.ts`
- `code/frontend/src/features/awd-inspector/ui/AWDRoundHeaderPanel.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`

## Similar implementations found
- `awdCheckerConfigSupport.ts` 已经在 `features/contest-awd-config/model/` 下有真实 owner，旧 `components/platform/contest/awdCheckerConfigSupport.ts` 只是历史重复物。
- `AWDRoundSelectionPanel.vue` 的轮次切换逻辑已经在 `features/awd-inspector/ui/AWDRoundHeaderPanel.vue` 内有现行实现，旧文件没有运行时 consumer。
- `ContestOperationsTopbarPanel.vue` 已被页面与测试明确排除，不再属于当前运维页结构。
- `awdCheckerPreviewProgress.ts` 当前虽然没有 consumer，但它表达的是 checker preview workflow 的阶段与耗时规则；如果后续恢复试跑进度条/提示，它应落在 `features/contest-awd-config/model/`，而不是留在旧 `components/platform/contest/*`。

## Decision
refactor_existing

## Reason
- 这次不做“把残片迁到新目录继续保留”的假收口。
- 旧 `components/platform/contest/*` 当前更像历史迁移残留，其中一部分已经由 feature 内 owner 覆盖，另一部分已经退出运行时。
- 但 `awdCheckerPreviewProgress.ts` 仍然属于明确的 preview workflow 资产，只是还未重新接回现行实现；因此应迁到 `features/contest-awd-config/model/` 待后续复用，而不是继续挂在旧目录或直接删掉。
- 其余重复实现或废弃 UI 继续删除，同时把残余测试断言改到 feature owner。

## Files to modify
- `.harness/reuse-decisions/contest-legacy-owner-cleanup.md`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/components/platform/contest/AWDRoundSelectionPanel.vue`
- `code/frontend/src/components/platform/contest/ContestOperationsTopbarPanel.vue`
- `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- `code/frontend/src/components/platform/contest/awdCheckerPreviewProgress.ts`
- `code/frontend/src/features/contest-awd-config/model/awdCheckerPreviewProgress.ts`

## After implementation
- `contest AWD config` 只认 `features/contest-awd-config/model/awdCheckerConfigSupport.ts` 为唯一 owner。
- `awdCheckerPreviewProgress.ts` 迁到 `features/contest-awd-config/model/`，作为该 feature 内的 preview workflow 辅助资产保留。
- `components/platform/contest/*` 其余历史残片从主路径退出，不保留桥接壳。
- 如果后续确实重新需要 round selection 共享能力，应在真实消费 feature 内重新建立 owner，而不是恢复到旧 `components/platform/contest/*`。
