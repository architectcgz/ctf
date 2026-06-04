# Class Students Teacher Summary Convergence Plan

> 状态：Current
> 事实源：`teacher-surface.css`、`ClassStudentsOverviewPanel.vue`、teacher summary family raw-source 测试

## Objective

- 把 `ClassStudentsOverviewPanel` 顶部 summary 收口到现有 teacher shared surface owner。
- 清理 teacher summary family 仍散在单页本地的响应式和 header-attached summary contract。
- 保持 class-students、class-management、student-management、instance-management 当前视觉和交互不变。

## Non-goals

- 本轮不扩到 admin / platform summary family。
- 本轮不重构 dashboard summary family；它已在上一刀收口到 feature 内共享 CSS。
- 本轮不处理 review-archive summary 或 student-analysis KPI family。

## Source Inputs

- `code/frontend/src/assets/styles/teacher-surface.css`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`

## Architecture Fit Check

- teacher summary family 已有共享 owner，当前 debt 不在“缺共享入口”，而在 consumer 仍保留局部 contract 补丁。
- `ClassStudentsOverviewPanel` 的 summary 是 header 附着型 summary，不应继续只靠本地 `.class-overview-summary { padding: 0; }`。
- `StudentManagementPage` 的 `.teacher-summary-grid` 媒体查询说明共享 owner 还缺 teacher 页 summary 折叠 contract，这部分应回到 `teacher-surface.css`。

## Task Breakdown

### Slice 1: 扩 teacher shared summary contract

**Files**
- Modify: `code/frontend/src/assets/styles/teacher-surface.css`

- [ ] Step 1: 为 header 附着型 teacher summary 增加显式共享 class，例如 panel summary contract。
- [ ] Step 2: 把 teacher summary family 的移动端单列折叠收回共享 owner。

**Validation**
- raw-source 测试应检查 shared owner，不再接受本地 `.teacher-summary-grid`。

### Slice 2: 迁移 class-students / teacher summary consumer

**Files**
- Modify: `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue`
- Modify: `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- Modify: `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- Modify: `code/frontend/src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`

- [ ] Step 1: 让 `ClassStudentsOverviewPanel` 改为显式声明新的 shared panel summary class。
- [ ] Step 2: 删除 `ClassStudentsOverviewPanel` 本地 summary contract 补丁。
- [ ] Step 3: 删除 `StudentManagementPage` 本地 `.teacher-summary-grid` 响应式规则，改由 shared owner 承接。
- [ ] Step 4: 更新原始源码护栏测试。

**Validation**
- `cd code/frontend && pnpm exec vitest run src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- `cd code/frontend && pnpm typecheck`

## Review Focus

- 新 shared class 是否真实承接了 header 附着型 summary 的基础 contract，而不是再换个名字继续本地补。
- teacher summary family 的响应式折叠是否已经从单页本地规则回到 shared owner。
- `ClassStudentsOverviewPanel` 是否只保留页面语义，而不再持有基础 surface contract。

## Rollback / Recovery

- 如果 header 附着型 summary 与常规 `teacher-summary` 差异超出本轮范围，优先保留更窄的 shared class，不回退到页面本地补丁。
- 若更多 teacher 页面暴露类似局部 contract，本轮先限制在 touched surface，剩余项进入下一批。
