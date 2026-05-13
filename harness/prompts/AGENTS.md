# harness/prompts/ — Harness Prompt Assets

本目录是仓库内 harness prompt 的唯一事实源。

## 当前模块

- `architecture-diagram-generation.md`：根据当前事实源生成可 review 的架构图输入包。
- `coding-agent-system-prompt.md`：reuse-first coding agent 约束。

## 入口约定

- 本目录只保留已经验证过、会复用的项目级 agent 工作流 prompt。
- 历史迁移 prompt、一次性初始化 prompt 和已沉淀为 skill 的规则不要继续保留在这里。
