# feedback/ — 反馈记录

实践中的踩坑、修正、迭代心得。把失败变成可复用经验。

## 入口规则

- 本项目只使用 `feedback/` 记录 agent 工作流、review、prompt、skill、policy 和项目协作方式相关反馈。
- 历史 `docs/improvements/` 条目迁入 `feedback/` 后，不再新增 `docs/improvements/`。
- 如果外部 skill 仍提示写入 `docs/improvements/`，在本项目内以本文件规则覆盖。
- 可执行的业务任务、需求拆分或待办不写入 `feedback/`，按性质进入 `docs/plan/` 或 `docs/todo/`。

## 文件约定

- 文件名：`{日期}-{简述}.md`
- 结构：问题描述 → 原因分析 → 解决方案 → 收获
- 如果反馈导致 prompts、concepts、脚本或 AGENTS 更新，必须交叉链接。

## 已迁移积累

- `improvements-index.md`：历史 improvement 条目迁移索引。
- `2026-05-03-structural-workflow-required.md`：结构性改动阶段化流程。
- `2026-05-03-independent-review-loop.md`：实现后独立 review 闭环。
- `2026-05-03-directory-contract-needs-static-check.md`：目录契约变更联动静态校验。
- `2026-05-04-pipeline-review-gate-subagent.md`：pipeline review gate 必须独立上下文。
- `2026-05-05-touched-structural-debt-must-close.md`：触达结构债必须在流水线中收口。
- `2026-05-05-review-oversized-frontend-debt.md`：触达超大前端文件时记录结构债。
- `2026-05-06-feedback-replaces-docs-improvements.md`：feedback 取代 docs/improvements。
- `2026-05-09-do-not-bend-tests-to-fit-broken-architecture.md`：测试不能迁就坏架构，测试失败时先回到 owner / contract / 架构分析。
