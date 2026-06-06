# Workflow Governance Refactor Plan

**Goal:** 把当前 CTF 仓库里混在一起的安装检查、编码前 gate、提交前轻量守卫和 review/doctor 治理审计重新分层，避免 `check-consistency.sh` 与旧 reuse-first 语义继续漂移。

**Architecture:** 全局层负责共享 workflow 模型与 agent 入口桥接，CTF 项目层负责本地受保护 surface、hook 接线、计划落点和治理审计入口；agent 负责编排，脚本 / hook 负责 enforcement。

**Tech Stack:** Bash scripts, Git hooks, AGENTS / harness docs, local `.agents` bridge, startup gate metadata

---

## Task Metadata

- Task Slug: `2026-06-05-workflow-governance-refactor`
- Started At: `2026-06-05T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `task/2026-06-05-workflow-governance-refactor`

## Objective And Non-Goals

- Objective:
  - 把 `check-consistency.sh` 从所有提交前的统一门禁，收回到 review / doctor 级治理审计。
  - 用 `start-implementation` / `check-startup-gate` 建立非琐碎任务的统一入口和本地 gate。
  - 收口 `CLAUDE.md`、`.claude/skills`、`.agents/skills`、agent entrypoint 校验和安装脚本的 owner 边界。
- Non-Goals:
  - 不一次性实现完整 CI 编排。
  - 不把所有任务都重新迁到新 worktree；这里只收口本地治理入口和 enforcement 语义。
  - 不把 CTF 项目特化的规则塞回全局 skill 正文。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/README.md`
  - `.githooks/README.md`
- Related architecture/contracts:
  - `.githooks/pre-commit`
  - `scripts/check-consistency.sh`
  - `scripts/check-task-intake.sh`
  - `scripts/start-implementation.sh`
  - `scripts/check-startup-gate.sh`
  - `scripts/install-agent-symlinks.sh`
  - `harness/checks/common.py`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-05-workflow-governance-refactor-plan.md`
  - `.agents/skills/`
  - `CLAUDE.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 hook、脚本、harness 文档、agent 入口桥接和本地 workflow 机制。
  - 需要明确区分安装层、启动 gate、提交前轻量守卫和 review 审计层。

## Files

- Create:
  - `.agents/skills/**`
  - `scripts/check-agent-entrypoints.sh`
  - `scripts/check-shared-skills.sh`
  - `scripts/install-agent-symlinks.sh`
  - `scripts/uninstall-agent-symlinks.sh`
  - `scripts/check-startup-gate.sh`
  - `scripts/start-implementation.sh`
  - `scripts/lib/check-consistency/*.sh`
  - `harness/checks/change_detection.py`
  - `harness/checks/check_startup_gate.py`
  - `harness/templates/implementation-plan-skeleton.md`
- Modify:
  - `AGENTS.md`
  - `docs/README.md`
  - `.githooks/pre-commit`
  - `.githooks/README.md`
  - `scripts/check-consistency.sh`
  - `scripts/check-task-intake.sh`
  - `scripts/check-commit-message.sh`
  - `scripts/doctor-local-harness.sh`
  - `.gitignore`
- Review:
  - `.claude/skills`
  - `CLAUDE.md`
  - `harness/prompts/AGENTS.md`
- Test:
  - `bash scripts/check-consistency.sh`
  - hook path on staged governance files

## 复用与 Owner 决策

- Existing patterns searched:
  - 现有 `scripts/check-consistency.sh`
  - 现有 `.githooks/pre-commit`
  - 当前全局 `.agents` / `.claude` / `CLAUDE.md` 入口约定
- Reuse / extend / split / create-new decision:
  - 保留现有治理脚本入口名，但按职责拆分 `check-consistency` 内部实现。
  - 新增 `start-implementation` / `check-startup-gate` 作为共享 workflow 在项目内的本地实现面。
  - 复用 `.agents/skills` 作为项目级共享 skill owner，并通过 `.claude/skills` 软链接读取。
- Owner boundary:
  - `check-consistency.sh` 只做治理审计。
  - `pre-commit` 只做与当前 staged diff 强相关的轻量守卫。
  - `install-agent-symlinks.sh` 负责 Codex 侧 skill 安装与 agent 入口桥接。
- Why this is the narrowest safe surface:
  - 不重写所有治理脚本，而是在现有入口下收口时机和 owner，减少后续继续双轨漂移。

## Validation

- Commands:
  - `bash scripts/check-consistency.sh`
  - staged governance commit 自带的 `pre-commit`
- Manual checks:
  - `CLAUDE.md -> AGENTS.md`
  - `.claude/skills -> ../.agents/skills`
  - `scripts/install-agent-symlinks.sh` 与 `scripts/check-agent-entrypoints.sh` 的路径契约
- Review focus:
  - shared workflow 与项目本地治理层是否分层清楚
  - 旧 reuse-first 提交前门禁是否已正确退场
