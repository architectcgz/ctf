# 历史实施计划归档

2026-06-01 批量归档：前端架构迁移冲刺（5/24 - 6/1）期间生成的 270 个实施计划。
迁移目标（路由收口、shared→features 边界、guardrail 覆盖、实体层补强、owner 拆分）已全部达成。

2026-06 之后，`archive/impl-plan/` 也会保留少量仍被活动 task group、review 或实现收尾引用的已归档 plan 快照；它们不再作为活动计划维护，但比只留 Git 历史更方便追溯上下文。

2026-06-24 批量归档：将 `docs/plan/impl-plan/` 中 2026-06-01 至 2026-06-21 的历史实施计划迁入 `2026-06/`，包括已完成的后端拆分、runtime、HA、错误治理、前端 owner 收口和 task group 计划。当前活动目录只保留 `README.md` 与仍在执行的 2026-06-24 实施计划。

- `2026-05/` — 5/24-5/31: decomposition、shell-convergence、feature-ui-normalization、contract-naming-neutralization 等
- `2026-06/` — 6/1-6/21: owner-cleanup、data-split、state-split、backend boundary、runtime、HA、error management 等历史实施计划和 task group 快照

如需追溯未保留正文的已归档计划，请查看 Git 历史。
当前仍在执行的计划请回到 `../../impl-plan/`。
