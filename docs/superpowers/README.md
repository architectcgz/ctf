# Superpowers 过程文档索引

## 定位

本目录保存使用 superpowers 产出的实施计划和过程索引。

- `plans/`：实施计划，描述当时计划修改的文件和验证步骤。

当前需要判断“最终设计是什么”时，先读 `docs/architecture/README.md`、`docs/architecture/frontend/` 和 `docs/architecture/features/`。本目录只用于追溯设计演进和实施上下文。

## 当前作用

- `plans/`：保留实施计划，供追溯“当时准备怎么实现、怎么验证”。
- `specs/README.md`：保留迁移说明，明确专题最终设计已移入 `docs/architecture/features/`。
- `docs/architecture/features/`：承接已经采用的专题最终设计。

## 与 architecture 的关系

- 已实现的 `plans/` 不再作为最终设计事实源；对应最终文档统一写入 `docs/architecture/features/`。
- `community-writeup-and-recommended-solutions-design.md` 替代旧的选手 writeup 教师评阅设计。旧文件已从仓库移除，避免继续把“教师评阅学生 writeup”当成当前方向。
- `awd-final-design.md` 对应的最终专题事实源位于 `docs/architecture/features/awd-final-design.md`。更早的 AWD phase 草稿只用于追溯阶段决策。
- 拓扑编排器、环境模板、全局设计系统都已迁到 `docs/architecture/frontend/`，不再从旧的 `design-system/ctf-platform/` 读取。

## 已实现计划的最终文档映射

- `plans/2026-03-22-remove-teacher-compat.md` → `docs/architecture/features/teaching-readmodel-boundary-design.md`
- `plans/2026-04-01-attack-evidence-review-implementation.md` → `docs/architecture/features/attack-evidence-review-design.md`
- `plans/2026-04-04-community-writeup-and-recommended-solutions-implementation.md` → `docs/architecture/features/community-writeup-and-recommended-solutions-design.md`
- `plans/2026-04-11-awd-engine-phase2-http-standard.md` → `docs/architecture/features/awd-http-standard-checker-design.md`
- `plans/2026-04-11-awd-engine-phase3-inspector.md` → `docs/architecture/features/awd-inspector-checker-result-design.md`
- `plans/2026-04-11-awd-engine-phase4-challenge-config.md` → `docs/architecture/features/awd-challenge-config-design.md`
- `plans/2026-04-11-awd-engine-phase5-structured-config-editor.md` → `docs/architecture/features/awd-checker-structured-editor-design.md`
- `plans/2026-04-11-awd-engine-phase6-checker-preview.md` → `docs/architecture/features/awd-checker-preview-design.md`
- `plans/2026-04-11-awd-engine-phase7-checker-validation-state.md` → `docs/architecture/features/awd-checker-validation-state-design.md`
- `plans/2026-04-12-awd-engine-phase8-readiness-gate.md` → `docs/architecture/features/awd-readiness-gate-design.md`
- `plans/2026-04-12-awd-phase9-review-archive-implementation.md` → `docs/architecture/features/awd-review-archive-design.md`
- `plans/2026-04-12-awd-phase10-student-battle-workspace-implementation.md` → `docs/architecture/features/awd-student-battle-workspace-design.md`
- `plans/2026-04-12-student-dashboard-action-panels.md` → `docs/architecture/features/student-dashboard-action-panels-design.md`
- `plans/2026-04-13-awd-phase11-assessment-integration-implementation.md` → `docs/architecture/features/awd-assessment-integration-design.md`
- `plans/2026-04-13-awd-phase12-report-archive-alignment-implementation.md` → `docs/architecture/features/awd-report-archive-alignment-design.md`
- `plans/2026-04-13-awd-phase13-student-review-archive-structured-ui-implementation.md` → `docs/architecture/features/awd-student-review-archive-structured-ui-design.md`
- `plans/2026-04-15-contest-orchestration-workbench-implementation.md` → `docs/architecture/features/contest-orchestration-workbench-design.md`
- `plans/2026-04-17-awd-runtime-cutover-phase2.md` → `docs/architecture/features/awd-runtime-service-read-path-design.md`
- `plans/2026-04-18-awd-runtime-target-identity-phase3.md` → `docs/architecture/features/awd-runtime-service-identity-design.md`
- `plans/2026-04-21-package-topology-sync.md` → `docs/architecture/features/package-topology-sync-design.md`
- `plans/2026-04-22-admin-contest-announcement-ops.md` → `docs/architecture/features/admin-contest-announcement-ops-design.md`
- `plans/2026-04-22-probe-easter-eggs-implementation.md` → `docs/architecture/features/probe-easter-eggs-design.md`
