# commit-message-governance-extract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把提交信息检查里的通用机制抽到全局 `~/.agents/harness`，同时让 `ctf` 改成项目 policy + wrapper，避免共享脚本继续携带 `ctf` 特化常量，并修正正文统计把 `Task:` 元数据算作业务说明的问题。

**Architecture:** 全局 harness 负责通用 `commit-msg` 解析与校验引擎，项目仓库负责本地提交 policy、active task slug 的解析入口以及 hook / 文档接线；初始化器负责把这套分层继续生成到后续新项目，而不是继续吐出内联脚本。

**Tech Stack:** Bash scripts, Python 3 stdlib, Git hooks, JSON policy files, AGENTS / hook docs

---

## Task Metadata

- Task Slug: `2026-06-06-commit-message-governance-extract`
- Started At: `2026-06-06T06:31:54Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract`
- Branch: `task/2026-06-06-commit-message-governance-extract`

## Objective And Non-Goals

- Objective:
  - 在 `~/.agents/harness` 新增可复用的提交信息检查器。
  - 让 `ctf/scripts/check-commit-message.sh` 退化为薄 wrapper，并把项目策略落到 `harness/policies/commit-message.json`。
  - 更新 `ctf` 的 hook 文档、AGENTS 说明和一致性检查，确保新的分层可发现、可验证。
  - 同步更新 `~/.agents/harness/harness-initializer.py`，避免新项目继续生成旧版内联脚本。
- Non-Goals:
  - 不重做 `pre-commit` 轻量门禁结构。
  - 不把 `ctf` 的中文描述、类型白名单或文案示例硬编码回全局检查器。
  - 不在这次任务里扩展到完整 CI 提交流程治理。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `docs/plan/impl-plan/2026-06-05-workflow-governance-refactor-plan.md`
- Related architecture/contracts:
  - `.githooks/commit-msg`
  - `scripts/check-commit-message.sh`
  - `scripts/check-startup-gate.sh`
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
  - `/home/azhi/.agents/harness/harness-initializer.py`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-05/2026-05-31-commit-message-detail-enforcement-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达全局 harness、项目 hook、治理检查和项目文档。
  - 需要重新划分 shared mechanism 与 project policy 的 owner 边界。

## Files

- Create:
  - `/home/azhi/.agents/harness/commit-message/check_commit_message.py`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/harness/policies/commit-message.json`
- Modify:
  - `/home/azhi/.agents/harness/harness-initializer.py`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/scripts/check-commit-message.sh`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/.githooks/README.md`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/AGENTS.md`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/scripts/lib/check-consistency/navigation.sh`
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/scripts/lib/check-consistency/architecture.sh`
- Review:
  - `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-06-commit-message-governance-extract/.githooks/commit-msg`
- Test:
  - `bash scripts/check-consistency.sh`
  - `bash scripts/check-commit-message.sh <tmp-message>`
  - direct invocation of the shared checker with valid / invalid samples

## 复用与 Owner 决策

- Existing patterns searched:
  - `ctf` 现有 `scripts/check-commit-message.sh`
  - `~/.agents/harness/harness-initializer.py` 中生成的旧脚本模板
  - `ctf` 当前 hook 与 consistency 接线
- Reuse / extend / split / create-new decision:
  - 复用现有项目 hook 入口名和用户交互文案风格，但把通用提交解析 / 结构校验下沉到共享 Python 检查器。
  - 项目侧新增独立 policy 文件，wrapper 只负责定位全局 harness、读取本地 policy、传入 active task slug。
- Owner boundary:
  - 全局 harness owner：提交标题结构、正文结构、`Task:` 绑定机制和 policy 解释。
  - `ctf` owner：允许的 type、中文描述要求、正文阈值、本地错误示例、startup gate slug 来源和仓库内文档。
- Why this is the narrowest safe surface:
  - 不改 hook 入口和提交习惯，只把重复且应共享的机制抽出来，同时保留项目策略可独立收紧或放宽。

## Validation

- Commands:
  - `bash scripts/check-consistency.sh`
  - `bash scripts/check-commit-message.sh "$tmpfile"`
  - `python3 /home/azhi/.agents/harness/commit-message/check_commit_message.py --message-file "$tmpfile" --policy-file "$policy"`
- Manual checks:
  - 确认 `Task:` 行不再计入正文最少说明行和字符数。
  - 确认缺少全局 harness 时 wrapper 会给出明确安装提示，而不是静默失败。
- Review focus:
  - shared mechanism 与 project policy 是否彻底分层
  - 初始化器是否跟随新结构，避免新项目继续回归旧实现
  - `ctf` 文档和 consistency 检查是否能机械发现这套新接线

## Execution Checklist

- [ ] Review and complete this plan with `writing-plans`
- [ ] Confirm the owning boundaries and reuse / owner decision
- [ ] Implement the first reviewable slice
- [ ] Run the planned validation for that slice
- [ ] Prepare the slice for review
