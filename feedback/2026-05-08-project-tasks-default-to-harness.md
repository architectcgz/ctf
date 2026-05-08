# 项目任务默认进入 Harness

## Status

agent-recorded

## 问题描述

此前处理通知中心和弹窗抽象问题时，曾出现没有先走 harness intake 的情况。结果是直接进入局部实现和规则追加，后续才发现规则归属、反馈记录和机械约束没有按 harness 分层处理，导致方向跑偏。

这个问题不是某个组件实现错误，而是任务入口缺少强制路由：项目已经有 `harness-router` skill 和 `prompts/harness-router.md`，但根 `AGENTS.md` 没有明确写出“默认先进入 harness”。

## 原因分析

Harness 的价值在任务开始阶段，而不是收尾阶段。若一开始没有判断 `SIMPLE` / `HARNESS`，agent 容易直接按普通代码任务推进：

- 没有先读项目事实源和目录归属。
- 没有判断是否需要计划、review、验证和反馈沉淀。
- 容易把局部规则写进错误位置，例如把组件细则塞进 `AGENTS.md`。
- 后续再补 harness 时，只能修正结果，不能避免过程偏移。

## 解决方案

根 `AGENTS.md` 增加项目入口规则：

- 本仓库默认先进入 harness。
- 只有明显简单、局部、可逆且不需要经验沉淀的任务，才允许判定为 `SIMPLE`。
- `HARNESS` 任务开始时先读相关 harness 入口，再决定计划、review、验证和沉淀位置。
- 可复用 prompt 继续放在 `prompts/harness-router.md`，入口文件只做导航和硬约束。

## 收获

默认 harness 应该发生在任务开始前。它不是文档归档动作，而是防止任务从入口处跑偏的路由机制。
