# Instance Panel Contract Boundary Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-instance-panel-contract-boundary-normalization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/instance-panel-contract-boundary-normalization.md`
  - `docs/plan/impl-plan/2026-05-29-instance-panel-contract-boundary-normalization-plan.md`
  - `docs/reviews/frontend/2026-05-29-instance-panel-contract-boundary-normalization-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/components/common/InstancePanel.vue`
  - `code/frontend/src/components/common/instancePanel.types.ts`
  - `code/frontend/src/components/common/__tests__/InstancePanel.test.ts`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- Classification check：同意按 common 组件本地展示 contract owner 收口处理，风险主要在本地类型是否遗漏当前模板依赖字段，以及 allowlist 是否真实清空。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `InstancePanel.vue` 已不再直接依赖 `@/api/contracts`，shared/common 层这条历史 API contract 例外已经回收到本地最小展示类型。
- `instancePanel.types.ts` 只保留该组件真实需要的状态、分享范围和实例卡片字段，没有把完整 API DTO 继续透传到 common 层。
- `InstancePanel.test.ts` 已补上 `@/api/contracts` 的负向源码断言，除了 architecture boundary 测试外，多了一层直接护栏。
- `commonForbiddenImportAllowlist` 已清空；allowlist A 里 common contract 这一组历史例外已经结束。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/common/__tests__/InstancePanel.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/instance-panel-contract-boundary-normalization.md docs/plan/impl-plan/2026-05-29-instance-panel-contract-boundary-normalization-plan.md docs/reviews/frontend/2026-05-29-instance-panel-contract-boundary-normalization-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/components/common/InstancePanel.vue code/frontend/src/components/common/instancePanel.types.ts code/frontend/src/components/common/__tests__/InstancePanel.test.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `InstancePanel` 与 `ChallengeInstanceCard`、`useInstanceListPage.ts` 之间仍有少量 instance 状态标签重复；这已经不属于 allowlist 例外，但后续如果要继续做 instance 展示层收敛，可以单独开题评估 `entities/instance` 或 shared instance presentation owner。
