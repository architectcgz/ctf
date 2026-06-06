# harness/prompts/ — 项目 Prompt 入口

本目录保留 CTF 仓库内的 prompt 入口和项目特化补充，不再作为共享 prompt 正文 owner。

## 归属

- 共享正文：`/home/azhi/.agents/harness/prompts/`
- 本地入口：稳定仓库路径、项目参数、CTF 特化补充和交叉引用

## 当前模块

- `architecture-diagram-generation.md`：共享正文位于 `/home/azhi/.agents/harness/prompts/architecture-diagram-generation.md`
- `coding-agent-system-prompt.md`：共享正文位于 `/home/azhi/.agents/harness/prompts/coding-agent-system-prompt.md`
- `frontend-architecture-review.md`：共享正文位于 `/home/azhi/.agents/skills/code-reviewer/frontend/architecture-review.md`

## 入口约定

- 跨项目可复用的 prompt 正文优先迁到 `/home/azhi/.agents/harness/prompts/`。
- 本目录只保留项目内仍需要稳定引用的入口文件和局部补充。
- 历史迁移 prompt、一次性初始化 prompt 和已沉淀为 skill 的规则不要继续保留在这里。
