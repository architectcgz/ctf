# Reuse Decision

## Change type
component / layout

## Existing code searched
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/style.css
- code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewDirectorySection.vue
- code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewIndexWorkspace.vue
- code/frontend/src/components/teacher/class-management/ClassManagementPage.vue
- code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue

## Similar implementations found
- `workspace-shell.css` 已经把工作区顶部、左右留白和内容起始间距收口到 `space-workspace-*` 语义 token，不应该在 AWD 复盘外壳里继续写局部固定 padding。
- `TeacherAWDReviewDirectorySection.vue`、`ClassManagementPage.vue`、`TeacherInstanceManagementPage.vue` 已经在教师端目录区块上复用 `workspace-directory-list` / `workspace-directory-list--catalog` 作为统一外边框和 surface owner。
- `style.css` 已经提供 `workspace-directory-chip` 作为全局胶囊边框与交互基类，轮次切换按钮应复用这套 contract，而不是维持只在局部组件里弱化到接近纯文字的样式。

## Decision
refactor_existing

## Reason
这次不是新增一套 AWD 复盘页面样式，而是把详情页外壳和轮次切换重新接回现有 workspace token 与目录壳层。这样能直接继承全局设计好的顶部节奏、边框和 chip 交互语言，避免 AWD 详情页继续漂移出教师端已存在的工作区样式体系。

## Files to modify
- .harness/reuse-decisions/academy-awd-review-round-selector-shell-tokens.md
- code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue
- code/frontend/src/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue
- code/frontend/src/views/__tests__/spacingSystemTokens.test.ts
- code/frontend/src/views/teacher/__tests__/teacherAwdReviewRoundSelectorExtraction.test.ts
