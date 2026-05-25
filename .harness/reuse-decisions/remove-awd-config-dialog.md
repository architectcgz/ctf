# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase*.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`

## Similar implementations found
- `code/frontend/src/views/platform/ContestAwdConfig.vue`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- `code/frontend/src/components/platform/contest/ContestAwdDebugStation.vue`

## Decision
refactor_existing

## Reason
`AWDChallengeConfigDialog.vue` 已不再承担真实运行态入口，当前 AWD 配置 owner 已收口到独立页面 `ContestAwdConfig.vue` 与工作室阶段面板 `AWDChallengeConfigPanel.vue`。本次不创建新组件，而是删除历史遗留 dialog，并把类型声明、测试断言、allowlist 与文档事实同步到现有 owner。

## Files to modify
- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase*.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
- `docs/architecture/features/AWD检查器结构化编辑器设计.md`
- `docs/architecture/features/AWD检查器试跑设计.md`
- `docs/architecture/features/AWD检查器运行器扩展设计.md`
- `docs/architecture/features/AWD检查器校验状态设计.md`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如后续 AWD 配置 owner 边界需要复查，把结论沉淀到 `.harness/reuse-index/` 对应目录，而不是重新恢复 dialog 入口。
