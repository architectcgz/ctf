# ContestAwdConfig Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/platform/ContestAwdConfig.vue`
    - `code/frontend/src/components/platform/contest/ContestAwdConfigWorkspaceShell.vue`
    - `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `.harness/reuse-decisions/contest-awd-config-workspace-shell-owner-convergence.md`
    - `docs/plan/impl-plan/2026-05-26-contest-awd-config-workspace-shell-owner-convergence-implementation-plan.md`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现继续把 `ContestAwdConfig.vue` 保持为 route view owner：路由参数、服务切换、页面数据加载、checker 预览、配置保存与草稿 state 仍由 `useContestAwdConfigPage()` 负责。
- 新增的 `ContestAwdConfigWorkspaceShell.vue` 只承接稳定的工作台模板、局部样式、展示分区和事件转发，没有重新吸入 API 调用、路由同步或第二份保存/预览 owner。
- `ContestAwdConfig.test.ts` 已改成按父页 + shell 组合源码检查，并额外锁定 `useContestAwdConfigPage.ts` 继续通过 `useAwdCheckerPreviewFlow`、`useAwdCheckerSaveFlow`、`useContestAwdConfigDataLoader` 持有主流程，能防止后续继续抽壳时把请求与路由 owner 回塞到 route view 或壳组件。
- `architectureAllowlist.ts` 已移除 `ContestAwdConfig.vue`，当前 oversized route view allowlist 已清空；这轮 touched surface 的已知结构债已完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- `ContestAwdConfigWorkspaceShell.vue` 仍然是一个偏大的展示壳组件，但 route view owner 已经收口；如果后续继续在配置画布追加展示区或局部表单，优先沿目录区 / 编辑区 / 调试区继续拆成更细的展示分区，而不是把业务 owner 再抬回父页。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `ContestAwdConfig.vue` oversized route view owner。
- 该债务在 touched surface 上已完成收口：父页已脱离 oversized allowlist，本体降到 `94` 行，当前没有把原本的路由、服务切换、预览、保存和草稿 owner 再次混回壳组件。
