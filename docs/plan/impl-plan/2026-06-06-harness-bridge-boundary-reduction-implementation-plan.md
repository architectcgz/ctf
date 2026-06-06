# harness-bridge-boundary-reduction Implementation Plan

**Goal:** 把 `ctf` 项目里被泛化成常态层的 `harness/bridges/` 收回，只保留 `code-workflow` 这类真正需要项目适配的 workflow bridge，避免普通稳定入口、项目检查和全局 harness 调用之间继续多长一层目录。

**Architecture:** `scripts/` 顶层继续作为稳定入口；项目本地机械检查主体继续留在 `harness/checks/`；共享 workflow 的项目适配继续放在 `harness/workflow-plugins/code-workflow/`。其余很薄的项目命令直接在稳定入口落实现，不再额外抽象成通用 `bridges/` 层。

**Tech Stack:** Bash scripts, project harness checks, Git hook docs, AGENTS governance docs

---

## Task Metadata

- Task Slug: `2026-06-06-harness-bridge-boundary-reduction`
- Started At: `2026-06-06T12:38:29Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `main`

## Objective And Non-Goals

- Objective:
  - 删除 `harness/bridges/` 这一层对普通项目命令的承载。
  - 把 `check-commit-message`、`check-skill-sync-reminder`、`install-agent-symlinks`、`uninstall-agent-symlinks` 收回 `scripts/` 顶层稳定入口。
  - 更新 `AGENTS.md`、hook 文档和 consistency 守卫，让“只有 workflow 适配才需要 bridge”变成明确约束。
- Non-Goals:
  - 不改 `code-workflow` 的 stage runner 和 `harness/workflow-plugins/code-workflow/` 结构。
  - 不修改业务代码、前后端架构守卫或全局 `~/.agents/harness` 共享实现。
  - 不处理项目内其他历史 plan 的归档整理。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `harness/workflow-plugins/code-workflow/README.md`
- Related architecture/contracts:
  - `scripts/check-commit-message.sh`
  - `scripts/check-skill-sync-reminder.sh`
  - `scripts/install-agent-symlinks.sh`
  - `scripts/uninstall-agent-symlinks.sh`
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-06-harness-bridges-relocation-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-05-workflow-governance-refactor-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达项目 harness 分层、稳定脚本入口、hook 文档和治理守卫。
  - 需要把之前已经落地的分层决策收回并重新定义 owner 边界。

## Files

- Create:
  - `docs/plan/impl-plan/2026-06-06-harness-bridge-boundary-reduction-implementation-plan.md`
- Modify:
  - `AGENTS.md`
  - `.githooks/README.md`
  - `scripts/check-commit-message.sh`
  - `scripts/check-skill-sync-reminder.sh`
  - `scripts/install-agent-symlinks.sh`
  - `scripts/uninstall-agent-symlinks.sh`
  - `scripts/lib/check-consistency/navigation.sh`
  - `scripts/lib/check-consistency/architecture.sh`
- Review:
  - `harness/workflow-plugins/code-workflow/README.md`
  - `harness/policies/script-layer-manifest.json`
- Test:
  - `bash scripts/check-script-layer.sh`
  - `bash scripts/check-workflow-governance.sh`
  - `bash scripts/check-commit-message.sh <tmp-file>`

## 复用与 Owner 决策

- Existing patterns searched:
  - 当前 `scripts/` wrapper 结构
  - `harness/bridges/` 四个脚本的真实逻辑
  - `harness/workflow-plugins/code-workflow/` 的 workflow 适配边界
- Reuse / extend / split / create-new decision:
  - 复用现有稳定入口脚本名，不再保留额外 bridge 目录。
  - 保留 `harness/checks/` 与 `harness/workflow-plugins/code-workflow/` 作为仍然有明确 owner 的两类下沉层。
  - 删除已经没有必要的通用 `harness/bridges/` 层。
- Owner boundary:
  - `scripts/check-*` / `install-*` / `uninstall-*`：稳定入口，可直接承载足够薄的项目逻辑。
  - `harness/checks/`：项目本地确定性检查主体。
  - `harness/workflow-plugins/code-workflow/`：共享 workflow 的项目适配层。
- Why this is the narrowest safe surface:
  - 不改入口名，不影响 hook 和操作者命令习惯，只收回一层没有独立价值的目录。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 任务核心是分层和 owner 边界调整，不是功能 bug 修复。
- grill-with-docs findings:
  - 仓库内真正需要项目适配语义的是 `code-workflow` stage/plugin 体系，而不是 `check-commit-message` 或 `check-skill-sync-reminder` 这类薄命令。
  - 现有 `AGENTS.md`、`.githooks/README.md`、`check-consistency` 仍把 `harness/bridges/` 当成常态层，这是当前主要漂移点。
  - `start-implementation.sh` 仍默认创建新 worktree，和“main 干净可直接作为隔离上下文”的新约定存在文案/工具漂移，本轮至少先在项目规则里收口文案。
- Plan adjustments after challenge:
  - 不再保留空壳 `harness/bridges/README.md`，直接让守卫校验该层不存在。
  - 除了脚本迁位，也同步把“main 干净可直接绑定 task”写回项目规则，减少后续 workflow 误导。

## Validation

- Commands:
  - `bash scripts/check-script-layer.sh`
  - `bash scripts/check-workflow-governance.sh`
  - `tmp="$(mktemp)" && printf 'refactor(harness): 收口 bridge 边界\n\n把普通项目脚本收回稳定入口。\n同步更新守卫与文档。\nTask: 2026-06-06-harness-bridge-boundary-reduction\n' > "$tmp" && bash scripts/check-commit-message.sh "$tmp" && rm -f "$tmp"`
- Manual checks:
  - `scripts/check-commit-message.sh` 不再转发到 `harness/bridges/`
  - `.githooks/README.md` 与 `AGENTS.md` 不再把 `harness/bridges/` 作为默认层
  - `check-workflow-governance` 能机械发现 `harness/bridges/` 漂移回归
- Review focus:
  - workflow 适配层与普通项目脚本的边界是否已经清楚
  - 是否还残留对 `harness/bridges/` 的路径依赖或文案依赖
