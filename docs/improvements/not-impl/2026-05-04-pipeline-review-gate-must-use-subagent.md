# Pipeline review gate must use subagent

## Status

not-impl

## Context

Observed on 2026-05-04 during AWD restart port isolation fix: the pipeline review gate was performed in the main agent context and archived as review evidence, but the user clarified that review cannot be in the same context and must use a subagent. Durable rule needed: for non-trivial pipeline/code-review gates, spawn a subagent or otherwise use a truly independent context; same-context review may be a self-check only, not the gate.

## Problem

Pipeline 的非平凡变更 review gate 如果在主实现上下文里完成，reviewer 会带着实现过程中的假设继续判断，容易漏掉架构边界、事务语义、测试缺口和回归风险。这样的 review 可以算 self-check，但不能算独立 review gate。

## Suggested Direction

把规则沉淀到对应的 pipeline / code review / agent policy 中：非平凡实现的最终 review gate 必须由 subagent 或其他独立上下文执行；同上下文 review 只能作为提交前自查。若工具规则阻止自动 spawn，必须明确向用户说明 review gate 尚未满足，而不是归档为独立 review。

## Target Owner

- skill: `development-pipeline`, `code-reviewer`
- agent: main coding agent policy
- docs: project `AGENTS.md` 可补项目级硬规则
- code area: none

## Evidence

- file: `docs/reviews/backend/2026-05-04-ctf-review-awd-restart-port-isolation.md`
- command: subagent `019df283-6d52-7d82-a400-477b92ed3151` was later spawned to perform the missing independent review.
- behavior: user identified that the initial review was same-context and therefore invalid as a gate.

## Decision Log

- 2026-05-04: Created.
