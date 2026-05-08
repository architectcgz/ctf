# Improvements Migration Index

## 定位

这里是历史 improvement 积累迁入严格 harness `feedback/` 后的索引。旧路径已清理；后续新增 agent 反馈默认写入 `feedback/`，业务改进项仍可按项目需要进入 `docs/plan/` 或 `docs/todo/`。

## 已迁移条目

- `feedback/2026-05-03-structural-workflow-required.md` ← 结构性改动默认走阶段化流程
- `feedback/2026-05-03-independent-review-loop.md` ← 实现后缺少强制独立 review 闭环
- `feedback/2026-05-03-directory-contract-needs-static-check.md` ← 目录规范变更需联动静态校验
- `feedback/2026-05-04-pipeline-review-gate-subagent.md` ← Pipeline review gate must use subagent
- `feedback/2026-05-05-touched-structural-debt-must-close.md` ← Touched structural debt must be closed in pipeline
- `feedback/2026-05-05-review-oversized-frontend-debt.md` ← Review touching oversized frontend files must record residual debt
- `feedback/2026-05-08-specialized-drawer-should-not-inherit-modal-template.md` ← 业务专用抽屉不应继承通用弹窗视觉模板
- `feedback/2026-05-08-agent-rules-should-not-accumulate-in-agents.md` ← Agent 规则不应持续堆进 AGENTS
- `feedback/2026-05-08-project-tasks-default-to-harness.md` ← 项目任务默认进入 Harness

## 后续规则

- Agent 工作流、review、prompt、skill、policy 类反馈优先写入 `feedback/`。
- 可执行业务任务不要堆在 `feedback/`，应进入 `docs/plan/` 或 `docs/todo/`。
- 反馈已经沉淀到 `AGENTS.md`、skill、脚本或 CI 后，在对应反馈文件中标记状态。
