> 状态：Current
> 事实源：`StudentAnalysisPage.vue` 当前模板、已完成的 hero / section 收口、现有教师端源码护栏
> 替代：无

# Student Analysis Page Dead Style Cleanup Implementation Plan

## 目标

- 清理 `StudentAnalysisPage.vue` 中已经没有模板挂载点的 scoped CSS。
- 保持页面当前的 tab / query / `StudentInsightPanel` 装配 owner 和用户可见行为不变。
- 让源码护栏继续只覆盖当前仍然存在的样式 owner，而不是为旧区块保留无效断言空间。

## 非目标

- 本轮不删除 `StudentAnalysisPage.vue` 的未使用 props。
- 本轮不改 `TeacherStudentAnalysis.vue`、`PlatformStudentAnalysis.vue` 或 page model。
- 本轮不调整任何运行时交互、路由或文案。

## 输入依据

- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## 当前结论

- 页面模板当前只保留：tabs、error alert、overview hero 装配、`StudentInsightPanel` 装配。
- `context-rail`、`class-switch*`、`student-directory*`、`quick-action__*`、`context-block*` 等样式在模板中已经没有对应挂载点。
- 这些残留样式继续留在组件里，只会扩大文件体积并混淆后续 review 的真实 owner 面。

## 任务切片

### Slice 1：删除已失效的 scoped CSS

- 目标：
  - 从 `StudentAnalysisPage.vue` 删除模板已不再使用的样式块。
  - 保留当前仍被模板引用的 `workspace-shell`、`content-pane`、`workspace-alert`、`quick-action`、`:deep(.section-card)` 和移动端 tabs 间距等样式。
- 预期改动：
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts src/views/teacher/__tests__/teacherSurface.test.ts`
- Review focus：
  - 是否只删除真正没有模板挂载点的私有样式。
  - 是否误删 `workspace-alert` / `quick-action` / `section-card` 这些当前仍在用的样式。

### Slice 2：必要时同步源码护栏

- 目标：
  - 如果 `teacherDetailSurfaceAlignment.test.ts` 还在隐式依赖旧样式块，更新为只检查当前仍然存在的 surface owner。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- 验证：
  - 同上
- Review focus：
  - 护栏是否继续锚定当前真实存在的样式 owner，而不是为旧区块残留护栏噪音。

## 风险

- 如果误把 `quick-action` 或 `workspace-alert` 相关样式删掉，会直接影响错误态按钮的展示。
- 仅靠字符串匹配判断“未使用样式”有风险，所以本轮只删除模板中完全找不到类名挂载点的样式块，不碰仍有任何模板引用的类。

## 回退方式

- 如清理后出现样式回归，可直接从本组件历史里恢复被删的具体样式块。
- 因本轮不改行为 owner 或 API，回退只涉及前端单文件样式与测试。
