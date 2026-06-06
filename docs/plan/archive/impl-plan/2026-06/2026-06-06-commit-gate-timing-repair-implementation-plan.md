# commit-gate-timing-repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 收口 `ctf` 当前提交门禁与治理审计的职责边界，解决 `check-consistency.sh` 命名/职责漂移、归档后 gate 失活导致无法提交，以及 worktree 下前端 quick check 依赖脆弱的问题。

**Architecture:** 把重型治理审计明确收口为独立 review/doctor 入口，保留 `check-consistency.sh` 作为兼容 wrapper；startup gate 增加“ready to merge”状态，允许 plan 已归档但 task 仍可继续提交；前端 quick check 通过共享依赖发现脚本在 worktree 与主工作区之间复用现有 `node_modules`。

**Tech Stack:** Bash scripts, Python 3 stdlib, Git hooks, local startup gate metadata, npm/vitest frontend tooling

---

## Task Metadata

- Task Slug: `2026-06-06-commit-gate-timing-repair`
- Started At: `2026-06-06T06:49:20Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-gate-timing-repair`
- Branch: `task/2026-06-06-commit-gate-timing-repair`

## Objective And Non-Goals

- Objective:
  - 把当前 `scripts/check-consistency.sh` 的 review/doctor 语义显式化，避免名字继续误导为提交前普通一致性检查。
  - 让 `archive-task-artifacts.sh` 不再在提交前把当前 task gate 直接失活。
  - 让 `scripts/check-frontend-architecture.sh --quick` 在 worktree 下自动复用主工作区前端依赖，而不是因为缺少 `node_modules` 假失败。
  - 同步更新 hook / AGENTS / README / doctor / consistency wiring，保证新分层可发现、可验证。
- Non-Goals:
  - 不重写前后端架构检查本身的业务规则。
  - 不在这次任务里重做全部历史文档中的 `check-consistency.sh` 文案。
  - 不把 review/doctor 审计重新塞回 pre-commit。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `docs/README.md`
  - `README.md`
- Related architecture/contracts:
  - `.githooks/pre-commit`
  - `scripts/check-consistency.sh`
  - `scripts/check-workflow-complete.sh`
  - `scripts/check-frontend-architecture.sh`
  - `scripts/archive-task-artifacts.sh`
  - `harness/checks/check_startup_gate.py`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-05-workflow-governance-refactor-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-06-commit-message-governance-extract-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 workflow gate、归档脚本、frontend quick check、review 审计入口和仓库规则文档。
  - 需要解决实际提交阻塞，并重新划清 pre-commit 与 review/doctor 的 owner 边界。

## Files

- Create:
  - `scripts/check-review-governance.sh`
  - `scripts/ensure-frontend-tooling.sh`
- Modify:
  - `scripts/check-consistency.sh`
  - `scripts/check-workflow-complete.sh`
  - `scripts/check-frontend-architecture.sh`
  - `scripts/archive-task-artifacts.sh`
  - `harness/checks/check_startup_gate.py`
  - `.githooks/README.md`
  - `AGENTS.md`
  - `README.md`
  - `docs/README.md`
  - `docs/contracts/README.md`
  - `docs/architecture/README.md`
  - `docs/文档规范.md`
  - `scripts/install-githooks.sh`
  - `scripts/doctor-local-harness.sh`
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
- Review:
  - `.githooks/pre-commit`
- Test:
  - `bash scripts/check-review-governance.sh`
  - `bash scripts/check-consistency.sh`
  - `bash scripts/check-frontend-architecture.sh --quick`
  - `bash scripts/archive-task-artifacts.sh`

## 复用与 Owner 决策

- Existing patterns searched:
  - 当前 `pre-commit -> check-startup-gate -> check-architecture --quick` 链路
  - 当前 `archive-task-artifacts.sh` 与 `harness/checks/check_startup_gate.py`
  - 当前 `check-workflow-complete.sh` 与 README / AGENTS 对 `check-consistency.sh` 的定义
- Reuse / extend / split / create-new decision:
  - 保留 `check-consistency.sh` 入口名作为兼容别名，但把真实重型审计收口到 `check-review-governance.sh`。
  - 复用现有 startup gate 元数据结构，只增加 `ready_to_merge` 这类中间状态，不重写整个 workflow。
  - 新增统一前端依赖发现脚本，供架构检查与本地 doctor 共同复用。
- Owner boundary:
  - pre-commit owner：当前 staged diff 的轻量、可本地快速满足的门禁。
  - review/doctor owner：仓库导航、OpenAPI bundle、文档引用、guardrail wiring、完整治理审计。
  - startup gate owner：任务是否按 workflow 进入实现，以及在“已归档但待提交/待合并”阶段是否仍允许继续收口。
- Why this is the narrowest safe surface:
  - 不重写整套 workflow，只补齐当前真正出问题的三个边界：命名/职责、归档/提交时序、worktree 依赖发现。

## Validation

- Commands:
  - `bash scripts/check-review-governance.sh`
  - `bash scripts/check-consistency.sh`
  - `bash scripts/check-frontend-architecture.sh --quick`
  - `bash scripts/check-workflow-complete.sh`
- Manual checks:
  - 确认 `.githooks/README.md` 与 `AGENTS.md` 不再把 `check-consistency.sh` 描述成提交前主门禁。
  - 确认 task plan 归档后，当前 worktree 仍能继续通过 startup gate 并完成提交。
  - 确认 worktree 缺少 `code/frontend/node_modules` 时，脚本会自动复用主工作区依赖或给出明确错误提示。
- Review focus:
  - review/doctor 审计与 pre-commit 是否彻底分层
  - ready-to-merge gate 状态是否会留下新的僵尸 gate
  - worktree 依赖复用是否只影响本地工具发现，不污染仓库事实源

## Execution Checklist

- [ ] Review and complete this plan with `writing-plans`
- [ ] Confirm the owning boundaries and reuse / owner decision
- [ ] Implement the first reviewable slice
- [ ] Run the planned validation for that slice
- [ ] Prepare the slice for review
