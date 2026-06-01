# Contest Announcements Panel Feature UI Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-contest-announcements-panel-feature-ui-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-announcements-panel-feature-ui-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-contest-announcements-panel-feature-ui-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-contest-announcements-panel-feature-ui-normalization-review.md`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/features/platform-contests/ui/index.ts`
    - `code/frontend/src/features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/ContestAnnouncementsWorkspacePanel.vue`
    - `code/frontend/src/views/platform/ContestAnnouncements.vue`
    - `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 `platform-contests` 单一 feature UI owner 的继续收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- `ContestAnnouncementsTopbarPanel.vue` 与 `ContestAnnouncementsWorkspacePanel.vue` 已迁入 `features/platform-contests/ui`，不再继续滞留在旧 `components/platform/contest/*` 目录。
- `ContestAnnouncements.vue` 已改为通过 `@/features/platform-contests` public API 组合 panel，route shell 继续只保留 loading / error / panel 装配与 page owner 调用。
- 本次没有顺手调整 `useContestAnnouncementsPage()` 或 `features/contest-announcements/*` 的 workflow owner，因此 UI owner 收口和公告行为 owner 保持分刀，边界清楚。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这次只处理 contest announcements 的 page-sized panel owner，不继续拆内部表单与历史列表 primitive；如果后续继续下钻，应再单独评估这些子块是否值得抽成更细粒度 UI。

## Touched known-debt status

- `platform-contests` 线上与公告路由直接绑定的 page-sized UI 已在 touched surface 内继续从旧 `components/platform/contest/*` 收口到 feature owner，当前这条线上的遗留重点开始转向其它 announcements / operations hub 的大颗粒 panel。
