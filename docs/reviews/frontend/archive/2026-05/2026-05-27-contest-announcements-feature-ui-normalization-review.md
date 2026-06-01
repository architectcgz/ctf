# Contest Announcements Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-contest-announcements-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-announcements-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-27-contest-announcements-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-27-contest-announcements-feature-ui-normalization-review.md`
    - `code/frontend/src/features/contest-announcements/index.ts`
    - `code/frontend/src/features/contest-announcements/ui/index.ts`
    - `code/frontend/src/features/contest-announcements/ui/ContestAnnouncementRealtimeBridge.vue`
    - `code/frontend/src/features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue`
    - `code/frontend/src/views/contests/ContestDetail.vue`
    - `code/frontend/src/views/platform/ContestManage.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的 feature-owned UI 收口，只迁公告 feature 的 bridge / drawer，不改 model owner。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 已修正：`features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue` 在迁入 feature UI 后一度直接 `import { useRouter } from 'vue-router'`，这会把路由跳转重新下沉到纯 UI owner，并触发 `feature router access should stay in reviewed route-aware composables` 架构护栏。
- 收口方式：抽屉组件改为只发出 `openFullPage(contest)` 事件；`useContestManagePage()` 新增 `openContestAnnouncementsPage(contest)`，由 route-aware composable 执行 `router.push({ name: 'ContestAnnouncements', params: { id: contest.id } })`；`ContestManage.vue` 只做事件透传。

## Senior implementation assessment

- 本轮收口仍然符合“单一 feature 的 UI 壳迁回 `features/*/ui`，route owner 留在 feature model / route-aware composable”的既定方向。
- `ContestAnnouncementRealtimeBridge` 与 `ContestAnnouncementManageDrawer` 现在都经由 `features/contest-announcements` public API 暴露，旧 `components/**` 路径已退场。
- `ContestManage.test.ts` 已补一条 owner 断言，确保“进入完整管理页”的导航不会再次滑回 drawer。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只收口公告 feature 的 bridge / drawer，不会顺手迁移更大块的 contest announcement workspace surface。

## Touched known-debt status

- `contest-announcements` 这组“单一 feature 仍滞留在 `components/**` 的 UI 壳”已在 touched surface 内完成收口。
- 下一批更适合继续看 `contest-workbench` 这类仍保留多块单一 feature UI 的存量。
