# Touched structural debt must be closed in pipeline

## Migrated From

历史 improvement 条目 `2026-05-05-touched-structural-debt-must-be-closed-in-pipeline`

## Status

agent-recorded

## 问题描述

当任务触达已记录的结构性技术债表面，例如超大 owner-mixed 前端组件，不能继续把债务写成 follow-up 后合并新功能。

## 原因分析

如果 pipeline 允许“已知这里有结构债，但这次先把功能塞进去”，review 会越来越像补记账，而不是风险闸门。超大组件、owner 混杂页面、已知待拆模块会被每次变更继续增厚，最后既难验证也难拆分。

## 解决方案

把规则前移到 intake、plan review 和 final review 三个阶段：

- intake 必须识别本次是否触达已知结构债表面。
- implementation plan 必须写明债务在本次流水线中的收口标准。
- final review 若发现 touched debt 仍然存在，直接判 blocker。

## 收获

这条规则已经记录到工作区 `AGENTS.md` 和 pipeline/review skills。后续迁移 harness 时应把它当成高优先级反馈规则保留。
