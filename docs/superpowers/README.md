# Superpowers 过程文档索引

## 定位

本目录保存使用 superpowers 产出的实施计划和过程索引。

- `plans/`：实施计划，描述当时计划修改的文件和验证步骤。

当前需要判断“最终设计是什么”时，先读 `docs/architecture/README.md`、`docs/architecture/frontend/` 和 `docs/architecture/features/`。本目录只用于追溯设计演进和实施上下文。

## 当前作用

- `plans/`：保留实施计划，供追溯“当时准备怎么实现、怎么验证”。
- `specs/README.md`：保留迁移说明，明确专题最终设计已移入 `docs/architecture/features/`。
- `docs/architecture/features/`：承接已经采用的专题最终设计。

## 与 architecture 的关系

- `2026-04-04-community-writeup-and-recommended-solutions-design.md` 替代旧的选手 writeup 教师评阅设计。旧文件已从仓库移除，避免继续把“教师评阅学生 writeup”当成当前方向。
- `2026-04-17-awd-final-design.md` 对应的最终专题事实源位于 `docs/architecture/features/2026-04-17-awd-final-design.md`。更早的 AWD phase 草稿只用于追溯阶段决策。
- 拓扑编排器、环境模板、全局设计系统都已迁到 `docs/architecture/frontend/`，不再从旧的 `design-system/ctf-platform/` 读取。
