# Reuse Decision

## Change type
frontend refactor / feature-owned AWD operations dialog cluster decomposition

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts
- docs/reviews/frontend/2026-05-29-awd-traffic-panel-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 最近的 `AWDServiceStatusPanel.vue`、`AWDTrafficPanel.vue`、`ClassReportExportDialog.vue` 都已经按“父层保留唯一 workflow / form owner，稳定 section 与 CSS 下沉，raw-source 护栏切到聚合源码”的模式完成收口。
- `contest-awd-admin` 当前剩余的这组 dialog 债也一致：`AWDRoundCreateDialog.vue`、`AWDAttackLogDialog.vue` 与 `AWDServiceCheckDialog.vue` 仍由父层同时持有 form / validation / submit owner、稳定表单 section 与 scoped CSS。

## Decision
refactor_existing

## Reason
`AWDRoundCreateDialog.vue`、`AWDAttackLogDialog.vue` 与 `AWDServiceCheckDialog.vue` 当前分别约 `254` / `378` / `326` 行，主要由：

- dialog shell 与 props / emits contract
- open watch 下的 form reset owner
- field validation、JSON parse、service id resolve 与 duplicate-action guard
- round settings / score fields
- 稳定 team / challenge / status / details section
- footer 按钮壳与 scoped CSS

组成。最小正确改动不是把逻辑上提回 `AWDOperationsDialogHub.vue`，也不是为每个字段单独造 composable，而是：

- 保持三个父 dialog 继续作为唯一 form、validation、submit 与 duplicate-action owner
- 为创建轮次提取 settings / score section
- 为攻击日志提取 team selection / details section
- 为服务检查提取 target / result section
- 把三者共用的 footer 壳、dialog 局部样式与 payload contract 收口到 feature 内共享 CSS / 子组件 / types
- 新增 dialog cluster extraction 护栏，并把 primitive / duplicate-action / backoffice dialog 护栏切到聚合源码视角

## Files to modify
- .harness/reuse-decisions/awd-operations-dialog-cluster-decomposition.md
- docs/plan/impl-plan/2026-05-29-awd-operations-dialog-cluster-decomposition-plan.md
- docs/reviews/frontend/2026-05-29-awd-operations-dialog-cluster-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateSettingsSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDRoundCreateScoreSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckDialog.vue
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogTargetSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDAttackLogDetailsSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckTargetSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDServiceCheckResultSection.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogFooter.vue
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsDialogs.css
- code/frontend/src/components/platform/__tests__/AWDOperationsDialogsExtraction.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts
- code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts
- code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts

## After implementation
- `AWDRoundCreateDialog.vue` 会回到“创建轮次 form / validation / submit owner”这一层，不再继续直接内联 round settings、score fields、footer 与 scoped CSS。
- `AWDAttackLogDialog.vue` 会回到“攻击日志 form / validation / submit owner”这一层，不再继续直接内联队伍选择、攻击细节、footer 与 scoped CSS。
- `AWDServiceCheckDialog.vue` 会回到“服务检查 form / validation / submit owner”这一层，不再继续直接内联目标选择、JSON 结果区、footer 与 scoped CSS。
- AWD operations dialogs 的 create round / service check / attack log payload shape 会先收口到 feature 内共享 contract owner，不再在各个 dialog 各自手写。
- 用户可见行为保持不变：轮次默认值、队伍/题目默认值、service id 推导、JSON 校验、重复提交短路、禁用态和提示文案都保持现状。
- backlog 中当前这条 contest / AWD feature 内残余大 surface，会把这组三个 dialog 从高优先 residual 里收掉，后续重点转到更深层 workflow / contract 清理。
