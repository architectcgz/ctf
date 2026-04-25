# Superpowers 设计索引

## 定位

本目录保存使用 superpowers 产出的专题设计和实施计划。

- `specs/`：专题设计，描述目标、范围、数据和交互决策。
- `plans/`：实施计划，描述当时计划修改的文件和验证步骤。

后续查设计时优先读 `specs/`。`plans/` 只作为执行上下文，不作为最终设计事实源。

## 当前专题设计

- `specs/2026-04-01-attack-evidence-review-design.md`：攻防证据链与复盘。
- `specs/2026-04-01-contest-ops-enhancement-design.md`：赛事运营增强。
- `specs/2026-04-01-judge-modes-extension-design.md`：判题模式补强。
- `specs/2026-04-04-community-writeup-and-recommended-solutions-design.md`：社区题解与推荐题解。
- `specs/2026-04-04-teacher-dark-surface-alignment-design.md`：教师端暗色 surface 对齐。
- `specs/2026-04-04-web-source-audit-double-encode-design.md`：Web 源码审计题。
- `specs/2026-04-11-awd-challenge-config-design.md`：AWD 题目配置。
- `specs/2026-04-11-awd-checker-preview-design.md`：AWD checker 预览。
- `specs/2026-04-11-awd-checker-structured-editor-design.md`：AWD checker 结构化编辑。
- `specs/2026-04-11-awd-checker-validation-state-design.md`：AWD checker 校验状态。
- `specs/2026-04-12-awd-readiness-gate-design.md`：AWD 赛前检查 gate。
- `specs/2026-04-12-awd-phase9-review-archive-design.md`：AWD 复盘归档。
- `specs/2026-04-12-awd-phase10-student-battle-workspace-design.md`：AWD 学生作战工作台。
- `specs/2026-04-12-student-dashboard-action-panels-design.md`：学生仪表盘行动面板。
- `specs/2026-04-13-awd-phase11-assessment-integration-design.md`：AWD 接入能力画像。
- `specs/2026-04-13-awd-phase12-report-archive-alignment-design.md`：AWD 报告归档对齐。
- `specs/2026-04-13-awd-phase13-student-review-archive-structured-ui-design.md`：学生复盘归档结构化 UI。
- `specs/2026-04-14-contest-orchestration-workbench-design.md`：赛事编排工作台。
- `specs/2026-04-17-awd-final-design.md`：AWD 总设计与最终目标。
- `specs/2026-04-22-admin-contest-announcement-ops-design.md`：管理员赛事公告运维。
- `specs/2026-04-22-probe-easter-eggs-design.md`：探针彩蛋。

## 替代关系

- `2026-04-04-community-writeup-and-recommended-solutions-design.md` 替代旧的选手 writeup 教师评阅设计。旧文件已从仓库移除，避免继续把“教师评阅学生 writeup”当成当前方向。
- `2026-04-17-awd-final-design.md` 是 AWD 方向的总设计。早期 AWD phase 设计仍可用于追溯某个阶段的决策，但如果与总设计或当前架构文档冲突，以总设计和当前代码为准。

## 不在此目录的设计

- 拓扑编排器和环境模板的页面级设计位于 `design-system/ctf-platform/pages/topology-editor.md` 与 `design-system/ctf-platform/pages/env-templates.md`。
- 通用 UI 风格、色彩、字体和页面稿索引位于 `design-system/ctf-platform/`。
- 后端专题架构位于 `docs/architecture/backend/design/`。
