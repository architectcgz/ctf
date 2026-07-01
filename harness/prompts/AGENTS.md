# harness/prompts/ — 项目 Prompt 入口

本目录保留 CTF 仓库内的 prompt 入口和项目特化补充，不再作为共享 prompt 正文 owner。

## 归属

- 共享正文：`/home/azhi/.agents/harness/prompts/`
- 本地入口：稳定仓库路径、项目参数、CTF 特化补充和交叉引用

## 当前模块

- `architecture-diagram-generation.md`：共享正文位于 `/home/azhi/.agents/harness/prompts/architecture-diagram-generation.md`
- `frontend-architecture-review.md`：CTF 前端架构审计历史入口和项目化补充；不再引用已移除的共享 code-reviewer prompt
- `network-security-review.md`：CTF 平台网络安全防护审查入口，配合 `security-vulnerability-scan` skill 使用

## 入口约定

- 跨项目可复用的 prompt 正文优先迁到 `/home/azhi/.agents/harness/prompts/`。
- 本目录只保留项目内仍需要稳定引用的入口文件和局部补充。
- 历史迁移 prompt、一次性初始化 prompt 和已沉淀为 skill 的规则不要继续保留在这里。
- 项目 prompt 也遵守共享 prompt 写作原则：先抽象问题类别，再把本次现象作为反例或正例；不要把一次事故直接写成只适用于本次的提示词。
