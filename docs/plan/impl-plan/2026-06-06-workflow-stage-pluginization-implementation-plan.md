# workflow-stage-pluginization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `ctf` 现有的提交前轻量检查、完成态全量检查、review/doctor 治理审计收口为显式的 workflow stage plugin 机制，避免继续在 hook 和 wrapper 脚本里硬编码一串项目守卫命令。

**Architecture:** 保持 shared `code-workflow` 只拥有阶段模型与 task gate；项目本地通过 `harness/workflow-plugins/code-workflow/<stage>.d/*.sh` 注册具体守卫，由统一 stage runner 调度执行。

**Tech Stack:** Bash、Git hooks、项目本地 harness 脚本、既有 Go/Vitest 架构守卫脚本

---

## Task Metadata

- Task Slug: `2026-06-06-workflow-stage-pluginization`
- Started At: `2026-06-06T07:32:15Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-workflow-stage-pluginization`
- Branch: `task/2026-06-06-workflow-stage-pluginization`

## Objective And Non-Goals

- Objective:
  - 为 `ctf` 增加统一 stage runner 和项目本地 plugin 目录约定。
  - 将 pre-commit、completion、workflow-governance 三类现有守卫改为通过 stage runner 执行。
  - 保留现有守卫脚本本身，不重写 backend/frontend/review 的具体检查逻辑。
  - 更新项目文档与入口说明，使“workflow 本体”和“项目插件”边界清楚。
- Non-Goals:
  - 不修改 shared `~/.agents/harness/workflows/code-workflow/` 包的安装器和共享语义。
  - 不把 `ctf` 的具体架构守卫提升为全局 workflow 默认行为。
  - 不重构各守卫脚本的内部检查集合，只改变接线方式。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `scripts/start-implementation.sh`
  - `scripts/check-workflow-complete.sh`
  - `scripts/check-workflow-governance.sh`
  - `.githooks/pre-commit`
  - `~/.agents/skills/code-workflow/SKILL.md`
  - `~/.agents/skills/workflow-package-manager/SKILL.md`
- Related architecture/contracts:
  - `scripts/check-architecture.sh`
  - `scripts/check-backend-architecture.sh`
  - `scripts/check-frontend-architecture.sh`
  - `scripts/check-code-changes.sh`
- Related prior work:
  - `2026-06-06-commit-gate-timing-repair`
  - `scripts/check-consistency.sh -> scripts/check-workflow-governance.sh` 的治理入口拆分

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 涉及 workflow、hooks、项目 harness 和 review/completion/check 阶段边界，属于结构性流程改动。

## Files

- Create:
  - `scripts/run-workflow-stage.sh`
  - `harness/workflow-plugins/code-workflow/pre-commit-quick.d/*.sh`
  - `harness/workflow-plugins/code-workflow/completion-full.d/*.sh`
  - `harness/workflow-plugins/code-workflow/workflow-governance.d/*.sh`
  - 可能需要 `harness/workflow-plugins/code-workflow/README.md`
- Modify:
  - `.githooks/pre-commit`
  - `scripts/check-workflow-complete.sh`
  - `scripts/check-workflow-governance.sh`
  - `AGENTS.md`
  - 相关 README / harness 文档（按实际 touched surface 收口）
- Review:
  - 现有架构守卫脚本是否仍保持 project-local owner
  - stage 名称和 hook / completion / review 时机是否一致
- Test:
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - `bash scripts/run-workflow-stage.sh workflow-governance`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/check-workflow-complete.sh`
  - 视情况补 `git commit` 触发 hook 的真实链路验证

## 复用与 Owner 决策

- Existing patterns searched:
  - shared `code-workflow` 提供 startup / archive / gate 骨架
  - `ctf` 已有 `check-architecture`、`check-workflow-complete`、`check-workflow-governance` 的分层
- Reuse / extend / split / create-new decision:
  - 复用现有守卫脚本。
  - 新增一个薄的 stage runner。
  - 新增项目本地 plugin 目录来替代 wrapper/hook 里的硬编码顺序。
- Owner boundary:
  - shared `code-workflow` 继续 owner “阶段模型”。
  - `ctf` owner “每个阶段要跑哪些项目守卫”。
  - plugin runner 只负责发现、排序、执行，不理解业务检查语义。
- Why this is the narrowest safe surface:
  - 不动 shared workflow 安装器，也不改各项目守卫内部逻辑，只把接线统一到同一执行模型。

## Validation

- Commands:
  - `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - `bash scripts/run-workflow-stage.sh workflow-governance`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/check-workflow-complete.sh`
  - `git status --short`
- Manual checks:
  - 确认 pre-commit 不再直接写死多条守卫命令
  - 确认 completion 和 review stage 能从 plugin 目录发现并执行脚本
- Review focus:
  - stage/plugin 边界是否稳定
  - 是否引入新的脚本递归或重复执行
  - 文档是否把“workflow 本体 vs 项目插件”讲清楚

## Execution Checklist

- [x] Review and complete this plan with `writing-plans`
- [x] Confirm the owning boundaries and reuse / owner decision
- [x] Implement stage runner and plugin directory skeleton
- [x] Rewire pre-commit / completion / review to use stage plugins
- [x] Update docs and AGENTS references for the plugin model
- [x] Run the planned validation for this slice
- [x] Prepare the slice for review
