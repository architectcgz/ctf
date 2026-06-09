# Frontend Architecture Boundary Fix Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-frontend-architecture-boundary-fix`
- Branch: `task/2026-06-09-frontend-architecture-boundary-fix`
- Task Slug: `2026-06-09-frontend-architecture-boundary-fix`
- Plan: `docs/plan/impl-plan/2026-06-09-frontend-architecture-boundary-fix-implementation-plan.md`
- Diff source: current uncommitted diff in the worktree
- Files reviewed:
  - `code/frontend/src/pages/challenges/ChallengeListRoutePage.vue`
  - `code/frontend/src/features/challenge-list/ui/ChallengeListPage.vue`
  - `code/frontend/src/features/challenge-list/index.ts`
  - `code/frontend/src/features/challenge-list/ui/index.ts`
  - `code/frontend/src/pages/challenges/__tests__/ChallengeList.test.ts`
  - `code/frontend/src/pages/platform/AuditLogRoutePage.vue`
  - `code/frontend/src/pages/platform/ImageManageRoutePage.vue`
  - `code/frontend/src/pages/platform/challenges/ChallengeTopologyStudioRoutePage.vue`
  - `code/frontend/src/pages/platform/contests/ContestAwdConfigRoutePage.vue`
  - `code/frontend/src/pages/platform/contests/ContestProjectorRoutePage.vue`
  - `code/frontend/src/features/platform/{audit-log,image-management,contest-projector,contest-awd-config,challenge-topology-studio}/**`
  - `code/frontend/scripts/frontend-architecture-policy.json`
  - `code/frontend/scripts/frontend-growth-baseline.json`
  - `code/frontend/scripts/vue-deep-allowlist.json`
  - `code/frontend/src/__tests__/asyncChunkBoundaries.test.ts`
  - `code/frontend/src/__tests__/duplicateActionGuardAudit.test.ts`
  - `docs/architecture/features/AWD检查器试跑设计.md`
  - `docs/architecture/features/AWD检查器结构化编辑器设计.md`
  - `docs/architecture/features/题包拓扑同步与导出架构.md`

## Classification Check

- 结论：同意 implementation plan 的 `非琐碎任务` 分类。
- 原因：这次同时调整 route/page owner、feature 命名空间、架构 guardrail 和 active 架构文档，属于明确的 frontend architecture boundary 修复，不是局部可逆小改。

## Gate Verdict

- `pass`
- 说明：上一轮 gate review 里的 active docs stale path blocker 已修复。当前代码 owner、platform feature 路径、active guardrail 和 touched active 架构文档已经一致，可以通过 `code-workflow` completion gate。

## Re-review Update

- 复核范围：仅针对上一轮 blocker 指向的 active 架构文档 stale path 修复。
- 复核结果：
  - `docs/architecture/features/AWD检查器试跑设计.md` 已改到当前真实 owner/path，旧 `components/platform/contest`、`ContestAwdConfig.vue`、`views/platform/ContestAwdConfig.vue` 路径已退出 active facts。
  - `docs/architecture/features/题包拓扑同步与导出架构.md` 已把 topology route entry 更新为 `code/frontend/src/pages/platform/challenges/ChallengeTopologyStudioRoutePage.vue`。
  - `rg "views/platform|components/platform/contest|ContestAwdConfig\\.vue|ChallengeTopologyStudio\\.vue|features/(contest-awd-config|challenge-topology-studio)" ...` 在这三份 active feature docs 上无匹配。

## Findings

### Blocking

- 无。上一轮 blocker 已关闭。

## Material Findings

- 无。

## Non-blocking Suggestions

- 这次 `python3 scripts/check-docs-consistency.py` 能通过，但没有捕获 active 架构文档中“路径存在性 / current fact stale path”问题。后续可以考虑给 `docs/architecture/features/*.md` 的代码路径增加存在性校验，减少这类文档漂移再次漏过。

## Missing Validation

- 无额外代码侧验证缺口。已有 `git diff --check`、前端 test guard、frontend architecture check、focused Vitest、`npm run typecheck` 和 `completion-full` 证据，对本轮代码形态已足够。
- reviewer 另行复用并补跑的最小检查：
  - `timeout 120s git diff --check`：通过
  - `timeout 180s python3 scripts/check-docs-consistency.py`：通过
  - `rg "views/platform|components/platform/contest|ContestAwdConfig\\.vue|ChallengeTopologyStudio\\.vue|features/(contest-awd-config|challenge-topology-studio)" docs/architecture/features/AWD检查器试跑设计.md docs/architecture/features/AWD检查器结构化编辑器设计.md docs/architecture/features/题包拓扑同步与导出架构.md -n`：无匹配

## Open Questions Or Assumptions

- 假设 `docs/architecture/features/*.md` 仍被当作 `Current/final` 事实源使用，而不是历史追溯材料。这个假设与仓库 `AGENTS.md`、`docs/文档规范.md`、以及本次 reviewer handoff 中列出的 `Architecture Inputs` 一致。

## Senior Implementation Assessment

- 代码实现本身是当前需求下更低风险的做法：`ChallengeListRoutePage.vue` 已收敛成真正的 thin route entry，且 `ChallengeListPage.vue` 与旧 route page 内容相比只改了 feature 内相对 import，DOM/class/style 语义没有漂移。
- 平台专属能力改为物理移动到 `features/platform/*`，而不是靠旧路径 wrapper alias 续命，这也符合仓库对 single owner 的要求；源码与 active guardrail 中未再发现旧顶层 `features/{audit-log,image-management,contest-projector,contest-awd-config,challenge-topology-studio}` 的活动消费面。
- active feature docs 现在也已收口到当前事实，因此本轮 touched surface 的代码与文档 owner 一致。

## Required Re-validation

- 本轮复核所需重验已完成，无额外必需重跑项。

## Residual Risk

- archive/history 文档中仍有大量旧路径引用，但本轮 non-goal 已明确不做历史清理；我没有把 `docs/plan/archive/`、`docs/reviews/frontend/archive/`、旧 backlog 文档中的历史路径当作 blocker。
- 当前 active/touched 文档 surface 已无 stale path blocker。

## Touched Known-Debt Status

- 本轮触达的已知 debt 是前端 route/feature owner 漂移。
- 代码侧在 touched surface 内基本已经收口：challenge list route 壳已薄化，platform-only feature 也已从顶层 `features/*` 迁到 `features/platform/*`，没有继续留下 wrapper alias 双轨。
- 文档侧在 touched surface 内也已同步收口，因此这条 debt 在本轮 touched surface 上可以视为关闭。
