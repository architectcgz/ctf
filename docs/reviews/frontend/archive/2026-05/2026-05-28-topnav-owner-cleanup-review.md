# TopNav Owner Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-topnav-owner-cleanup-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/topnav-owner-cleanup.md`
    - `docs/plan/impl-plan/2026-05-28-topnav-owner-cleanup-plan.md`
    - `docs/reviews/frontend/2026-05-28-topnav-owner-cleanup-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/components/layout/TopNav.vue`
    - `code/frontend/src/components/layout/topnav/useTopNavViewState.ts`
    - `code/frontend/src/components/layout/topnav/topNavShell.css`
    - `code/frontend/src/components/layout/__tests__/TopNav.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意按 layout 共享基础设施上的非 trivial frontend refactor 处理，目标是继续收口 `TopNav.vue` 的本地 view-state、breadcrumb detail 推导与样式混写，而不是改 route/theme/session bridge contract。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `TopNav.vue` 现在只保留 props、notification trigger slot、logout / theme / breadcrumb / user card 的顶层组合，不再继续混放 mobile 判定、brand picker lifecycle、breadcrumb detail 推导与壳体样式。
- `useTopNavViewState.ts` 已明确承接 `isMobile`、brand picker open/close、outside click / `Escape` cleanup、`currentBrandLabel`、`userDisplayName`、`userInitial`、`roleCaption`、`backofficeBreadcrumb` 和 detail label 推导；这批本地 topnav view-model 不再滞留在父 SFC。
- `topNavShell.css` 已承接 topnav shell、collapsed 宽度补偿、brand picker、user card 和 light/dark token 样式，raw-source 与 shared-theme 护栏也已切到聚合源码视角，不再假设样式必须留在父 SFC。
- `TopNav.vue` 文件体量从约 `680` 行降到 `125` 行；本轮 touched surface 上“route/theme/session bridge + 本地 view-state + breadcrumb detail 推导 + shell CSS”混写债已完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据。当前流程里没有额外派生 reviewer，所以这条缺口需要在交付说明里明确。
- breadcrumb detail label 规则已经迁出父组件，但仍集中在一个 topnav 局部 composable；如果未来继续增长，更适合单独拆出 breadcrumb detail support，而不是再回灌父组件。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 layout P2 对 `TopNav.vue` 的“更深层本地 view-state / shell CSS 清理”已在 touched surface 内完成收口；当前三大 layout 组件的同级别大壳体债已基本出清，残余重点已从大文件混写转向更细粒度 contract 或表现层细节。
