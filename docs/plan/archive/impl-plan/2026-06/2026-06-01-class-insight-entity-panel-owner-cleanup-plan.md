# 班级洞察实体 panel owner 收口计划

## Objective

- 新建 `entities/class-insight`，承接 `ClassInsightsPanel`、`ClassReviewPanel`、`ClassTrendPanel` 这组三个班级洞察稳定展示块。
- 让 `class-students-workspace` 与 `class-report-export` 都通过实体 public API 复用这些 panel，清掉 cross-feature internal import。

## Non-goals

- 不修改 `class-insight-window` 的时间窗口、query、路由同步或数据获取 owner。
- 不处理 `student-analysis-workspace`、`StudentInsightPanel.vue` 或 AWD / review archive 相关模块。
- 不调整三块 panel 的业务呈现逻辑、空状态文案或样式方向。

## Source Inputs

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/plan/impl-plan/2026-05-30-teacher-components-owner-migration-plan.md`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassTrendPanel.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- `code/frontend/src/entities/challenge/index.ts`
- `code/frontend/src/entities/training-timeline/index.ts`

## Brainstorming Conclusion

- 推荐方向：把三块 panel 一次性上提到 `entities/class-insight/ui/*`，并只暴露实体 public API。
- 不推荐方向：继续留在 `class-students-workspace` 再加 bridge，或复制一份给 `class-report-export`。前者边界仍然脏，后者会制造双 owner。

## Plan Review Result

- 这组三个文件回答的是“班级洞察对象如何稳定展示”，不回答“页面如何加载、切 tab、开导出弹窗”，符合 `entities/*` 的边界。
- `ClassStudentsPage.vue` 仍保留 workspace owner；`ClassReportExportPreviewSection.vue` 仍保留预览区组合 owner；实体层只提供稳定展示块，不吸入 route 或 dialog workflow。

## Task Slices

### Slice 1: 建立 `class-insight` 实体公共入口

- 目标：新增 `entities/class-insight/index.ts` 与 `ui/index.ts`，承接三块共享 panel。
- 变更面：
  - `code/frontend/src/entities/class-insight/index.ts`
  - `code/frontend/src/entities/class-insight/ui/index.ts`
  - `code/frontend/src/entities/class-insight/ui/ClassInsightsPanel.vue`
  - `code/frontend/src/entities/class-insight/ui/ClassReviewPanel.vue`
  - `code/frontend/src/entities/class-insight/ui/ClassTrendPanel.vue`
- 风险：
  - 如果把 query、window label 或 page actions 也拉进实体层，会污染 owner。

### Slice 2: 切换两个 consumer 到实体 public API

- 目标：让 `ClassStudentsPage.vue` 与 `ClassReportExportPreviewSection.vue` 只从 `@/entities/class-insight` 导入 panel。
- 变更面：
  - `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
  - `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
  - `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- 风险：
  - 如果 `class-students-workspace` 仍继续导出旧内部文件路径，会留下假 owner。

### Slice 3: 更新源码护栏与迁移台账

- 目标：把 raw-source 护栏切到实体路径，并记录 backlog 进展。
- 变更面：
  - `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
  - `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
  - `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- 风险：
  - 如果测试只跟着改路径，不校验 consumer 已转向实体 public API，后续仍可能回流深导入。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision class-insight-entity-panel-owner-cleanup`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/class-insight-entity-panel-owner-cleanup.md docs/plan/archive/impl-plan/2026-06/2026-06-01-class-insight-entity-panel-owner-cleanup-plan.md code/frontend/src/entities/class-insight/index.ts code/frontend/src/entities/class-insight/ui/index.ts code/frontend/src/entities/class-insight/ui/ClassInsightsPanel.vue code/frontend/src/entities/class-insight/ui/ClassReviewPanel.vue code/frontend/src/entities/class-insight/ui/ClassTrendPanel.vue code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue code/frontend/src/features/teaching/class-students-workspace/ui/index.ts code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `entities/class-insight` 是否只承接稳定展示，不吸入 page workflow owner。
- `ClassStudentsPage.vue` 与 `ClassReportExportPreviewSection.vue` 是否都已经走实体 public API，而不是继续深导入 feature internal UI。
- `class-students-workspace` 是否不再假装拥有这组三个 panel 的文件落点。
- raw-source 护栏是否真正锁住“共享 panel 走 entity owner”。

## Rollback / Recovery

- 如果实体命名不合适，可以回退命名或 public API 组织方式，但不能回退到 cross-feature internal import。
- 如果发现有第三个 consumer 依赖旧路径，补充到同一实体 public API 下，不再恢复 feature internal file owner。
