# Reuse Decision

## Change type
+test / cleanup / docs

## Existing code searched
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDReadinessSummary.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformPaginationControls.test.ts`
- `code/frontend/src/components/common/__tests__/PagePaginationControls.test.ts`
- `code/frontend/src/features/awd-readiness/ui/*`
- `code/frontend/src/features/awd-inspector/ui/*`
- `code/frontend/src/features/contest-workbench/ui/*`
- `code/frontend/src/features/platform/contests/ui/*`

## Similar implementations found
- 这批文件仍然是活动 Vitest 测试，但目录落点已经过时；它们测试的对象都已经迁到 `features/*` 或 `components/common/*`。
- `PlatformPaginationControls.test.ts` 与 `components/common/__tests__/PagePaginationControls.test.ts` 测试的是同一个共享组件，适合合并到共享测试，而不是继续保留平台专属副本。
- 其余测试分别对应明确的 feature UI owner，适合迁到对应 feature 邻近位置，不再挂在 `components/platform/__tests__/`。

## Decision
refactor_existing

## Reason
- 这次不是删除测试覆盖，而是把“测试还活着但路径已经错位”的残余归位。
- 测试应该跟随当前 owner，避免 `components/platform/__tests__/*` 继续制造旧架构入口假象。
- 对共享组件 `PagePaginationControls`，最小正确动作是合并断言到共享测试并删除错位副本。

## Files to modify
- `.harness/reuse-decisions/platform-legacy-test-shell-cleanup.md`
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDReadinessSummary.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDRoundInspector.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/ContestWorkbenchStageTabs.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- `code/frontend/src/components/platform/__tests__/PlatformPaginationControls.test.ts`
- `code/frontend/src/components/common/__tests__/PagePaginationControls.test.ts`
- `code/frontend/src/features/awd-readiness/ui/AWDReadinessSummary.test.ts`
- `code/frontend/src/features/awd-inspector/ui/AWDRoundInspector.test.ts`
- `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.test.ts`
- `code/frontend/src/features/contest-workbench/ui/ContestChallengeEditorDialog.test.ts`
- `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.test.ts`
- `code/frontend/src/features/contest-workbench/ui/ContestWorkbenchStageTabs.test.ts`
- `code/frontend/src/features/platform/contests/ui/AWDChallengeConfigPanel.test.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.test.ts`

## After implementation
- `components/platform/__tests__/` 不再承载这批业务测试。
- 共享组件测试保留在 `components/common/__tests__/`。
- feature 业务测试跟随当前 `features/*/ui/*` owner 落位。
