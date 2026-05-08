# Agent 规则不应持续堆进 AGENTS

## Status

agent-recorded

## 问题描述

通知中心抽屉修复后，为避免后续继续误用 `ModalTemplateShell`，曾把 overlay 选型细则直接写入根 `AGENTS.md`。这能提醒 agent，但也会让入口文件逐渐变成细则堆叠区，降低可读性，并和 harness 的分层目标冲突。

根 `AGENTS.md` 应主要承担入口导航、少量高频硬规则和目录归属说明；具体组件契约、架构事实、踩坑复盘和机械不变量应分别进入对应事实源。

## 原因分析

这次错误的根因是把“需要让以后不要再犯”直接等同于“写进 AGENTS.md”。在 harness 架构里，持久化知识需要先分类：

- 当前架构事实进入 `docs/architecture/`。
- agent 工作方式、踩坑和纠偏进入 `feedback/`。
- 可机械判断的不变量进入测试、脚本或 CI。
- 根 `AGENTS.md` 只保留导航和漏掉会反复造成事故的高层硬规则。

如果把每个局部组件规则都写入根入口，后续 agent 会读到越来越多低层细节，却更难判断真正的事实源在哪里。

## 解决方案

后续新增规则时先问它属于哪一种资产：

- 组件、页面、接口或模块的最终事实：写入对应 `docs/architecture/`、`docs/contracts/` 或页面设计文档。
- 某次排错暴露的 agent 协作问题：写入 `feedback/`，并按需更新索引。
- 能用代码检查的约束：优先落到测试、lint、脚本或 `scripts/check-consistency.sh`。
- 只有跨任务、跨领域、遗漏后会反复出事故的入口级规则，才考虑写入根 `AGENTS.md`。

通知中心 overlay 这类具体组件选型，不再放在根 `AGENTS.md`；它的事实源是 `docs/architecture/frontend/06-components.md`，防回归依赖 `ModalTemplates.test.ts` 里的静态架构测试，复盘记录留在 `feedback/`。

## 收获

Harness 的目标不是把所有知识集中到一个文件，而是让每类知识有明确归属。`AGENTS.md` 应保持导航性和入口性，避免成为新规则的默认收纳箱。
