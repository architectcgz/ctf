<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 后端模块设计文档拆分 Implementation Plan

**Goal:** 将后端模块级设计拆成可按模块读取的当前事实源，并把每个模块的 API、Application / Service 与数据设计写在同一篇文档中。

**Architecture:** 文档事实源落在 `docs/architecture/backend/modules/`，按 `code/backend/internal/module/*` 的模块 owner 组织。此次只更新文档和入口路由说明，不改变后端代码、API 行为、数据库 schema 或运行时配置。

**Tech Stack:** Markdown、Go source inspection、`rg`、`python3 scripts/check-docs-consistency.py`、`bash scripts/check-workflow-governance.sh`

---

## Task Metadata

- Task Slug: `2026-06-24-backend-module-docs`
- Started At: `2026-06-24T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `task/2026-06-24-backend-module-docs`
- Plan Type: `slice`

## Plan Status

- Status: `implemented`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 建立后端模块级设计文档入口，并让每个模块文档覆盖定位、API、Application / Service、数据设计、边界、依赖和 Guardrail。
- Non-Goals: 不修改后端实现、不新增 API、不改 OpenAPI bundle、不调整数据库 migration、不拆分 Go package。

## Problem Statement

- Current behavior / structure: 后端 docs 以总体架构、数据库、API 和关键流程为主，模块内部 API、service 和数据 owner 分散在多份文档中。
- Target behavior / structure: `docs/architecture/backend/modules/` 成为模块化单体的模块级事实源，按模块聚合 API、service、db 和跨模块依赖。
- Why this task is needed now: 用户明确要求参考 `zhicore-go` 把架构设计和各服务具体设计写开，同时要求同一个 API、DB、service 设计放在同一篇模块文档里。

## Inputs

- Source docs: `AGENTS.md`、`docs/文档规范.md`、`docs/architecture/backend/README.md`
- Related architecture/contracts: `docs/architecture/backend/01-system-architecture.md`、`docs/architecture/backend/02-database-design.md`、`docs/architecture/backend/04-api-design.md`
- Related prior work: 本次源码对照 review 发现的 handler 命名、路由遗漏、竞赛导出 owner 和 `TagHandler` 接线问题。

## Task Classification

- Classification: `非琐碎任务`
- Why: 该任务新增长期架构事实源目录，并修改 `docs/文档规范.md` 与 `AGENTS.md` 路由，影响后续 agent 和开发者读取模块 owner 的默认路径。

## Files

- Create: `docs/architecture/backend/modules/*.md`、`docs/plan/impl-plan/2026-06-24-backend-module-docs-implementation-plan.md`
- Modify: `AGENTS.md`、`docs/architecture/backend/README.md`、`docs/文档规范.md`
- Review: `code/backend/internal/app/router_*.go`、`code/backend/internal/module/*/api/http/*.go`、`code/backend/internal/module/*/{application,entity,infrastructure,ports}/`
- Test: 文档一致性、workflow governance、源码关键字反查。

## 复用与 Owner 决策

- Existing patterns searched: 使用 `rg --files`、`rg -n "func \\(h \\*.*Handler\\)"`、`rg -n "CreateContestExport|TagHandler|instances/:id/proxy"` 对照路由、handler 和实体。
- Reuse / extend / split / create-new decision: 新建 `docs/architecture/backend/modules/` 作为模块级事实源，复用现有 `docs/architecture/backend/README.md` 作为父级入口。
- Owner boundary: 文档按 `code/backend/internal/module/*` owner 归档；跨 owner 的 URL 挂载和数据 owner 必须在相关模块中同时写清。
- Why this is the narrowest safe surface: 只改文档事实源和入口路由说明，不触碰后端代码、契约生成或数据库。

## Intake Analysis Gate

- Relevant superpowers analysis pass: 使用 `harness-router`、`documentation-architecture`、`architect-agent` 和源码对照 review。
- Why this pass fits: 任务属于结构化文档事实源修复，需要先确认项目文档归属和当前 Go 模块边界。
- grill-with-docs findings: 当前文档规范要求架构事实写明 API、数据影响、边界和 Guardrail；模块文档需要避免把单体误写成多进程服务。
- Plan adjustments after challenge: 用户指出同一 API、DB、service 设计应写在同一模块文档中，因此文档结构改为按模块聚合，而不是按 API / DB / service 分拆。

## Execution Slices

### Slice 1: 建立并修正后端模块设计事实源

- Goal: 新增模块文档并补齐 review 中确认的缺失、错误和模糊点。
- Dependencies: 当前 Go 源码路由、handler、entity 和 repository 是事实来源。
- Files:
  - Create: `docs/architecture/backend/modules/*.md`
  - Modify: `AGENTS.md`、`docs/architecture/backend/README.md`、`docs/文档规范.md`
  - Review: `code/backend/internal/app/router_*.go`、`code/backend/internal/module/*/api/http/*.go`
  - Test: `scripts/check-docs-consistency.py`、`scripts/check-workflow-governance.sh`
- 步骤：
  - [x] 步骤 1：新增模块文档入口和各模块文档。
  - [x] 步骤 2：对照源码修正 handler 名、路由遗漏、竞赛导出 owner、角色维护和 runtime allocation 表述。
- Validation: 运行文档一致性检查和 workflow governance，并用 `rg` 反查旧错误关键字无残留。
- Review focus: 每个模块内部是否同时覆盖 API、Application / Service、数据设计和跨模块 owner。
- Done criteria: 所有模块文档可以作为当前事实源读取，且项目文档检查通过。

## Impact And Compatibility

- API / DTO: none，文档只描述当前已存在 API。
- Data / migration: none，文档只描述当前表和 owner。
- State / cache / queue / event: none，文档只描述当前副作用。
- Runtime / config: none。
- Frontend route / state / UX: none。
- Docs / contracts: 新增后端模块设计事实源，并更新父级入口和文档归属规则。

## Plan Review / Architecture Fit

- Target owner boundary: `auth`、`identity`、`challenge`、`container_runtime`、`instance`、`practice`、`contest`、`assessment`、`ops`、`teaching_analysis` 各自拥有对应模块文档。
- Reuse points / landing zones: 父入口仍是 `docs/architecture/backend/README.md`，模块入口是 `docs/architecture/backend/modules/README.md`。
- Known structural debt touched: 只记录 `contest` 和 `practice` 等大 owner 的当前边界，不处理 Go 代码拆分。
- How this plan avoids behavior-only convergence: 文档按 owner、API、service 和数据副作用共同收口，而不是只列路由或只列表。
- Hidden second-redesign risk: 如果后续 OpenAPI 或代码 owner 继续变化，模块文档需要同步维护；本次不新增机械路由全量 diff 检查。
- Decision after review: 当前方案符合项目文档规范和用户要求，可以作为本次文档事实源提交。

## Documentation Owner

- Current fact sources to read: `docs/architecture/backend/modules/`、`docs/architecture/backend/README.md`、`docs/文档规范.md`
- Fact sources to update after implementation: `docs/architecture/backend/modules/` 已更新。
- Plan-only notes that must not become architecture source: 本计划中的执行步骤和验证证据只作为过程记录，不替代模块文档。
- Archive condition: 本次提交完成并被用户接受后，可按项目流程归档计划。

## Validation

- 计划验证范围：文档引用、架构文档基本结构、workflow governance、旧错误关键字反查。
- 命名 / 契约检查范围：handler 名称、路由路径、表名和跨模块 owner。
- 完成判定：检查通过，且旧错误关键字不再出现在模块文档中。

## Validation Plan

- Per-slice commands: `python3 scripts/check-docs-consistency.py`
- Integration commands: `bash scripts/check-workflow-governance.sh`
- Manual checks: 对照 `router_user_practice_routes.go`、`router_user_teacher_routes.go`、`router_admin_contest_core_routes.go`、`router_authoring_challenge_routes.go`、`router_authoring_asset_routes.go`。
- Commands intentionally skipped and why: 后端单元测试未运行，因为本次没有修改 Go 代码、API 实现或数据库迁移。

## Validation Evidence

- Command: `python3 scripts/check-docs-consistency.py`
  - Result: PASS
  - Notes: 文档引用和图表源一致。
- Command: `bash scripts/check-workflow-governance.sh`
  - Result: PASS
  - Notes: 文档入口、workflow guardrail 和项目治理检查通过。
- Command: `rg -n "StopInstance|StopTeacherInstance|CreateAWDDefenseSSHAccess|StartInstance|ReviewManualSubmission|StartAWDServiceInstance|StartContestInstance|端口 allocation 表|migration / 管理端" docs/architecture/backend/modules/*.md`
  - Result: 无输出
  - Notes: review 中发现的旧错误命名和模糊短语已清理。

## Independent Review Handoff

- Review target: 后端模块文档是否完整反映当前源码 owner、API、service 和数据设计。
- Validation evidence summary: 文档一致性与 workflow governance 均已通过，并完成关键字反查。
- Architecture / contract inputs: `docs/architecture/backend/modules/*.md`、`code/backend/internal/app/router_*.go`、`code/backend/internal/module/*/api/http/*.go`
- Known risks / review focus: `challenge` 的 `TagHandler` 当前未公开接线，应保持文档中的限制说明；竞赛导出 URL 挂在 contest 路由但 owner 是 assessment。
- Project-local checks to consider: `python3 scripts/check-docs-consistency.py`、`bash scripts/check-workflow-governance.sh`

## Rollback / Recovery

- Safe revert boundary: 回退本次提交即可移除模块文档入口、计划和相关索引更新。
- Data / config / runtime recovery notes: none。
- Irreversible operations: none。

## Residual Risks

- Risk: 模块文档当前依赖人工源码对照，后续新增路由时仍需要开发者同步维护。
- Why acceptable: 项目已在 `AGENTS.md` 中把模块 owner、用例、依赖和 Guardrail 修改路由到模块文档。
- Follow-up owner, if any: 后续修改对应 Go 模块的开发者负责同步对应模块文档。
