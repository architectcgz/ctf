# Reuse Decision

## Change type
frontend architecture guardrail enhancement / growth guard + AWD owner boundary

## Existing code searched
- scripts/check-architecture.sh
- scripts/check-workflow-complete.sh
- scripts/check-consistency.sh
- docs/architecture/README.md
- code/frontend/src/__tests__/architectureBoundaries.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdContestStateFlags.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts

## Similar implementations found
- 全局前端架构护栏已通过 `architectureBoundaries.test.ts` / `routeViewArchitectureBoundary.test.ts` / `ModalTemplates.test.ts` / `check-theme-tail.mjs` 接入 `scripts/check-architecture.sh --full`。
- AWD 线已经有 feature 级 raw-source 拆分守卫，例如 `awdOperationsPanelTabsExtraction.test.ts` 与 `usePlatformContestAwdBoundary.test.ts`。
- 当前缺口不是“没有边界测试”，而是“没有热点文件增长守卫”和“AWD owner split 没有纳入全局 architecture gate”。

## Decision
refactor_existing

## Reason
最近两轮 `runtime policy owner` 和 `readiness override workflow owner` 收口已经完成，但当前全局架构门禁仍只覆盖：

- 分层导入边界
- route view 约束
- overlay 结构
- theme token 尾巴检查

它不能防止：

- 热点文件在后续 PR 中逐步长大
- `contest-awd-admin` 的 owner split 回退到旧的“大 shell + 混写 workflow”

本轮最小正确改动是：

- 新增前端热点文件 growth guard，使用显式 baseline + max growth / max lines 约束
- 新增 AWD owner boundary tests，守住 `useAwdContestStateFlags` / `useAwdReadinessDecision` / `useAwdRoundOperations` / `usePlatformContestAwd` 的 owner 分工
- 把这两类守卫接入 `scripts/check-architecture.sh --full`
- 同步 `scripts/check-workflow-complete.sh` 与 `docs/architecture/README.md`

本轮不做：

- 全仓库所有 feature 的统一 growth baseline
- 重新引入 allowlist 模式
- 把所有 extraction tests 都并入 architecture gate；先只接最容易回退、最近连续重构的 AWD owner 线

## Files to modify
- .harness/reuse-decisions/frontend-architecture-growth-guards.md
- docs/plan/impl-plan/2026-05-30-frontend-architecture-growth-guards-plan.md
- docs/reviews/frontend/2026-05-30-frontend-architecture-growth-guards-review.md
- docs/architecture/README.md
- scripts/check-architecture.sh
- scripts/check-workflow-complete.sh
- code/frontend/package.json
- code/frontend/scripts/check-frontend-growth-guard.mjs
- code/frontend/scripts/frontend-growth-baseline.json
- code/frontend/src/features/contest-awd-admin/model/useAwdOwnerBoundaries.test.ts

## After implementation
- 热点 AWD 文件会有明确的 `max_lines` 和 `max_growth` 守卫
- AWD owner split 会从局部 feature test 进一步进入全局 architecture gate
- 后续再碰这些热点文件时，不会只靠 code review 人脑判断“是不是又长回去了”
