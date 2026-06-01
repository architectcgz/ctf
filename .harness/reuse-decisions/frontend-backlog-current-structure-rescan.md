# Reuse Decision

## Change type
frontend docs cleanup / architecture backlog rescan

## Existing code searched
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `code/frontend/src/AGENTS` constraints via project `AGENTS.md`
- `code/frontend/src/features/**`
- `code/frontend/src/widgets/**`
- `code/frontend/src/entities/**`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`

## Similar implementations found
- 最近几轮迁移已经把 active runtime surface 收口到 `pages / widgets / features / entities / shared`。
- 当前 backlog 前半段仍以 `components/*` 存量清理为主线，这和现在 `src/components` 已不存在的目录事实不一致。

## Decision
refactor_existing

## Reason
这次不是继续做某一刀 feature 迁移，而是先修正 backlog 的事实源，避免后续所有“下一步做什么”都继续被过期目录误导。

最小正确改动是：

- 把 backlog 前半段的“当前行动指南”从 `components/*` 迁移清单改成基于 `features / widgets / entities / shared` 的当前结构盘点
- 用最新体量扫描结果重新列出真实偏大的 feature owner 面
- 对下方保留的历史 `components/*` 迁移日志补一句总说明，明确那是迁移轨迹，不代表当前目录仍存在这些路径

本轮不做：

- 不回写所有历史 dated progress 文案
- 不新增新的前端架构规则
- 不顺带推进某个 feature 的实现类拆分

## Files to modify
- `.harness/reuse-decisions/frontend-backlog-current-structure-rescan.md`
- `docs/plan/impl-plan/2026-06-01-frontend-backlog-current-structure-rescan-plan.md`
- `docs/reviews/frontend/2026-06-01-frontend-backlog-current-structure-rescan-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- backlog 的“当前建议”会重新对齐到现有目录事实
- 后续前端迁移将优先围绕真实偏大的 feature owner 面推进，而不是继续盯已删除的 `src/components`
- 历史 `components/*` 记录会被保留为迁移轨迹，但不再伪装成当前目录状态
