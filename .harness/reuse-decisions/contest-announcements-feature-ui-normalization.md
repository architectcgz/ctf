# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/contests/ContestAnnouncementRealtimeBridge.vue
- code/frontend/src/components/platform/contest/ContestAnnouncementManageDrawer.vue
- code/frontend/src/features/contest-announcements/index.ts
- code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts
- code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.ts
- code/frontend/src/views/contests/ContestDetail.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `platform-awd-challenges` 和 `platform-challenge-detail` 前两轮已经把只服务单一 feature 的 dialog / panel / workspace UI 收口到 `features/*/ui`，同时清掉对应 allowlist。
- `contest-announcements` 当前已经拥有稳定的 `model/` owner，但仍缺少 `ui/` 落点，导致 realtime bridge 与管理 drawer 继续滞留在业务组件目录。
- `ContestDetail.vue` 与 `ContestManage.vue` 都只是 route / page 组合层，不需要自己持有公告实时订阅或管理抽屉的实现细节。

## Decision
refactor_existing

## Reason
`ContestAnnouncementRealtimeBridge.vue` 和 `ContestAnnouncementManageDrawer.vue` 都只服务 `contest-announcements` 这一条 feature，并直接消费该 feature 的 model contract。继续把它们留在 `components/**` 只会让 `componentFeatureImportAllowlist` 继续冻结历史路径。最小正确改动是补齐 `features/contest-announcements/ui`，把 bridge 和 drawer 迁进去，并让 `ContestDetail.vue`、`ContestManage.vue` 改为通过 feature public API 组合它们。

## Files to modify
- .harness/reuse-decisions/contest-announcements-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-27-contest-announcements-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-27-contest-announcements-feature-ui-normalization-review.md
- code/frontend/src/features/contest-announcements/index.ts
- code/frontend/src/features/contest-announcements/ui/index.ts
- code/frontend/src/features/contest-announcements/ui/ContestAnnouncementRealtimeBridge.vue
- code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue
- code/frontend/src/views/contests/ContestDetail.vue
- code/frontend/src/views/platform/ContestManage.vue
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `contest-announcements` 会同时持有 realtime bridge 与管理 drawer 这组单一 feature UI。
- `ContestDetail.vue` 和 `ContestManage.vue` 继续只做 route / page 组合，不再直连旧组件路径。
- `componentFeatureImportAllowlist` 中 `contest-announcements` 的 2 条历史例外可以移除。
