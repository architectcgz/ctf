# Reuse Decision

## Change type
frontend test architecture refactor / views residue cleanup

## Existing code searched
- code/frontend/src/views/**/__tests__/*.test.ts
- code/frontend/src/__tests__/architectureBoundaries.test.ts
- code/frontend/src/__tests__/frontendArchitecturePolicy.ts
- code/frontend/scripts/frontend-architecture-policy.json
- scripts/check-frontend-architecture.sh
- code/frontend/src/pages/**
- code/frontend/src/features/**
- code/frontend/src/widgets/**

## Similar implementations found
- 运行时 route page 已统一落到 `src/pages/**`，`views/**/*.vue` 已基本退场。
- feature / widget 的结构测试已经落在 owner 邻近目录，例如 `src/features/**/model/*.test.ts`、`src/widgets/**/**/*.test.ts`。
- 全局前端 guardrail 已统一由 `frontend-architecture-policy.json` 驱动。

## Decision
refactor_existing

## Reason
这轮不是新增页面能力，而是把 `src/views` 剩余的历史测试容器彻底迁离：

- 架构守卫移到更中性的 `src/__tests__`
- 页面行为测试迁回 `pages/**`、`features/**`、`widgets/**` 邻近
- 把只记录某次迁移阶段的 `Phase / Extraction / SurfaceAlignment` 测试压缩成更稳定的策略测试

不做：

- 不重写历史 review / plan 文档里当时的旧测试路径
- 不为了减少文件数而把本来应该邻近 owner 的行为测试重新集中到单一目录
- 不在这轮新增新的前端分层或 allowlist

## Files to modify
- .harness/reuse-decisions/frontend-views-test-migration-phase1.md
- docs/plan/impl-plan/2026-05-30-frontend-views-test-migration-phase1-plan.md
- scripts/check-frontend-architecture.sh
- code/frontend/scripts/frontend-architecture-policy.json
- code/frontend/src/__tests__/frontendArchitecturePolicy.ts
- code/frontend/src/__tests__/architectureBoundaries.test.ts
- code/frontend/src/views/**/__tests__/*.test.ts
- code/frontend/src/pages/**/__tests__/*.test.ts
- code/frontend/src/features/**/__tests__/*.test.ts
- code/frontend/src/widgets/**/__tests__/*.test.ts
- docs/architecture/README.md

## After implementation
- `src/views` 不再作为活动测试 owner；只允许为过渡支撑文件短暂存在，目标是清空
- 架构守卫脚本引用与当前测试实际路径一致
- 页面行为测试按 route / feature / widget owner 就近落位
- 迁移阶段性测试收敛成更少、更稳定的策略守卫
