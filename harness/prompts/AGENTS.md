# harness/prompts/ — Harness Prompt Assets

本目录是仓库内 harness prompt 的唯一事实源。

## 当前模块

- `harness-router.md`：仓库任务默认先进入 harness intake。
- `experience-extraction-closeout.md`：任务收尾前必须判断是否需要沉淀经验。
- `ctf-harness-initialization.md`：严格参考 harness 结构初始化项目。
- `architecture-diagram-generation.md`：根据当前事实源生成可 review 的架构图输入包。
- `ctf-ui-theme-system-skill.md`：CTF UI theme system 的仓库内 prompt 入口。
- `coding-agent-system-prompt.md`：reuse-first coding agent 约束。

## 入口约定

- 顶层 `prompts/` 目录只保留目录导航，不再承载这些 prompt 文件本身。
- 任何新增 harness prompt，都应直接进入 `harness/prompts/`，不要再回写到顶层 `prompts/`。
