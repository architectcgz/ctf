# reuse-first harness

## 问题描述

前端页面、hook、API wrapper 和目录型组件已经出现多条成熟实现路径，但 agent 在没有先检索既有实现的情况下，仍然容易直接新建一套并行实现。

## 原因分析

- “先复用再新建”长期只停留在 prompt 语义，没有形成提交前和 PR 阶段都能执行的硬约束。
- 仓库中已有 `views/`、`components/*Page.vue`、`features/*/model`、`composables` 多种入口，没有模式索引时，agent 不知道当前项目认定的标准写法。
- 只要求“写个计划”不够，必须留下本轮 `Reuse Decision` 证据，并让脚本去检查有没有绕过。

## 解决方案

- 新增 `harness/policies/reuse-first.yaml` 和 `harness/policies/project-patterns.yaml`，把受保护创建面和项目标准模式机器可读化。
- 新增 `.harness/reuse-decision.md`，要求在受保护改动前先写复用决策。
- 新增 `harness/checks/*` 和 `scripts/check-reuse-first.sh`，让本地 pre-commit 与手工自检都能拒绝未做复用检索的新实现。
- 在 `AGENTS.md` 中补上 `Classify -> Search -> Decide -> Implement` 的硬步骤。

## 收获

对这类“不要再新增平行实现”的规则，最有效的方式不是继续堆 prompt，而是同时补：

1. 模式索引
2. 决策证据
3. 提交前脚本
4. 本地 hook / 本地脚本复查

这样 agent 即使知道原则，也很难绕开执行闭环。
