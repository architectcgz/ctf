> 状态：Current
> 事实源：`ChallengeTopologyStudioPage.vue` 当前 owner、`feature-owned UI` 规则、拓扑编辑 route page 与 feature page model 边界
> 替代：无

# Challenge Topology Studio Feature UI Migration Implementation Plan

## 目标

- 把 `ChallengeTopologyStudioPage.vue` 从 `components/platform/topology/` 迁到 `features/challenge-topology-studio/ui/`。
- 让 `views/platform/ChallengeTopologyStudio.vue` 直接通过 `features/challenge-topology-studio` public API 组合 page-sized UI，同时继续通过 `features/platform-challenges` 持有 route owner。
- 收掉 `ChallengeTopologyStudioPage.vue` 对应的 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist` 例外。

## 非目标

- 本轮不改 `useChallengeTopologyStudioPage()` 的加载、保存、导出、删除、模板管理或画布交互 owner。
- 本轮不改 `useChallengeTopologyStudioRoutePage()` 的 router owner 和返回路径策略。
- 本轮不拆 `TopologyChallengeWorkbench.vue`、`TopologyTemplateWorkbench.vue` 或其它 topology 子分区。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/views/platform/ChallengeTopologyStudio.vue`
- `code/frontend/src/features/challenge-topology-studio/index.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`
- `code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ChallengeTopologyStudioPage.vue` 当前约 `453` 行，只服务 `ChallengeTopologyStudio` route，并且直接依赖 `features/challenge-topology-studio` 的 page model，已经是典型的单一 feature page shell。
- 这页本身不直接持有 `vue-router`，route owner 已经由 `useChallengeTopologyStudioRoutePage()` 留在 `features/platform-challenges`，所以本轮不需要做 owner 回收，只需要做目录归位。
- 拓扑页已经在前几轮把大部分工作区细分到独立组件，当前最佳切片不是再拆模板，而是清掉 `components/*Page.vue` 这层遗留中转。

## 设计边界

### route view 继续负责

- 组合 `useChallengeTopologyStudioRoutePage()` 与 `ChallengeTopologyStudioPage`
- 不直接持有 topology page model、API 调用或画布交互 owner

### `features/platform-challenges/model` 继续负责

- route params 解析、返回详情页导航等 route owner

### `features/challenge-topology-studio/model` 继续负责

- 拓扑加载、保存、导出、删除、模板管理、选择状态和画布交互 owner

### `features/challenge-topology-studio/ui` 本轮负责

- 拓扑编辑 page-sized UI shell
- 消费上层派生状态与事件 handler
- 不直接持有 `vue-router`

### `components/platform/topology/*` 继续保留

- 稳定的 topology header、workbench、context rail、template hero 等子分区
- 不再承担拓扑编辑整页 shell owner

## 任务切片

### Slice 1：迁移 topology page shell

- 目标：
  - 新增 `features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.vue`
  - `features/challenge-topology-studio/index.ts` 暴露 page shell
  - `ChallengeTopologyStudio.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/challenge-topology-studio/index.ts`
  - `code/frontend/src/features/challenge-topology-studio/ui/*`
  - `code/frontend/src/views/platform/ChallengeTopologyStudio.vue`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- Review focus：
  - page shell 是否已经不再滞留在 `components/`
  - route view 是否继续保持薄壳

### Slice 2：清理 guardrail 与 backlog

- 目标：
  - 清理 topology page 对应 allowlist 例外
  - 更新 raw-source 测试路径和 backlog 进展
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-challenge-topology-studio-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - raw-source 测试是否已切到新 owner

## 结构收口检查

- `ChallengeTopologyStudioPage.vue` 不再作为 `components/*Page.vue` 遗留页存在。
- `ChallengeTopologyStudio.vue` 只组合 route model 与 feature public API。
- 拓扑 page shell 不直接持有 `vue-router`。
- touched surface 上至少移除一条 component->feature allowlist 和一条 legacy component page allowlist。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-topology-studio-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-challenge-topology-studio-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-challenge-topology-studio-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/challenge-topology-studio code/frontend/src/views/platform/ChallengeTopologyStudio.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/challenge-topology-studio/ui` 是否成为 topology page shell 的唯一 owner。
- `features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts` 是否仍保留 route owner，没有被 page shell 重新吸回。
- 测试与 allowlist 是否同步反映新边界，而不是继续绑定旧路径。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `ChallengeTopologyStudioPage.vue` 移回 `components/platform/topology/` 并恢复 route view import。
- 本轮不涉及 route owner、API 或持久化变更，回退主要是 page shell 目录回退。

## 残余风险

- topology 子分区仍然留在 `components/platform/topology/`，本轮只做 page shell 归位，不处理更深层组件层级。
- `UserGovernancePage.vue` 等同类 `feature-owned UI` 候选仍在 backlog 中，本轮不一并迁移。
