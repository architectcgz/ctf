# Pipeline review gate must use subagent

## Migrated From

历史 improvement 条目 `2026-05-04-pipeline-review-gate-must-use-subagent`

## Status

not-impl

## 问题描述

AWD restart port isolation 修复中，pipeline review gate 曾在主 agent 上下文完成并归档为 review evidence；用户随后明确指出 review 不能在同一上下文里做，必须使用 subagent 或真正独立上下文。

## 原因分析

非平凡变更的 review gate 如果在主实现上下文里完成，reviewer 会带着实现过程中的假设继续判断，容易漏掉架构边界、事务语义、测试缺口和回归风险。

## 解决方案

非平凡实现的最终 review gate 必须由 subagent 或其他独立上下文执行；同上下文 review 只能作为提交前自查。若工具规则阻止自动 spawn，必须明确说明 review gate 尚未满足，而不是归档为独立 review。

## 收获

独立 review 是质量闸门，不是格式动作。上下文隔离是这道闸门成立的条件之一。
