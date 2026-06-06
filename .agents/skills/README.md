# Shared Skills

本目录保存当前仓库对 `Claude` 与 `Codex` 共享的 skill 源，不再把跨 agent 可复用的 workflow 只留在 `~/.codex/skills/`。

## 目标

- 共享仓库级 workflow，而不是共享某个 agent 的私有配置目录。
- 让 `Claude` 通过 `.claude/skills -> ../.agents/skills` 直接读取。
- 让 `Codex` 通过 `bash scripts/install-agent-symlinks.sh` 安装到 `~/.codex/skills/`。

## 当前项目特有 / 项目补充

- `harness-engineering`
- `project-template`
- `harness-router`
- `documentation-architecture`
- `improvement-tracker`
- `grillme`
- `brainstorming`
- `dispatching-parallel-agents`
- `executing-plans`
- `finishing-a-development-branch`
- `receiving-code-review`
- `requesting-code-review`
- `subagent-driven-development`
- `systematic-debugging`
- `test-driven-development`
- `using-git-worktrees`
- `using-superpowers`
- `verification-before-completion`
- `writing-plans`
- `writing-skills`
- `ctf-go-backend-architecture-reference`

## 当前保留的项目镜像 / 项目入口

以下 skill 在 `~/.agents/skills/` 已经有全局主体，但项目里仍保留入口，目的是让仓库内的 agent 发现路径稳定，并补充本仓库语境：

- `requirements-analyst`
- `runtime-ops-safety`
- `development-pipeline`
- `code-reviewer`
- `frontend-engineer`
- `go-backend`
- `onion-clean-architecture`
- `security-vulnerability-scan`

这些 skill 的共同特点是：

- 主要描述流程、入口、检查和模板
- 对当前仓库的 harness 形态有直接复用价值
- 经过最小去全局路径化后，能够被 `Claude` 和 `Codex` 共用
- `superpowers` 相关 skill 在仓库共享层中统一按扁平目录名保存，不再依赖全局私有目录或命名空间桥接
- 已经拥有全局主体的 skill，在项目层逐步收口成“项目入口 / 项目补充”，而不是继续长期复制完整通用正文

## 暂不迁入

以下类型先保留在 `~/.codex/skills/`：

- 明显依赖 `Codex` 专属工具或全局 prompt 约定的 skill
- 强烈绑定个人实验环境或绝对路径的 skill
- 只适合个人工作方式、暂时不应进入项目事实源的 skill
- 设计/领域类大包 skill，在没有确认跨项目价值前不进入仓库

## 安装

```bash
bash scripts/install-agent-symlinks.sh
```

默认行为：

- 保持 `.claude/skills -> ../.agents/skills`
- 在 `~/.codex/skills/` 下创建 `ctf-<skill>` 形式的软链接

卸载：

```bash
bash scripts/uninstall-agent-symlinks.sh
```

默认只移除 `~/.codex/skills/` 下的项目级软链接，保留仓库内 `.claude/skills -> ../.agents/skills`，因为后者属于项目事实源的一部分。
