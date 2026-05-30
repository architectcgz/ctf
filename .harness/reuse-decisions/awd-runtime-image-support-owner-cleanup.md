# Reuse Decision

## Change type
+component / cleanup / docs

## Existing code searched
- `code/frontend/src/components/platform/awdRuntimeImageSupport.ts`
- `code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeImportSection.vue`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue`
- `code/frontend/src/features/platform/contests/ui/AWDChallengeConfigPanel.vue`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDReadinessSummary.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`

## Similar implementations found
- 当前前端没有任何 consumer 再 import `awdRuntimeImageSupport.ts`，说明这不是“待迁移 owner”，而是已经退出主路径但尚未删除的孤立残片。
- `AwdChallengeImportSection.vue` 目前只保留镜像来源/状态/目标引用的简单展示，没有再使用这组 placeholder 或 runtime error 格式化 helper。

## Decision
refactor_existing

## Reason
- 这次不做“迁到新目录继续空挂”的假收口。
- 既然全仓已无消费方，最小正确动作就是删除 dead file，并让前端结构继续朝“没有中间迁移残片”的方向收口。

## Files to modify
- `.harness/reuse-decisions/awd-runtime-image-support-owner-cleanup.md`
- `code/frontend/src/components/platform/awdRuntimeImageSupport.ts`

## After implementation
- 如果后续又需要同类 runtime image placeholder / preview error 格式化能力，应当在真实消费 feature 内重新建立 owner，而不是恢复到 `components/platform/*` 历史目录。
