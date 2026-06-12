# 实施计划入口

## 定位

`docs/plan/` 负责保存实现过程中的计划文档，区分正式实施计划和探索性计划。

- 负责：区分正式实施计划、探索性计划和历史归档 plan。
- 不负责：替代 `docs/architecture/`、`docs/contracts/` 或 `docs/design/` 的事实源职责。

## Plan 分层

本项目区分两类计划：

### 1. 正式实施计划（`impl-plan/`）

用于结构性改动、跨模块重构、需要正式 review 和 task gate 的实现方案。

**特征**：
- 必须绑定 task slug 和 startup gate
- 需要正式 code-workflow 和独立 review
- 完成后结论回收到 `docs/architecture/` 或 `docs/contracts/`
- 需要长期追溯

**归档**：
- 完成后移到 `docs/plan/archive/impl-plan/YYYY-MM/`
- 默认通过 Git 历史追溯，仅在仍被活动引用时保留正文快照

### 2. 探索性计划（`exploratory/`）

用于快速起草、技术验证、原型探索和临时调研。

**特征**：
- 不绑定 task gate
- 不需要正式 review
- 可以不完整，只记录思路
- 生命周期短

**归档**：
- 完成后提取有价值结论到对应事实源
- plan 本身移到 `archive/exploratory/` 或直接删除
- 不进入正式 plan 归档流程

详见 `exploratory/README.md`。

## 读取顺序

1. 先判断是不是要读当前事实；如果是，优先回到 `docs/architecture/`、`docs/contracts/`、`docs/design/`。
2. 判断是正式实施计划还是探索性计划：
   - 正式实施计划读 `impl-plan/`
   - 探索性计划读 `exploratory/`
3. 旧历史计划默认优先看 Git 历史；如果某条已归档 plan 仍被活动引用，可以从 `archive/impl-plan/` 或 `archive/exploratory/` 读取归档快照。

## 目录归属

- `impl-plan/`
  - 仅保留当前仍在执行、尚未归档、或本轮明确要继续追加的实施计划。
  - 允许按 task group 子目录组织同一 task group 的活动切片计划和进度索引。
- `archive/impl-plan/`
  - 默认优先通过 Git 历史追溯旧计划。
  - 对仍被活动 task group、review 或实现收尾引用的少量已归档计划，可以保留正文快照。
  - 不作为默认读取入口，不应再被 agent 当成当前设计依据。

## 归档规则

- 当实施计划对应的结论已经回收到 `docs/architecture/`、`docs/contracts/`、`docs/design/` 或活动 `todo`，应从 `impl-plan/` 移到 `archive/impl-plan/`。
- 如果历史 review 或活动 task group 仍需要追溯计划原文，可先移到 `archive/impl-plan/` 保留快照；不再继续作为活动计划维护。
- 新增或迁移计划路径时，同步更新 `docs/文档规范.md`、`docs/README.md`、`AGENTS.md` 和机械检查。
