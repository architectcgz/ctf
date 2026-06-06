# harness-bridges-relocation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `ctf` 项目本地 harness adapter 从 `scripts/` 收口到 `harness/bridges/`，同时保留稳定命令入口不变。

**Architecture:** `scripts/` 继续承担用户、hook、文档直接引用的稳定入口；真正的项目本地接线逻辑迁入 `harness/bridges/`，由 wrapper 负责定位 repo root 并转发。共享 `code-workflow` 的 managed 入口仍保留在 `scripts/`，不混进项目本地 bridge 层。

**Tech Stack:** Bash, Python shared harness tooling, repo-local harness governance scripts

---

## Task Metadata

- Task Slug: `2026-06-06-harness-bridges-relocation`
- Started At: `2026-06-06T08:27:01Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-harness-bridges-relocation`
- Branch: `task/2026-06-06-harness-bridges-relocation`

## Objective And Non-Goals

- Objective:
  - 新增 `harness/bridges/` 作为项目长期提交的本地 adapter 层。
  - 迁移 `check-commit-message`、`check-skill-sync-reminder`、`install-agent-symlinks`、`uninstall-agent-symlinks` 四个项目本地 bridge。
  - 把 `scripts/` 中对应文件收成薄 wrapper，并更新 AGENTS / hook 文档 / consistency checks。
- Non-Goals:
  - 不改 shared `code-workflow` managed 脚本的 owner 边界。
  - 不迁移 `start-implementation`、`check-task-intake`、`check-startup-gate`、`archive-task-artifacts` 这类共享 workflow 入口到 `harness/bridges/`。
  - 不调整项目业务逻辑、guardrail 规则内容或 hook 执行顺序。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `works/harness-good-practices.md`
- Related architecture/contracts:
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
  - `harness/policies/commit-message.json`
- Related prior work:
  - `refactor(workflow): 将项目守卫收口到 stage plugins`
  - `refactor(harness): 统一共享 reminder 与治理入口 wrapper`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 涉及项目 harness 目录职责划分、稳定入口约定、guardrail 和文档同步。
  - 会影响多个脚本、review 守卫和协作约定，属于结构性调整。

## Files

- Create:
  - `harness/bridges/README.md`
  - `harness/bridges/check-commit-message.sh`
  - `harness/bridges/check-skill-sync-reminder.sh`
  - `harness/bridges/install-agent-symlinks.sh`
  - `harness/bridges/uninstall-agent-symlinks.sh`
- Modify:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `works/harness-good-practices.md`
  - `scripts/check-commit-message.sh`
  - `scripts/check-skill-sync-reminder.sh`
  - `scripts/install-agent-symlinks.sh`
  - `scripts/uninstall-agent-symlinks.sh`
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
- Review:
  - `harness/workflow-plugins/code-workflow/README.md`
  - `scripts/check-review-governance.sh`
- Test:
  - `scripts/check-review-governance.sh`
  - `scripts/check-consistency.sh`
  - `scripts/check-commit-message.sh`
  - `scripts/check-skill-sync-reminder.sh --working`
  - `scripts/install-agent-symlinks.sh`
  - `scripts/uninstall-agent-symlinks.sh`

## 复用与 Owner 决策

- Existing patterns searched:
  - `scripts/` 中现有薄 wrapper
  - `harness/workflow-plugins/code-workflow/` 的 stage plugin 注册方式
  - 全局 `~/.agents/harness/commit-message/` 与 `~/.agents/harness/skill-sync/`
- Reuse / extend / split / create-new decision:
  - 复用现有 `scripts/` 入口名与 hook 接线。
  - 新建 `harness/bridges/` 承载项目本地 adapter，实现“稳定入口”和“本地 bridge owner”分层。
- Owner boundary:
  - 共享 checker / reminder / workflow scaffold 仍归 `~/.agents/harness/*` owner。
  - `ctf` 只 owner 项目特有 policy、路径、skill 安装接线和 wrapper 组织。
- Why this is the narrowest safe surface:
  - 只迁移项目本地 adapter，不改共享 workflow 文件落点，不引入新入口名，不改变 hook 使用方式。

## Validation

- Commands:
  - `bash scripts/check-skill-sync-reminder.sh --working`
  - `bash scripts/check-review-governance.sh`
  - `bash scripts/check-consistency.sh`
  - `bash scripts/install-agent-symlinks.sh`
  - `bash scripts/uninstall-agent-symlinks.sh`
- Manual checks:
  - 从 repo 外直接调用上述入口时，仍能正确回到仓库根目录。
  - hook / AGENTS / consistency checks 仍引用 `scripts/` 稳定入口，而不是直接绑死 bridge 路径。
- Review focus:
  - `harness/` 与 `.harness/` 职责是否真正分开。
  - shared workflow managed 入口是否被错误迁入项目 bridge 层。
  - 守卫是否能机械发现 bridge 漂移。

## Execution Checklist

- [x] Review and complete this plan with `writing-plans`
- [x] Confirm the owning boundaries and reuse / owner decision
- [x] Implement the first reviewable slice
- [x] Run the planned validation for that slice
- [x] Prepare the slice for review
