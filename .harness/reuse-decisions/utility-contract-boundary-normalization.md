# Reuse Decision

## Change type
frontend refactor / utility contract boundary normalization

## Existing code searched
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/utils/contest.ts
- code/frontend/src/utils/platformContestAwdChallengeLinks.ts
- code/frontend/src/utils/skillProfile.ts
- code/frontend/src/api/assessment.ts
- code/frontend/src/api/teacher/students.ts
- code/frontend/src/api/teaching/students.ts
- code/frontend/src/features/contest-awd-admin/model/*.ts
- code/frontend/src/features/contest-workbench/model/*.ts

## Similar implementations found
- 当前仓库已经用 challenge entity / InstancePanel 这两刀证明：如果某层只需要最小展示字段和归一化语义，就应该把 contract 收到本地，而不是继续直接依赖 `@/api/contracts`。
- `utilityBoundaryImportAllowlist` 当前剩下的 3 条也属于同类问题，只是 owner 落在 `utils/*`。

## Decision
refactor_existing

## Reason
`utilityBoundaryImportAllowlist` 当前只剩 3 条：

- `utils/contest.ts -> @/api/contracts`
- `utils/platformContestAwdChallengeLinks.ts -> @/api/contracts`
- `utils/skillProfile.ts -> @/api/contracts`

这 3 个文件分别承担：

- contest 状态 / 模式的展示文案与颜色
- AWD service 到 contest challenge link 的本地映射
- skill profile / recommendation 的归一化与展示辅助

它们都不需要直接持有 API contract owner。最小正确改动是：

- 在各自 utility 文件内定义或引用本地最小 contract
- 保持现有 consumer 路径不变，避免扩大调用面
- 清空 `utilityBoundaryImportAllowlist`

## Files to modify
- .harness/reuse-decisions/utility-contract-boundary-normalization.md
- docs/plan/impl-plan/2026-05-29-utility-contract-boundary-normalization-plan.md
- docs/reviews/frontend/2026-05-29-utility-contract-boundary-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/utils/contest.ts
- code/frontend/src/utils/platformContestAwdChallengeLinks.ts
- code/frontend/src/utils/skillProfile.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## After implementation
- `utils/contest.ts`、`utils/platformContestAwdChallengeLinks.ts`、`utils/skillProfile.ts` 不再直接依赖 `@/api/contracts`。
- `utilityBoundaryImportAllowlist` 清空。
- 本轮不引入兼容 bridge，不改 consumer import 路径。
