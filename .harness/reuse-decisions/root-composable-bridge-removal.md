# Reuse Decision

## Change type
composable / test / docs

## Existing code searched
- `code/frontend/src/composables/use*.ts`
- `code/frontend/src/features/**/index.ts`
- `code/frontend/src/features/**/model/index.ts`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`

## Similar implementations found
- `code/frontend/src/composables/useChallengeDetailPresentation.ts`
- `code/frontend/src/composables/usePlatformAwdChallenges.ts`
- `code/frontend/src/composables/useTeacherAwdReviewIndex.ts`
- `code/frontend/src/composables/usePlatformOverviewWorkspace.ts`
- `code/frontend/src/features/challenge-detail/index.ts`
- `code/frontend/src/features/platform-awd-challenges/index.ts`
- `code/frontend/src/features/teacher-awd-review/index.ts`
- `code/frontend/src/features/platform-overview/index.ts`
- `code/frontend/src/composables/useNotificationDropdown.ts`
- `code/frontend/src/features/notifications/index.ts`
- `code/frontend/src/composables/useChallengeManagePresentation.ts`
- `code/frontend/src/features/platform-challenges/index.ts`

## Decision
refactor_existing

## Reason
这次不新增任何新的 feature owner，而是继续收口已经完成迁移的历史桥接壳。当前 `featureBoundaries.test.ts` 维护了一整份 `@/composables/use*.ts -> @/features/*` 的迁移名单，但这些根层 composable 现在已经不再被 runtime 或测试引用，仓库搜索只剩这一个边界测试在引用它们。继续保留这些单行 re-export 壳和名单，只会让根层 `composables` 与 `features/*` 双入口长期共存，把 P1-1 里的 allowlist 架构债留在磁盘结构上。`useChallengeManagePresentation.ts` 现在已经是最后一个仍位于根层 `composables` 的 feature re-export 壳，真实 owner 一直在 `features/platform-challenges`。更低风险的做法是直接删掉它，让根层 composable 目录回到只承载真正通用逻辑的状态，并保持 `ChallengeManage` 等调用方继续通过 feature public API 读取真正 owner。

## Files to modify
- `.harness/reuse-decisions/root-composable-bridge-removal.md`
- `docs/plan/impl-plan/2026-05-24-root-composable-bridge-removal-implementation-plan.md`
- `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- `code/frontend/src/composables/use*.ts` 中仅剩历史桥接职责且无 runtime 引用的文件
- `code/frontend/src/composables/useNotificationDropdown.ts`
- `code/frontend/src/composables/useChallengeManagePresentation.ts`

## After implementation
- 后续如果再出现 `src/composables` 与 `features/*` 双入口，默认先判断是否已经只剩历史桥接壳；若是，直接按 task-scoped reuse decision + 删除桥接 + 更新边界测试的模式收口，不再长期保留“迁移中”名单。
