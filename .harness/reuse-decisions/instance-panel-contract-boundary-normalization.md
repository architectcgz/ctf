# Reuse Decision

## Change type
frontend refactor / common contract boundary normalization

## Existing code searched
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/common/InstancePanel.vue
- code/frontend/src/components/common/__tests__/InstancePanel.test.ts
- code/frontend/src/features/instance-list/model/useInstanceListPage.ts
- code/frontend/src/components/challenge/ChallengeInstanceCard.vue

## Similar implementations found
- 当前仓库已经多次把 shared / entity / feature UI 的展示契约从 `@/api/contracts` 收回到本地最小字段接口，避免基础层组件直接依赖 API DTO owner。
- `InstancePanel.vue` 这条例外与刚完成的 challenge entity 边界收口是同类问题：组件真实需要的只是少量展示字段和状态枚举，而不是完整 API contract。

## Decision
refactor_existing

## Reason
`commonForbiddenImportAllowlist` 目前只剩 1 条：

- `components/common/InstancePanel.vue -> @/api/contracts`

`InstancePanel.vue` 是 shared/common 组件，当前直接依赖 `InstanceListItem` 与 `InstanceStatus`，但它真正需要的只是：

- 实例卡片展示字段
- expiring soon 事件回传的同一最小字段
- 实例状态和分享范围的本地表示

最小正确改动是：

- 在 `components/common` 下为 `InstancePanel` 补本地最小类型
- 让组件改为消费本地 contract
- 删除 `commonForbiddenImportAllowlist` 中最后 1 条例外

不扩大到 instance feature / challenge detail 流程 owner，也不引入新的过渡 bridge。

## Files to modify
- .harness/reuse-decisions/instance-panel-contract-boundary-normalization.md
- docs/plan/impl-plan/2026-05-29-instance-panel-contract-boundary-normalization-plan.md
- docs/reviews/frontend/2026-05-29-instance-panel-contract-boundary-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components/common/InstancePanel.vue
- code/frontend/src/components/common/instancePanel.types.ts
- code/frontend/src/components/common/__tests__/InstancePanel.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## After implementation
- `InstancePanel.vue` 不再直接依赖 `@/api/contracts`。
- `commonForbiddenImportAllowlist` 清空。
- shared/common 层不保留这个历史 API contract 例外。
