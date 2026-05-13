# Planning Archive Migration Index

## Migrated From

历史 planning 归档

## 定位

历史 planning 归档包含 `task_plan.md`、`progress.md`、`findings.md`。迁入严格 harness 后，它属于 `practice/` 的历史实验记录，而不是当前架构事实源。

## 分类

- `*-layering-*`、`*-crossdeps-*`：分层与跨依赖治理实践。
- `contest-*`、`contest-awd-*`：竞赛与 AWD 模块拆分实践。
- `runtime-*`、`runtimeinfra-*`：运行时与基础设施重排实践。
- `identity-*`、`auth-*`：身份与认证边界实践。
- `practice-*`、`assessment-*`、`challenge-*`、`ops-*`：子域拆分和组合实践。

## 使用规则

- 需要追溯“当时如何拆任务”时，查看 Git 历史中的 planning 归档或本文件的分类索引。
- 需要追溯“执行进度和阶段状态”时，查看 Git 历史中的 progress 记录。
- 需要追溯“发现了什么结构问题”时，查看 Git 历史中的 findings 记录。
- 当前事实和新增实施计划仍应写入 `docs/plan/impl-plan/` 或 `docs/architecture/`。

## Harness 结论

历史 planning 归档的价值是实践记忆，不是事实源。后续如果某个阶段经验会重复影响 agent 行为，应提炼到 `feedback/`、`concepts/`、`harness/prompts/` 或 `scripts/check-consistency.sh`。
