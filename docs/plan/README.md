# 实施计划入口

## 定位

`docs/plan/` 只负责保存实现过程中的计划文档，不承载当前架构事实。

- 负责：区分当前仍在执行的 plan 和已经归档的历史 plan。
- 不负责：替代 `docs/architecture/`、`docs/contracts/` 或 `docs/design/` 的事实源职责。

## 读取顺序

1. 先判断是不是要读当前事实；如果是，优先回到 `docs/architecture/`、`docs/contracts/`、`docs/design/`。
2. 只有在需要追溯实施切片、验证步骤、当时的改动边界时，再进入本目录。
3. 当前进行中的计划读 `docs/plan/impl-plan/`。
4. 历史计划追溯读 `docs/plan/archive/impl-plan/`。

## 目录归属

- `impl-plan/`
  - 仅保留当前仍在执行、尚未归档、或本轮明确要继续追加的实施计划。
- `archive/impl-plan/`
  - 保存已完成、已 superseded、已被当前事实源吸收的历史实施计划。
  - 不作为默认读取入口，不应再被 agent 当成当前设计依据。

## 归档规则

- 当实施计划对应的结论已经回收到 `docs/architecture/`、`docs/contracts/`、`docs/design/` 或活动 `todo`，应从 `impl-plan/` 移到 `archive/impl-plan/`。
- 如果历史 review 仍需要追溯计划原文，直接通过 archive 路径访问，不再把旧 plan 留在活动目录。
- 新增或迁移计划路径时，同步更新 `docs/文档规范.md`、`docs/README.md`、`AGENTS.md` 和机械检查。
