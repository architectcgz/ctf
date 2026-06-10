<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# 前端架构风险收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent task-by-task execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修正前端架构护栏误报，并用题目详情链路作为首个 slice 收口 route page 厚度和 feature public API 暴露面。

**Architecture:** 保持当前 `pages -> widgets -> features -> entities -> shared` 分层，不做全仓重排。先让架构测试只扫描真实 runtime source，再把 `ChallengeDetailRoutePage.vue` 收成薄 route entry；题目详情绑定层落在 `widgets/challenge-detail-workspace/ChallengeDetailPage.vue`，保持依赖方向为 `widgets -> features`，并把 `features/challenge-detail/index.ts` 从宽 barrel 改为显式 public API。

**Tech Stack:** Vue 3、TypeScript、Vite、Vitest、Vue Test Utils。

---

## Task Metadata

- Task Slug: `2026-06-09-frontend-architecture-risk-fixes`
- Started At: `2026-06-09T08:42:26Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-frontend-architecture-risk-fixes`
- Branch: `task/2026-06-09-frontend-architecture-risk-fixes`

## Objective And Non-Goals

- Objective:
  - 修正 `featureBoundaries` 对 `__tests__/*.test-harness.ts` 的误报。
  - 为题目详情新增 widget-owned page surface，让 route page 只做运行时入口，并保持 `widgets -> features` 的依赖方向。
  - 收窄 `features/challenge-detail` public API，避免 `export * from './model'` / `export * from './ui'` 全量暴露内部实现。
- Non-Goals:
  - 不一次性重构所有 feature `index.ts`。
  - 不拆分所有 500+ 行组件；本次只处理已触达的题目详情 route/page 组合。
  - 不改变题目详情用户可见行为、路由、API 契约或 UI 文案。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/architecture/frontend/01-architecture-overview.md`
  - `docs/architecture/frontend/03-state-management.md`
  - `docs/architecture/frontend/06-components.md`
  - `docs/architecture/frontend/07-pages-dataflow.md`
- Related architecture/contracts:
  - `code/frontend/scripts/frontend-architecture-policy.json`
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
- Related prior work:
  - Existing `SecuritySettingsRoutePage.vue -> useSecuritySettingsPage -> SecuritySettingsWorkspaceShell.vue` pattern in `07-pages-dataflow.md`.

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 触达架构测试、route page 边界和 feature public API。
  - 需要保持现有架构文档和 guardrail 的语义一致。

## Files

- Create:
  - `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailPage.vue`
- Modify:
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
  - `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
  - `code/frontend/src/features/challenge-detail/index.ts`
  - `code/frontend/src/widgets/challenge-detail-workspace/index.ts`
  - `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test-harness.ts`
  - `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
- Review:
  - `code/frontend/src/features/challenge-detail/model/index.ts`
  - `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test-harness.ts`
  - `code/frontend/src/pages/challenges/__tests__/*.test.ts`
- Test:
  - `code/frontend/src/features/__tests__/featureBoundaries.test.ts`
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
  - `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
  - Relevant challenge detail tests under `code/frontend/src/pages/challenges/__tests__/`

## 复用与 Owner 决策

- Existing patterns searched:
  - route page thin shell pattern in `PlatformOverviewRoutePage.vue` and `TeacherDashboardRoutePage.vue`
  - page dataflow owner rules in `docs/architecture/frontend/07-pages-dataflow.md`
  - `?raw` source guard usage in `pages/challenges/__tests__/ChallengeDetail.test-harness.ts`
- Reuse / extend / split / create-new decision:
  - Reuse `useChallengeDetailPage()` as the state/workflow owner.
  - Create one widget-owned `ChallengeDetailPage.vue` to hold the existing `ChallengeDetailWorkspace` binding.
  - Do not create a generic route-page wrapper abstraction.
- Owner boundary:
  - `ChallengeDetailRoutePage.vue`: route runtime entry only.
  - `widgets/challenge-detail-workspace/ChallengeDetailPage.vue`: page surface binding between feature model and widget workspace.
  - `useChallengeDetailPage.ts`: route-aware page workflow, async loading, interactions, tab/query state.
  - `ChallengeDetailWorkspace.vue`: display composition only.
- Why this is the narrowest safe surface:
  - The user-visible workflow stays unchanged.
  - The route page stops being a large prop-forwarding glue layer.
  - Public API tightening is limited to the touched `challenge-detail` slice.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - The task starts from architecture risks and needs boundary decisions before implementation.
- grill-with-docs findings:
  - No domain glossary or user-visible behavior ambiguity blocks the work.
  - Existing frontend architecture docs already define route page, feature model, feature UI and widget owners.
- Plan adjustments after challenge:
  - Limit public API tightening to `features/challenge-detail` instead of attempting a full-repo barrel migration.
  - Keep the new binding component in `widgets` rather than `features`, because existing architecture policy forbids `features` importing `widgets`.

## Validation

- Commands:
  - `cd code/frontend && pnpm exec vitest run src/features/__tests__/featureBoundaries.test.ts`
  - `cd code/frontend && pnpm exec vitest run src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts src/features/__tests__/featureBoundaries.test.ts`
  - `cd code/frontend && pnpm exec vitest run src/pages/challenges/__tests__`
  - `cd code/frontend && pnpm run typecheck`
- Manual checks:
  - Inspect `ChallengeDetailRoutePage.vue` remains a thin route entry.
  - Inspect `features/challenge-detail/index.ts` exposes only intentional public API.
- Review focus:
  - Architecture guard accuracy.
  - Route/page/widget owner drift.
  - Public API accidentally hiding imports used by runtime/tests.

## Execution Tasks

### Task 1: Fix feature boundary guard runtime filtering

- [x] Update `featureBoundaries.test.ts` so runtime scanning excludes `__tests__`, `.test.ts`, `.spec.ts`, and `.test-harness.ts`.
- [x] Run `cd code/frontend && pnpm exec vitest run src/features/__tests__/featureBoundaries.test.ts`.
- [x] Mark this task complete only if the prior `?raw` harness imports no longer fail the runtime deep-import guard.

### Task 2: Move challenge detail route binding into widget page surface

- [x] Create `widgets/challenge-detail-workspace/ChallengeDetailPage.vue` using the existing binding logic from `pages/challenges/ChallengeDetailRoutePage.vue`.
- [x] Replace `ChallengeDetailRoutePage.vue` with a thin render of `ChallengeDetailPage`.
- [x] Export `ChallengeDetailPage` from `widgets/challenge-detail-workspace/index.ts`.
- [x] Run focused challenge detail page tests under `src/pages/challenges/__tests__`.

### Task 3: Tighten challenge-detail public API

- [x] Replace `features/challenge-detail/index.ts` wide barrel exports with explicit exports needed by route page, widgets, tests, and callers.
- [x] Run `cd code/frontend && rg "from ['\\\"]@/features/challenge-detail/(model|ui)" src -g '!**/*.test.ts' -g '!**/__tests__/**'` and verify runtime callers still use the public feature entry where appropriate.
- [x] Run the architecture guard command from Validation.

### Task 4: Final verification and review

- [x] Run `cd code/frontend && pnpm run typecheck`.
- [x] Run `bash scripts/check-frontend-test-guard.sh --files code/frontend/src/features/__tests__/featureBoundaries.test.ts code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test-harness.ts code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`.
- [x] Perform a same-context review pass focused on owner drift, public API misses, and test brittleness.

Note: independent reviewer subagent was not spawned because the current multi-agent tool policy only permits spawning when the user explicitly requests delegation. The same-context review does not satisfy the independent review gate.
