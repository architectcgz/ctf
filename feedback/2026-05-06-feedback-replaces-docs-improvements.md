# feedback 取代 docs/improvements

## 问题描述

项目里同时出现过 `docs/improvements/` 和顶层 `feedback/` 两套入口，二者都在记录 agent 发现的问题、review 流程缺口、skill/policy 改进点和可复用经验。

这会让后续 agent 不确定新反馈应该写到哪里，也会导致索引、检查脚本和 review 记忆分散。

## 原因分析

`docs/improvements/` 更像旧阶段的 agent 改进 backlog；严格 harness 初始化后，顶层 `feedback/` 已经承担“踩坑、修正、迭代心得，把失败变成可复用经验”的职责。

继续保留两个写入入口会产生重复：

- 同一类 agent 工作流反馈可能被写到两个目录。
- 迁移索引需要长期维护双向关系。
- `harness-router` 判断“产生可复用经验”后难以给出唯一落点。
- `improvement-tracker` 的默认路径和本项目 harness 结构不一致。

## 解决方案

本项目后续只保留 `feedback/` 作为 agent 反馈和可复用经验入口。

默认规则：

- Agent 工作流、review、prompt、skill、policy、项目协作方式相关反馈写入 `feedback/`。
- 历史 `docs/improvements/` 条目迁入 `feedback/` 后，不再新增 `docs/improvements/`。
- 可执行的业务任务、需求拆分或待办不写入 `feedback/`，按性质进入 `docs/plan/` 或 `docs/todo/`。
- 如果外部 skill 仍提示写入 `docs/improvements/`，在本项目内以 `feedback/` 规则覆盖。

## 收获

Harness 目录应有唯一事实源。反馈闭环只有一个默认入口，agent 才能稳定地记录、检索和复用项目经验。
