# Improvements Migration Index

## 定位

这里是历史 improvement 积累迁入严格 harness `feedback/` 后的索引。旧路径已清理；后续新增 agent 反馈默认写入 `feedback/`，业务改进项仍可按项目需要进入 `docs/plan/` 或 `docs/todo/`。

## 已迁移条目

- `feedback/2026-05-03-independent-review-loop.md` ← 实现后缺少强制独立 review 闭环
- `feedback/2026-05-03-directory-contract-needs-static-check.md` ← 目录规范变更需联动静态校验
- `feedback/2026-05-08-specialized-drawer-should-not-inherit-modal-template.md` ← 业务专用抽屉不应继承通用弹窗视觉模板
- `feedback/2026-05-08-agent-rules-should-not-accumulate-in-agents.md` ← Agent 规则不应持续堆进 AGENTS
- `feedback/2026-05-10-top-tab-panels-should-not-repeat-eyebrow.md` ← 带顶部 tab 的页面不应在 tab 面板内重复渲染 eyebrow
- `feedback/2026-05-11-shared-modal-visible-border-owned-by-slot-root.md` ← 共享弹窗可见边框应由 slot 内容根承接，`:deep` 只用于结构层命中

## 已归位后移除

- `feedback/2026-05-03-structural-workflow-required.md` -> 全局 AGENTS / `development-pipeline`
- `feedback/2026-05-04-pipeline-review-gate-subagent.md` -> `development-pipeline`
- `feedback/2026-05-05-touched-structural-debt-must-close.md` -> 全局 AGENTS / `development-pipeline` / `code-reviewer`
- `feedback/2026-05-05-review-oversized-frontend-debt.md` -> 被 touched structural debt 规则覆盖
- `feedback/2026-05-06-closeout-must-evaluate-experience-extraction.md` -> `harness-router`
- `feedback/2026-05-08-dark-mode-controls-must-avoid-light-highlight.md` -> `frontend-engineer`
- `feedback/2026-05-08-project-tasks-default-to-harness.md` -> `harness-router` / 项目 `AGENTS.md`
- `feedback/2026-05-10-list-primary-title-column-should-stay-clean.md` -> `ctf-ui-theme-system`
- `feedback/2026-05-10-reuse-first-harness.md` -> 项目 `AGENTS.md` / `harness/policies/` / `harness/checks/` / `scripts/check-reuse-first.sh`

## 后续规则

- Agent 工作流、review、prompt、skill、policy 类反馈优先写入 `feedback/`。
- 可执行业务任务不要堆在 `feedback/`，应进入 `docs/plan/` 或 `docs/todo/`。
- 反馈已经沉淀到 `AGENTS.md`、skill、脚本或 CI 后，从活动反馈列表移除；需要追溯时使用 Git 历史。
