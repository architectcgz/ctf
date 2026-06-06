# 架构图生成输入包模板（CTF 入口）

共享正文 owner：`/home/azhi/.agents/harness/prompts/architecture-diagram-generation.md`

使用时先读取共享正文，再应用本仓库补充。

## CTF 补充

- 生成前先读 `docs/文档规范.md` 的“架构图生成规范”。
- 默认允许事实源优先从 `docs/architecture/`、`docs/contracts/`、`code/backend/`、`code/frontend/`、`scripts/` 选择。
- `docs/design/` 中未标记已落地的 Draft、`docs/reviews/` 中的历史 finding、已删除路径或 superseded 文档，默认不画成当前架构。
- 输入中的路径必须使用仓库根相对路径。
- 如果图表会进入仓库，文本源码优先保存为 `.diagram.mmd` 或放入 owning 目录的 `diagrams/` 子目录。
