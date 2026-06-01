# Frontend Backlog Current Structure Rescan 计划

## Objective

- 校正 `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中仍以 `src/components` 为当前事实的过期段落。
- 基于当前 `features / widgets / entities / shared` 结构，重新给出下一步前端迁移优先面。

## Non-goals

- 不重写整份历史迁移日志。
- 不在本轮直接拆任何 feature 大组件。
- 不调整前端架构 policy 或测试规则。

## Source Inputs

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `AGENTS.md`
- `code/frontend/src/features/**`
- `code/frontend/src/widgets/**`
- `code/frontend/src/entities/**`
- 当前非测试 `.vue/.ts` 体量扫描结果

## Plan Review Result

- 需要修的是“现在时的行动指南”，不是整个历史年表。
- 最合适的做法是替换 backlog 前半段当前结构概览，并给后文的历史 `components/*` 记录加显式说明。

## Task Slices

### Slice 1: 校正当前结构概览

- 目标：删除把 `components/*` 当成当前目录事实的清单，改成现有 `shared / entities / widgets / features` 分层概览。
- 风险：如果只删旧段落，不补新结构，backlog 会失去接下来几轮迁移的导航作用。

### Slice 2: 重排真实偏大的 owner 面

- 目标：基于最新扫描结果列出当前最值得继续收口的 feature owner 面，而不是历史组件目录。
- 风险：要区分“体量大但 owner 已正确”和“体量大且仍值得近期拆”的差别。

### Slice 3: 标记历史组件路径为迁移轨迹

- 目标：保留 dated progress 的追溯价值，同时避免读者误以为 `src/components` 仍存在。
- 风险：如果不加说明，后续 agent 仍可能把历史段落误判成当前任务入口。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision frontend-backlog-current-structure-rescan`
- `git diff --check -- .harness/reuse-decisions/frontend-backlog-current-structure-rescan.md docs/plan/impl-plan/2026-06-01-frontend-backlog-current-structure-rescan-plan.md docs/reviews/frontend/2026-06-01-frontend-backlog-current-structure-rescan-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `bash scripts/check-consistency.sh`

## Review Focus

- backlog 前半段是否已经完全摆脱把 `src/components` 视为当前事实的叙述
- 新的优先级排序是否基于当前真实 owner 面，而不是机械沿用旧迁移目录
- 历史 `components/*` 记录是否被明确标成迁移轨迹
