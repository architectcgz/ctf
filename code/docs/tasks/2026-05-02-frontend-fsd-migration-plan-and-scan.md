# 前端分层迁移计划与扫描基线（2026-05-02）

## 背景
- 目标：按 `frontend-sliced-architecture` 约束，把 `views` 路由页收敛为“组合层”，把流程、副作用与数据访问下沉到 `features/*/model/useXxxPage.ts`。
- 范围：`code/frontend/src/views` 与对应 `features`。

## 迁移计划（执行顺序）
1. 路由壳层收口  
   - 清理 `views` 里 `useRoute/useRouter/router.push/router.replace/useRouteQueryTabs`。  
   - 跳转、query tab 编排、route params 解析迁移到 `features/*/model/useXxxRoutePage.ts`。
2. 数据访问与异步流程收口  
   - 清理 `views` 里对 `@/api/*`（除类型导入）的直接访问。  
   - 请求、错误处理、提交流程下沉到 `useXxxPage.ts`。
3. 生命周期与业务编排收口  
   - `views` 中与业务相关的 `onMounted/watch` 下沉；保留纯展示必要状态（如局部分页）可暂留。
4. 边界测试补齐  
   - 每批增加 source 边界断言：  
   - 包含 `useXxxPage` / `useXxxRoutePage`  
   - 不包含 `useRoute/useRouter`（路由壳页）  
   - 不包含 `from '@/api/(非 contracts)'`（视图层）
5. 收尾与一致性  
   - 统一 `index.ts` 导出与命名（`useXxxPage`/`useXxxRoutePage`）。  
   - 最终跑相关 vitest + `npm run typecheck`。

## 扫描结果（本次基线）

### 1) 路由壳层直连扫描
命令：
```bash
rg -n "useRoute|useRouter|router\.push|router\.replace|useRouteQueryTabs" code/frontend/src/views/**/*.vue
```
结果：无命中（已清零）。

### 2) 视图层 API 直连扫描（排除 contracts）
命令：
```bash
rg -nP "from ['\"]@/api/(?!contracts)" code/frontend/src/views/**/*.vue
```
结果：无命中（已清零）。

补充：若使用宽匹配
```bash
rg -n "from ['\"]@/api/" code/frontend/src/views/**/*.vue
```
当前仅剩类型导入：
- `src/views/platform/ContestManage.vue` -> `@/api/contracts`
- `src/views/scoreboard/ScoreboardView.vue` -> `@/api/contracts`

### 3) 视图层生命周期扫描（用于下一轮收口）
命令：
```bash
rg -n "onMounted\(|watch\(|useEffect\(" code/frontend/src/views --glob '*.{vue,tsx,jsx}'
```
结果：无命中（已清零）。

## 当前状态

### 已完成（P0）
- `src/views/platform/ContestManage.vue`
  - 已新增 `features/platform-contests/model/useContestManagePage.ts`。
  - `onMounted` 刷新、创建后面板切换、公告抽屉状态、筛选 setter 和 dialog open change 已下沉。
- `src/views/platform/UserManage.vue`
  - 已新增 `features/platform-users/model/usePlatformUserManagePage.ts`。
  - `onMounted` 刷新、删除确认、筛选 setter 和 dialog open change 已下沉。
- `src/views/platform/AWDChallengeLibrary.vue`
  - 已扩展 `useAwdChallengeLibraryPage` 为页面组合 owner。
  - `onMounted` 刷新、筛选 setter、dialog open change 已下沉。
- `src/views/platform/AWDChallengeImport.vue`
  - 已新增 `useAwdChallengeImportPage`。
  - `onMounted` 导入队列刷新已下沉。

### Batch D：低优先级展示收口（已完成）
- `ScoreboardView.vue`
  - 已新增 `features/scoreboard/model/useScoreboardContestDirectoryPage.ts`。
  - 本地分页 `watch` 与展示 helper 已下沉。
- `InstanceList.vue`
  - 已新增 `features/instance-list/model/useInstanceWarningFocus.ts`。
  - warning 弹窗焦点管理 `watch` 已下沉。
- `LoginView.vue`
  - 控制台提示已改为 setup 阶段函数调用，不再使用 `onMounted`。

## 下一批迁移方案（从 route view 清理转向 slice 内部治理）

### 新扫描基线（2026-05-02）

#### 1) feature 深层导入扫描
命令：
```bash
rg -n "from ['\"]@/features/.+/model/" code/frontend/src --glob '*.{ts,vue}'
```
结果：运行时代码无命中；当前命中均为 `?raw` source 测试文件，可接受。后续约束是：业务代码从其他 feature 读取能力时必须走对应 slice 的 `index.ts`，不允许直接导入 `features/*/model/*`。

#### 2) feature 反向依赖旧 components 扫描
命令：
```bash
rg -n "from ['\"]@/components/" code/frontend/src/features code/frontend/src/entities --glob '*.{ts,vue}'
```
结果：需要迁移的代表性命中：
- `features/student-dashboard/model/useStudentDashboardPage.ts` 直接导入 `components/dashboard/student/*Page.vue`。
- `features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts` 直接导入 `components/platform/topology/*` 类型与 helper。
- `features/contest-projector/model/useContestProjectorPage.ts` 与 `useContestProjectorDerived.ts` 直接导入 `components/platform/contest/projector/*` 类型与格式化 helper。
- `features/contest-awd-config/model/useContestAwdConfigPage.ts` 直接导入 `components/platform/contest/awdCheckerConfigSupport`。
- `features/platform-challenges/model/useChallengeManagePage.ts`、`features/audit-log/model/useAuditLogPage.ts`、`features/image-management/model/useImageManagePage.ts` 直接依赖 `WorkspaceDirectoryToolbar.vue` 的排序类型。

判断：这是下一批优先级最高的架构问题。`features/model` 不应该依赖旧 `components` 目录里的具体 UI 文件；否则 model 与展示组件耦合，后续很难把 `components` 按 `widgets/entities/shared/ui` 迁走。

#### 3) 组件层直接 API 扫描（排除 contracts）
命令：
```bash
rg -nP "from ['\"]@/api/(?!contracts)" code/frontend/src/components code/frontend/src/widgets code/frontend/src/entities --glob '*.{ts,vue}'
```
结果：仍需收口：
- `components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - 直接调用 `@/api/admin/contests`、`@/api/admin/authoring`，并处理 `ApiError`。
- `components/platform/writeup/ChallengeWriteupManagePanel.vue`
  - 直接调用 `@/api/admin/authoring`、`@/api/teacher`。
- `components/teacher/TeacherInterventionPanel.vue`
  - 直接调用 `getStudentRecommendations`。
- `components/platform/contest/AWDChallengeConfigDialog.vue`
  - 直接调用 `runContestAWDCheckerPreview`。
- `components/platform/topology/topologyDraft.ts`
  - 直接依赖 `@/api/admin/authoring` 的 payload 类型。
- `components/platform/images/ImageCreateModal.vue`
  - 类型依赖 `@/api/admin/authoring`，低风险，可随 image slice 一并处理。

判断：组件层直接 API 是比“组件里有 watch”更高优先级的问题。需要把 API 调用、错误策略、数据归一化移动到 feature model；组件只通过 props/emits 或注入的 action 回调触发。

#### 4) 大型 feature model 扫描
命令：
```bash
find code/frontend/src/features -path '*/model/*.ts' -type f -print0 | xargs -0 wc -l | sort -nr | head -20
```
当前最高风险文件：
- `features/challenge-topology-studio/model/useChallengeTopologyStudioPage.ts`：999 行。
- `features/contest-awd-admin/model/usePlatformContestAwd.ts`：966 行。
- `features/contest-awd-config/model/useContestAwdConfigPage.ts`：592 行。
- `features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`：498 行。
- `features/contest-detail/model/useContestDetailPage.ts`：438 行。
- `features/platform-contests/model/usePlatformContests.ts`：404 行。
- `features/instance-list/model/useInstanceListPage.ts`：370 行。

判断：这些 composable 已经承担了页面装配、请求、表单 draft、派生展示、错误策略和动作编排。下一阶段不追求一次拆完，而是按“可测试的业务块”拆出小 composable 或纯函数，避免 `useXxxPage.ts` 继续膨胀成新的胖路由页。

#### 5) route view 残留展示逻辑扫描
命令：
```bash
rg -n "\b(computed|function|const [A-Za-z0-9_]+ = \(|async function)\b" code/frontend/src/views --glob '*.vue'
```
需要继续评估的命中：
- `views/auth/RegisterView.vue`：仍有 `onSubmit`，注册页尚未迁移为 `features/auth/model/useRegisterPage.ts`。
- `views/auth/LoginView.vue`：仍有 `showProbeMessage`、`emitLoginConsoleHints`、`handleHeroProbe`、`onSubmit`，可继续收口到 `useLoginPage.ts` 或拆出登录页 presentation helpers。
- `views/platform/AWDReviewIndex.vue`：仍有 `computed` 和 `resetFilters`，可迁移到 `features/teacher-awd-review` 或独立 platform review slice。
- `views/instances/InstanceList.vue`：仍有实例访问格式化 helper，低风险，可迁移到 `features/instance-list/model` 或 `entities/instance`。
- `views/teacher/TeacherStudentAnalysis.vue`：仍有 `openClassReportDialog`，低风险，可并入 `useTeacherStudentAnalysisPage`。

判断：这些不是 P0，因为 route view 的 API/路由/生命周期边界已经清零；但它们会影响“route view 只组合 UI”的最终完成度。

### Batch E：组件层 API 收口（优先执行）
目标：把仍在 `components` / `widgets` / `entities` 中直接调用 API 的业务流程迁移到 feature model。

#### E1：竞赛题目编排面板
- 迁移对象：`components/platform/contest/ContestChallengeOrchestrationPanel.vue`。
- 建议落点：`features/contest-workbench/model/useContestChallengeOrchestration.ts`，必要时拆纯函数到 `features/contest-workbench/model/challengeOrchestration.ts`。
- 改法：
  - API 调用 `getContestChallenges`、`updateContestChallenges`、`getChallenges` 等移动到 feature model。
  - 组件改为接收 `state`、`availableChallenges`、`selectedChallenges`、`loading`、`error` 和 action props/emits。
  - 错误文案与 `ApiError` 处理由 feature model 统一返回。
- 验证：
  - 迁移现有 `ContestChallengeOrchestrationPanel.test.ts` 的业务断言到 composable 测试。
  - 为组件补 source 边界断言：不包含 `from '@/api/admin/contests'`、`from '@/api/admin/authoring'`、`ApiError`。

#### E2：题解管理面板
- 迁移对象：`components/platform/writeup/ChallengeWriteupManagePanel.vue`。
- 建议落点：`features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`。
- 改法：
  - `getChallengeWriteup`、`deleteChallengeWriteup`、`getTeacherWriteupSubmissions` 移到 feature。
  - 面板保留表格、弹窗和按钮事件，不直接决定请求生命周期。
  - 如果面板需要同时服务管理员和教师视角，feature model 通过参数区分 mode，不让组件直接拼 API。
- 验证：
  - 补 composable 测试覆盖加载、删除、教师提交列表、错误分支。
  - source 测试保证组件不再导入 `@/api/admin/authoring` 与 `@/api/teacher`。

#### E3：教师干预推荐
- 迁移对象：`components/teacher/TeacherInterventionPanel.vue`。
- 建议落点：`features/teacher-student-analysis/model/useTeacherInterventionRecommendations.ts`。
- 改法：
  - `getStudentRecommendations` 下沉到 feature。
  - 组件只接收推荐列表、loading、错误与刷新事件。
- 验证：
  - 直接测 composable 的 student id 变化、重复请求防护和错误态。

#### E4：AWD checker 预览
- 迁移对象：`components/platform/contest/AWDChallengeConfigDialog.vue`。
- 建议落点：`features/contest-awd-config/model/useAwdCheckerPreview.ts`。
- 改法：
  - `runContestAWDCheckerPreview` 从 dialog 移到 feature。
  - dialog 内保留表单 draft、展示和触发事件；预览 action、loading、错误归 feature。
  - 由于该组件 2525 行，先只迁移 API action，不在同一批重写 UI。
- 验证：
  - source 测试保证 dialog 不直接导入 `@/api/admin/contests`。
  - composable 测试覆盖 payload 映射、成功预览、`ApiError`。

### Batch F：feature 对旧 components 的反向依赖治理
目标：解除 `features/model` 对 `components/*` 的直接依赖，使 model 可独立测试，也为后续 `widgets/entities/shared` 目录迁移留空间。

#### F1：通用目录排序类型下沉
- 迁移对象：
  - `WorkspaceDirectorySortOption` 目前从 `components/common/WorkspaceDirectoryToolbar.vue` 被多个 feature model 导入。
- 建议落点：`entities/workspace-directory/model/sort.ts` 或 `features/*/model/sortOptions.ts` 中的本地类型；若多个 slice 共用，优先建 `entities/workspace-directory`。
- 改法：
  - 从 `.vue` 中抽出纯类型，组件和 feature 都从新位置导入。
  - 避免 feature model 依赖具体 toolbar 组件。
- 验证：
  - `rg -n "WorkspaceDirectorySortOption.*@/components/common/WorkspaceDirectoryToolbar" code/frontend/src` 无命中。

#### F2：student dashboard 页面组件注册迁移
- 迁移对象：`features/student-dashboard/model/useStudentDashboardPage.ts` 直接导入 `components/dashboard/student/*Page.vue`。
- 建议落点：把 tab 配置中的组件引用上移到 `views/dashboard/DashboardView.vue` 或新建 `widgets/student-dashboard`；feature model 只返回 tab key、label、badge、可见性。
- 改法：
  - model 不再 import `.vue` 页面组件。
  - route/widget 层负责把 key 映射到具体组件。
- 验证：
  - source 测试保证 `useStudentDashboardPage.ts` 不含 `@/components/dashboard/student`。

#### F3：topology / projector / AWD config helper 迁移
- 迁移对象：
  - `components/platform/topology/topologyLayout`、`topologyDraft`。
  - `components/platform/contest/projector/contestProjectorTypes`、`contestProjectorFormatters`。
  - `components/platform/contest/awdCheckerConfigSupport`。
- 建议落点：
  - `features/challenge-topology-studio/model/topologyLayout.ts`、`topologyDraft.ts`。
  - `features/contest-projector/model/projectorTypes.ts`、`projectorFormatters.ts`。
  - `features/contest-awd-config/model/awdCheckerConfigSupport.ts`。
- 改法：
  - 纯 helper/type 移到对应 feature model。
  - UI 组件改为从 feature public API 或本 slice model 导入。
  - 对仍然跨多个 feature 使用的类型，再评估是否放入 entity，而不是放入旧 `components`。
- 验证：
  - `rg -n "from ['\"]@/components/platform/(topology|contest/projector|contest/awdCheckerConfigSupport)" code/frontend/src/features` 无命中。

### Batch G：大型 feature model 分解
目标：防止 `useXxxPage.ts` 变成新的“胖页面”。每次只拆一个业务块，保证测试可读。

#### G1：`challenge-topology-studio`
- 当前：`useChallengeTopologyStudioPage.ts` 999 行。
- 拆分建议：
  - `useTopologyDraftState.ts`：草稿、节点/边编辑、dirty 判断。
  - `useTopologyTemplateLoader.ts`：模板列表加载、选择、错误态。
  - `topologyValidation.ts`：纯校验函数。
  - `topologyPayload.ts`：DTO <-> draft 映射。
- 保留：`useChallengeTopologyStudioPage.ts` 只组装这些 composable 并暴露 route/widget 需要的状态。
- 验证：纯函数单测 + 页面 composable 回归测试 + typecheck。

#### G2：`contest-awd-admin`
- 当前：`usePlatformContestAwd.ts` 966 行。
- 拆分建议：
  - `useAwdRoundManagement.ts`：轮次创建、选择、刷新。
  - `useAwdServiceOperations.ts`：服务检查、实例操作、批量动作。
  - `useAwdReadinessDecision.ts`：readiness 复核、覆盖、状态派生。
  - `awdAdminPresentation.ts`：纯展示计数、状态标签、排序。
- 验证：保留现有 `usePlatformContestAwd.test.ts`，新增拆出模块的直接测试，避免只测总 composable。

#### G3：`contest-awd-config`
- 当前：`useContestAwdConfigPage.ts` 592 行，且依赖 components helper。
- 拆分建议：
  - 先执行 F3/E4，解除 helper/API 与 dialog 耦合。
  - 再拆 `useAwdCheckerConfigDraft.ts`、`useAwdChallengeSelection.ts`、`useAwdCheckerPreview.ts`。
- 验证：重点覆盖 payload 映射、preview action、保存后的刷新。

### Batch H：route view 最后一层展示逻辑收口
目标：不是为了清扫描，而是让 route view 接近“组合层”。

优先顺序：
1. `RegisterView.vue`：新增 `features/auth/model/useRegisterPage.ts`，迁移 `onSubmit`、loading、错误、成功跳转策略。
2. `LoginView.vue`：把 probe/console helper 与 `onSubmit` 收到 `useLoginPage.ts`，route view 保留 hero 交互绑定。
3. `AWDReviewIndex.vue`：把筛选计算、统计计数、`resetFilters` 下沉到 `features/teacher-awd-review`。
4. `TeacherStudentAnalysis.vue`：把 `openClassReportDialog` 合并进 `useTeacherStudentAnalysisPage`。
5. `InstanceList.vue`：把实例访问展示 helper 下沉到 `features/instance-list` 或 `entities/instance`。

验证：
- 更新对应 `views/**/__tests__` 的 source 边界断言。
- 保留用户可见行为不变，优先做组件浅层测试和 composable 直接测试。

### Batch I：边界自动化与文档收尾
目标：把本次迁移沉淀成可持续约束，避免后续新增页面重新退回旧模式。

改法：
- 新增或扩展 source boundary 测试：
  - `views` 不允许导入非 contracts API。
  - `views` 不允许直接使用 `useRoute/useRouter/router.push/router.replace/useRouteQueryTabs`。
  - 业务代码不允许跨 slice 深导入 `features/*/model/*`，测试文件 `?raw` 例外。
  - `features/model` 不允许导入 `@/components/*`，少量迁移期间例外需要在测试中白名单并写明截止批次。
- 将本文件的“下一批扫描命令”同步到 `frontend-sliced-architecture` skill 的 reference，作为之后新项目或新页面评审清单。

## 下一批方案执行进展

### 已完成：Batch E1（竞赛题目编排面板 API 收口）
- 新增 `features/contest-workbench/model/useContestChallengeOrchestration.ts`，承接原 `ContestChallengeOrchestrationPanel` 内部的数据加载、保存、移除、弹层触发、生命周期与错误处理。
- `ContestChallengeOrchestrationPanel.vue` 已改为路由/组件组合层，不再直接依赖：
  - `@/api/admin/contests`
  - `@/api/admin/authoring`
  - `@/api/request`
- 补充 source 边界断言：
  - `components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts` 新增对 `useContestChallengeOrchestration` 的引用断言与 API import 禁止断言。

验证：
```bash
npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
npm run typecheck
```

### 已完成：Batch E2（题解管理面板 API 收口）
- 新增 `features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`，承接 `ChallengeWriteupManagePanel` 的题解目录加载、学员题解分页、删除流程、生命周期触发和派生目录行数据。
- `ChallengeWriteupManagePanel.vue` 已收敛为展示 + 导航交互层，不再直接依赖：
  - `@/api/admin/authoring`
  - `@/api/teacher`
- 补充 source 边界断言：
  - `views/platform/__tests__/ChallengeWriteupManagePanel.test.ts` 新增 `useChallengeWriteupManagement` 引用断言与 API import 禁止断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts
npm run typecheck
```

### 已完成：Batch E3（教师介入面板推荐 API 收口）
- 新增 `features/teacher-student-analysis/model/useTeacherInterventionRecommendations.ts`，承接 `TeacherInterventionPanel` 的介入候选计算、推荐目标选择和推荐题异步拉取 watch。
- `TeacherInterventionPanel.vue` 已收敛为展示 + 学员跳转层，不再直接依赖：
  - `@/api/teacher`（保留 `@/api/contracts` 类型导入）
- 补充 source 边界断言：
  - `views/teacher/__tests__/teacherInterventionPanelLayout.test.ts` 新增 `useTeacherInterventionRecommendations` 引用断言与 API import 禁止断言。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/teacherInterventionPanelLayout.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts
npm run typecheck
```

### 已完成：Batch E4（AWD Checker 试跑 API 收口）
- 新增 `features/contest-awd-config/model/useAwdCheckerPreview.ts`，承接 `runContestAWDCheckerPreview` 请求组装与执行。
- `AWDChallengeConfigDialog.vue` 已改为调用 `runAwdCheckerPreview`，不再直接依赖：
  - `@/api/admin/contests`（保留 `@/api/contracts` 类型导入）
- 补充 source 边界断言：
  - `views/__tests__/duplicateActionGuardAudit.test.ts` 新增 `AWDChallengeConfigDialog` 对 `features/contest-awd-config` 的引用断言与 `@/api/admin/contests` 禁止断言。

验证：
```bash
npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts
npm run typecheck
```

### 已完成：Batch F1（通用目录排序类型下沉）
- 新增实体层类型出口：
  - `entities/workspace-directory/model/sort.ts`
  - `entities/workspace-directory/model/index.ts`
  - `entities/workspace-directory/index.ts`
- `WorkspaceDirectorySortOption` 已从 `WorkspaceDirectoryToolbar.vue` 抽离，改由实体层提供。
- 已替换依赖点：
  - `features/audit-log/model/useAuditLogPage.ts`
  - `features/image-management/model/useImageManagePage.ts`
  - `features/platform-challenges/model/useChallengeManagePage.ts`
  - `components/platform/audit/AuditLogDirectoryPanel.vue`
  - `components/platform/images/ImageDirectoryPanel.vue`
  - `components/platform/challenge/ChallengeManageDirectoryPanel.vue`
- 结果：`features/*` 不再从 `@/components/common/WorkspaceDirectoryToolbar.vue` 导入排序类型。

验证：
```bash
rg -n "WorkspaceDirectorySortOption.*@/components/common/WorkspaceDirectoryToolbar" code/frontend/src --glob '*.{ts,vue}'
rg -n "from '@/components/common/WorkspaceDirectoryToolbar.vue'" code/frontend/src/features --glob '*.ts'
npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ImageManage.test.ts src/views/platform/__tests__/AuditLog.test.ts
npm run typecheck
```

### 已完成：Batch F2（student dashboard 组件映射上移）
- `features/student-dashboard/model/useStudentDashboardPage.ts` 不再导入：
  - `StudentOverviewPage.vue`
  - `StudentRecommendationPage.vue`
  - `StudentCategoryProgressPage.vue`
  - `StudentTimelinePage.vue`
  - `StudentDifficultyPage.vue`
- 组件映射已上移到 `views/dashboard/DashboardView.vue`，feature model 仅保留：
  - tab 元数据
  - panel 绑定数据 `resolveDashboardPanelBindings`
  - 页面流程与导航行为
- 补充边界断言：
  - `views/dashboard/__tests__/DashboardView.test.ts` 新增 `useStudentDashboardPage.ts` 不允许导入上述 `.vue` 面板组件的断言。

验证：
```bash
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/dashboard/__tests__/dashboardPanelExtraction.test.ts
npm run typecheck
```

### 已完成：Batch F3 子项（contest-projector helper/type 迁移）
- 新增 feature 内部文件：
  - `features/contest-projector/model/projectorTypes.ts`
  - `features/contest-projector/model/projectorFormatters.ts`
- `features/contest-projector/model/useContestProjectorPage.ts` 与 `useContestProjectorDerived.ts` 已改为引用本 feature 内部 `projectorTypes/projectorFormatters`，不再反向依赖 `components/platform/contest/projector/*`。
- 补充边界断言：
  - 新增 `features/contest-projector/model/useContestProjectorBoundary.test.ts`，约束 projector feature model 不允许导入 `components/projector` 内部文件。

验证：
```bash
rg -n "from ['\"]@/components/platform/contest/projector" code/frontend/src/features --glob '*.ts'
npm run test:run -- src/features/contest-projector/model/useContestProjectorBoundary.test.ts src/features/contest-projector/model/useContestProjectorDerived.test.ts src/features/contest-projector/model/useContestProjectorData.test.ts
npm run typecheck
```

### 已完成：Batch F3 子项（awd checker config support 下沉）
- 新增 `features/contest-awd-config/model/awdCheckerConfigSupport.ts`，承接原 `components/platform/contest/awdCheckerConfigSupport.ts` 的 checker 配置草稿与构造逻辑。
- `useContestAwdConfigPage.ts` 已改为引用 feature 内部 `awdCheckerConfigSupport`，不再反向依赖组件层 helper。
- `AWDChallengeConfigDialog.vue` 已改为从 `features/contest-awd-config/model/*` 引用：
  - `runAwdCheckerPreview`
  - `awdCheckerConfigSupport`
- 更新 source 边界断言：
  - `views/platform/__tests__/ContestAwdConfig.test.ts` 断言 `useContestAwdConfigPage` 不再导入 `@/components/platform/contest/awdCheckerConfigSupport`。
  - `views/__tests__/duplicateActionGuardAudit.test.ts` 断言 `AWDChallengeConfigDialog` 走 feature model 入口。

验证：
```bash
rg -n "from ['\"]@/components/platform/contest/awdCheckerConfigSupport" code/frontend/src/features --glob '*.ts'
npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts
npm run typecheck
```

### 已完成：Batch F3 子项（topology helper/type 迁移）
- 新增 feature 内部模块：
  - `features/challenge-topology-studio/model/topologyLayout.ts`
  - `features/challenge-topology-studio/model/topologyDraft.ts`
  - `features/challenge-topology-studio/model/topologyTypes.ts`
- `useChallengeTopologyStudioPage.ts` 已改为使用 feature 内部模块，不再反向依赖：
  - `@/components/platform/topology/topologyLayout`
  - `@/components/platform/topology/topologyDraft`
  - `@/components/platform/topology/TopologyCanvasBoard.vue`（类型）
- 补充边界断言：
  - 新增 `features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts`

验证：
```bash
rg -n "from ['\"]@/components/platform/topology" code/frontend/src/features/challenge-topology-studio --glob '*.ts'
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch F3 子项（student-dashboard 类型解耦）
- `useStudentDashboardPage.ts` 已去除 `@/components/dashboard/student/types` 导入，改为 feature 内部本地类型声明。
- 结果：`features/*` 运行时代码中的 `@/components/*` 反向依赖已清零（剩余命中仅为边界测试源码字符串断言）。

验证：
```bash
rg -n "from ['\"]@/components/" code/frontend/src/features --glob '*.{ts,vue}'
npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts
npm run typecheck
```

### 已完成：Batch H1/H2（auth 路由页流程收口）
- 新增 feature 流程模型：
  - `features/auth/model/useRegisterPage.ts`
  - `features/auth/model/useLoginPage.ts`
- `RegisterView.vue` 与 `LoginView.vue` 已改为组合层：
  - 表单状态、提交流程、错误态、重复提交保护均由 feature composable 承担。
  - `LoginView` 的 probe 提示与控制台提示逻辑已下沉至 `useLoginPage`。
  - 浏览器自动填充场景仍通过 fallback 值保留原有提交行为。
- 补充测试：
  - `features/auth/model/useRegisterPage.test.ts`
  - `features/auth/model/useLoginPage.test.ts`
- 更新路由页边界断言：
  - `views/auth/__tests__/LoginView.test.ts`
  - `views/auth/__tests__/RegisterView.test.ts`

验证：
```bash
npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/features/auth/model/useRegisterPage.test.ts src/features/auth/model/useLoginViewPage.test.ts src/views/auth/__tests__/LoginView.test.ts src/views/auth/__tests__/RegisterView.test.ts
npm run typecheck
```

### 已完成：Batch H3（AWDReviewIndex 展示逻辑收口）
- `features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts` 新增并统一返回：
  - `hasActiveFilters`
  - `reviewRows`
  - `resetFilters`
- `AWDReviewIndex.vue` 已移除本地筛选/统计/映射计算，改为直接消费 feature 输出。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/AWDReviewIndex.test.ts
npm run typecheck
```

### 已完成：Batch H4（TeacherStudentAnalysis 弹窗状态收口）
- `useTeacherStudentAnalysisPage.ts` 新增：
  - `reportDialogVisible`
  - `openClassReportDialog`
- `TeacherStudentAnalysis.vue` 已去除本地 `ref` 与 `openClassReportDialog`，完全由 feature 输出驱动报告导出弹窗开关。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
npm run typecheck
```

### 已完成：Batch H5（InstanceList 访问展示 helper 下沉）
- `features/instance-list/model/useInstanceListPage.ts` 新增并导出：
  - `formatInstanceAccessDisplay`
  - `canOpenInstanceInBrowser`
- `views/instances/InstanceList.vue` 已移除本地访问展示 helper，改为复用 feature 层实现，避免 route view 累积业务判断。

验证：
```bash
npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts
npm run typecheck
```

### 已完成：Batch G3 子项（contest-awd-config draft 模块拆分）
- 新增 `features/contest-awd-config/model/useAwdCheckerConfigDraft.ts`，收口以下能力：
  - checker 草稿状态（legacy/http/tcp/script）
  - 表单分值与字段错误态
  - checker config 预览 JSON 与签名派生
  - 校验、草稿 hydrate、HTTP 预设、TCP step 增删折叠
- `useContestAwdConfigPage.ts` 已改为组装层，保留路由、加载、保存、试跑流程；草稿细节不再堆叠在主 composable。
- `features/contest-awd-config/model/index.ts` 已补充 `useAwdCheckerConfigDraft` 导出。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（topology 校验/节点创建纯逻辑抽离）
- `features/challenge-topology-studio/model/topologyDraft.ts` 新增：
  - `buildTopologyDraftValidationIssues`
  - `createUniqueNodeDraft`
- `useChallengeTopologyStudioPage.ts` 已改为复用上述纯函数，删除本地重复逻辑：
  - 草稿校验问题计算由内联实现改为 `buildTopologyDraftValidationIssues(draft)`
  - `addNode` 与 `handleCanvasCreateNode` 的唯一 key 生成、默认网络挂载逻辑统一到 `createUniqueNodeDraft`
- 结果：主 composable 行为不变，但纯逻辑的可复用性与后续继续拆分（payload/loader/editor 子模块）的基础已就绪。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch I 子项（边界自动化落地）
- 新增 `views` 架构边界测试：
  - `views/__tests__/routeViewArchitectureBoundary.test.ts`
  - 规则 1：`views` 运行时代码禁止直接导入 `@/api/*`（`@/api/contracts` 类型导入除外）
  - 规则 2：`views` 运行时代码禁止直接使用 `useRoute/useRouter/router.push/router.replace/useRouteQueryTabs`
- 扩展 `features` 边界测试：
  - `features/__tests__/featureBoundaries.test.ts` 新增规则
  - `features` 运行时代码禁止导入 `@/components/*`（测试文件除外）
- 结果：将本轮迁移目标从“人工扫描结论”升级为“可执行回归约束”。

验证：
```bash
npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（contest-awd-admin 支撑模块拆分）
- 新增 `features/contest-awd-admin/model/awdAdminSupport.ts`，收口纯支撑能力：
  - 过滤器/覆盖弹窗状态类型与默认构造
  - 轮次选中持久化（sessionStorage）
  - round 选择策略 `pickRoundId`
  - readiness 阻断错误识别与通用错误文案归一
  - 空实例编排数据构造
- `usePlatformContestAwd.ts` 已移除上述内联定义，主 composable 代码聚焦在数据流与动作编排。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（readiness 决策流拆分）
- 新增 `features/contest-awd-admin/model/useAwdReadinessDecision.ts`，承接：
  - readiness 拉取与 loading 状态
  - override dialog 生命周期（打开、关闭、确认）
  - 被阻断操作的强制放行决策分支
- `usePlatformContestAwd.ts` 改为组装调用该子模块，不再内联 override 决策细节。
- `usePlatformContestAwd.ts` 行数进一步下降（本批由 864 降至 776）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（traffic 过滤状态拆分）
- 新增 `features/contest-awd-admin/model/useAwdTrafficFilterState.ts`，承接：
  - 流量筛选状态 `trafficFilters`
  - traffic events 查询参数构造
  - patch/page/pagination/reset 的状态更新
- `usePlatformContestAwd.ts` 改为通过该子模块维护 traffic filter 状态，不再内联相关细节。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 776 降至 753）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（实例编排操作拆分）
- 新增 `features/contest-awd-admin/model/useAwdServiceOperations.ts`，承接：
  - 实例编排数据刷新
  - 单队伍单服务启动
  - 单队伍批量服务启动
  - 全队伍全服务批量启动
  - 相关 loading / action key 状态
- `usePlatformContestAwd.ts` 改为组合调用 `useAwdServiceOperations`，移除内联实例操作细节。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 753 降至 649）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G3 子项（challenge 选择状态拆分）
- 新增 `features/contest-awd-config/model/useAwdChallengeSelection.ts`，承接：
  - `selectedServiceId`/`selectedService`/`selectedCheckerType` 状态与派生
  - service query 读取与 URL 同步
  - service 列表排序与选中项 reconcile
- `useContestAwdConfigPage.ts` 改为组合调用该子模块，移除内联 query/service 选择细节。
- `useContestAwdConfigPage.ts` 行数进一步下降（本批由 383 降至 347）。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（round detail 状态拆分）
- 新增 `features/contest-awd-admin/model/useAwdRoundDetailState.ts`，承接：
  - round detail 拉取（services/attacks/summary/traffic/scoreboard）
  - traffic events 拉取与并发请求 token 抑制
  - round detail 相关 loading 状态与清空逻辑
- `usePlatformContestAwd.ts` 改为组合调用该子模块，移除内联 round detail 读取细节。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 649 降至 556）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G2 子项（题目关联操作拆分）
- 新增 `features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts`，承接：
  - AWD 题目关联创建
  - AWD 题目关联更新
  - 关联列表刷新与保存 loading 状态
- `usePlatformContestAwd.ts` 改为组合调用该子模块，不再内联 challenge link 增改细节。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 556 降至 461）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（模板选择状态拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyTemplateSelection.ts`，承接：
  - 模板选择状态（keyword/id/name/description）
  - `selectedTemplate` / `canSaveTemplate` / `selectedTemplateSummary` 派生
  - 模板表单 reset / clear / reconcile
- `useChallengeTopologyStudioPage.ts` 改为组合调用该子模块，移除内联模板选择与表单状态逻辑。
- `useChallengeTopologyStudioPage.ts` 行数进一步下降（本批由 930 降至 916）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（模板增删改动作拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyTemplateMutations.ts`，承接：
  - 模板创建
  - 模板更新
  - 模板删除（含删除确认）
  - `templateBusy` 与异常提示流程
- `useChallengeTopologyStudioPage.ts` 改为组合调用该子模块，移除内联模板 CRUD 命令式流程。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 916 降至 849）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（模板应用动作拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyTemplateApply.ts`，承接：
  - 模板载入草稿
  - 模板应用到题目拓扑（含覆盖确认）
- `useChallengeTopologyStudioPage.ts` 改为组合调用该子模块，移除内联模板应用流程。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 849 降至 833）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（拓扑数据加载流程拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyDataLoader.ts`，承接：
  - 模板列表加载
  - 页面基础数据加载（模板库模式 / 题目模式）
  - `reloadAll` 编排
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联 `loadTemplates/loadPageData/reloadAll` 分支。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 833 降至 801）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（拓扑持久化动作拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyPersistenceActions.ts`，承接：
  - 题目拓扑保存
  - 题目包导出
  - 已保存拓扑删除（含删除确认）
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联保存/导出/删除流程。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 801 降至 747）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（交互绑定层拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyInteractionBindings.ts`，承接：
  - 全局键盘事件绑定与解绑
  - 首次 `reloadAll` 初始化加载
  - draft 节点数量变化后的位置归一与选中修正
  - interactionMode 变化时的 pending source 清理
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联 lifecycle/watch/keyboard 逻辑。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 747 降至 702）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（拓扑结构增删操作拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyStructureMutations.ts`，承接：
  - 网络增删
  - 节点增删
  - 连线/策略追加
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联结构增删逻辑。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 702 降至 642）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（选中边编辑逻辑拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyEdgeEditing.ts`，承接：
  - 选中边 kind 切换（link/policy 互转）
  - 选中边 source/target 更新
  - kind 字符串输入分发
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联边编辑实现。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 642 降至 595）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch G1 子项（画布交互动作拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyCanvasActions.ts`，承接：
  - 画布节点选中/连线选中
  - 画布创建节点/创建连线（link/policy）
  - 节点编辑器聚焦与交互模式切换
  - 删除选中项流程（含确认）
- `useChallengeTopologyStudioPage.ts` 改为组合调用该模块，移除内联 canvas 交互动作。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 595 降至 499）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch I 子项（跨 slice 深导入边界固化）
- 在 `src/features/__tests__/featureBoundaries.test.ts` 新增运行时代码边界断言：
  - 禁止通过 `@/features/*/model/*` 深导入其他 feature 的内部实现。
  - 同 feature 内部 `model` 相互引用仍允许，避免误伤同 slice 局部模块化。
- `AWDChallengeConfigDialog.vue` 改为通过 `@/features/contest-awd-config` 公共入口导入，不再深导入该 feature 的 `model` 私有路径。
- `duplicateActionGuardAudit.test.ts` 同步更新断言：
  - 正向要求使用 `@/features/contest-awd-config` 公共入口。
  - 负向要求不出现 `@/features/contest-awd-config/model/` 深导入。

验证：
```bash
npm run test:run -- src/features/__tests__/featureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts
npm run typecheck
```

### 已完成：Batch J 子项（拓扑选中态与草稿应用逻辑拆分）
- 新增 `features/challenge-topology-studio/model/useTopologySelectionState.ts`，承接：
  - 当前选中节点/边的派生（`selectedNodeDraft`、`selectedEdgeMeta`、`selectedLinkDraft`、`selectedPolicyDraft`）
  - 选中边源/目标/kind 派生与 `selectedCanvasSummary`
  - `applyTopologyDraft` 与 `syncEntryNode` 草稿应用流程
  - 节点端口快捷更新、节点网络勾选切换
- `useChallengeTopologyStudioPage.ts` 改为组合调用该子模块，移除内联选中态与草稿同步细节。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 499 降至 418）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch K 子项（拓扑页面展示派生拆分）
- 新增 `features/challenge-topology-studio/model/useTopologyStudioPresentation.ts`，承接：
  - page header / hero / status card / secondary card
  - 题包来源摘要、基线摘要、修订记录、文件列表
  - `nodeOptions`、`topologySummary`、`canvasModeLabel`
- `useChallengeTopologyStudioPage.ts` 改为组合调用该子模块，移除大段展示派生逻辑。
- `useChallengeTopologyStudioPage.ts` 行数继续下降（本批由 418 降至 303）。

验证：
```bash
npm run test:run -- src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
npm run typecheck
```

### 已完成：Batch L 子项（AWD 轮次动作流程拆分）
- 新增 `features/contest-awd-admin/model/useAwdRoundOperations.ts`，承接：
  - 轮次巡检触发（当前轮 / 指定轮）
  - 轮次创建（含 readiness override 分支）
  - 服务检查记录创建
  - 攻击日志记录创建
- `usePlatformContestAwd.ts` 改为组合调用该模块，移除内联轮次动作流程与相关 loading 状态。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 461 降至 368）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch M 子项（AWD 生命周期绑定拆分）
- 新增 `features/contest-awd-admin/model/useAwdLifecycleBindings.ts`，承接：
  - contest 切换触发刷新与 traffic filter reset
  - round 选择变化触发明细刷新
  - round 选择持久化
  - 自动刷新定时器启停与卸载清理
- `usePlatformContestAwd.ts` 改为组合调用该模块，移除内联 watch/onBeforeUnmount 绑定细节。
- `usePlatformContestAwd.ts` 行数继续下降（本批由 368 降至 317）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch N 子项（前端架构 skill 记录与索引）
- 新增主索引：
  - `code/docs/skills/SKILL.md`
  - 收录本仓库前端 skill 入口与使用场景
- 新增前端架构 skill：
  - `code/docs/skills/frontend-sliced-architecture/SKILL.md`
  - 同步本轮迁移的分层约束、标准扫描命令、最小验证集与记录链接
- 结果：后续前端迁移任务可直接复用项目内 skill，不再依赖单次会话说明。

### 已完成：Batch O 子项（平台赛事 AWD 启动 override 流程拆分）
- 新增 `features/platform-contests/model/useAwdStartOverrideFlow.ts`，承接：
  - AWD 启动前 readiness 拉取
  - override dialog 状态维护（open/close/confirmLoading/pendingPayload）
  - override 确认后强制更新与错误分支处理
- `usePlatformContests.ts` 改为组合调用该模块，移除内联 AWD override 流程细节。
- `usePlatformContests.ts` 行数下降（本批由 404 降至 343）。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts
npm run typecheck
```

### 已完成：Batch P 子项（竞赛详情页队伍动作流程拆分）
- 新增 `features/contest-detail/model/useContestTeamActions.ts`，承接：
  - 创建队伍弹窗状态与提交流程
  - 加入队伍弹窗状态与提交流程
  - 踢出队员确认与提交动作
- `useContestDetailPage.ts` 改为组合调用该模块，移除内联队伍动作流程细节。
- `useContestDetailPage.ts` 行数下降（本批由 438 降至 355）。

验证：
```bash
npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts
npm run typecheck
```

### 已完成：Batch Q 子项（教师学员分析复盘归档流程拆分）
- 新增 `features/teacher-student-analysis/model/useReviewArchiveExportFlow.ts`，承接：
  - 复盘归档导出触发
  - 状态轮询与失败分支处理
  - 报告下载动作与导出提示
  - 报告弹窗可见性状态
- `useTeacherStudentAnalysisPage.ts` 改为组合调用该模块，移除内联复盘导出细节。
- `useTeacherStudentAnalysisPage.ts` 行数下降（本批由 505 降至 420）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts
npm run typecheck
```

### 已完成：Batch R 子项（实例页告警状态流拆分）
- 新增 `features/instance-list/model/useInstanceWarningState.ts`，承接：
  - 运行实例剩余时间倒计时更新
  - 临期阈值告警触发
  - 告警弹窗关闭 / Esc 关闭 / 告警内延时动作
- `useInstanceListPage.ts` 改为组合调用该模块，移除内联告警状态流程。
- `useInstanceListPage.ts` 行数下降（本批由 386 降至 352）。

验证：
```bash
npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts
npm run typecheck
```

### 已完成：Batch S 子项（教师分析题解/评审流程拆分）
- 新增 `features/teacher-student-analysis/model/useTeacherSubmissionReviewFlows.ts`，承接：
  - 题解分页刷新与翻页动作
  - 社区题解推荐/取消推荐/隐藏/恢复
  - 人工评审详情读取与审批提交
  - 题解与人工评审状态重置/写入
- `useTeacherStudentAnalysisPage.ts` 改为组合调用该模块，移除内联题解与评审流程细节。
- `useTeacherStudentAnalysisPage.ts` 行数下降（本批由 420 降至 315）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts
npm run typecheck
```

### 已完成：Batch T 子项（竞赛编排保存/移除动作拆分）
- 新增 `features/contest-workbench/model/useContestChallengeMutations.ts`，承接：
  - Jeopardy/AWD 题目保存流程
  - AWD 批量关联失败汇总与提示
  - 题目移除确认与删除流程
- `useContestChallengeOrchestration.ts` 改为组合调用该模块，移除内联保存/删除命令式实现。
- `useContestChallengeOrchestration.ts` 行数下降（本批由 398 降至 271）。

验证：
```bash
npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
npm run typecheck
```

### 已完成：Batch U 子项（平台赛事弹窗状态拆分）
- 新增 `features/platform-contests/model/useContestDialogState.ts`，承接：
  - 创建态初始化
  - 编辑态草稿装载
  - 弹窗开关与编辑上下文状态
- `usePlatformContests.ts` 改为组合调用该模块；`saveContest`、AWD readiness override 与分页刷新仍保留在主流程组合器中。
- 结果：主模块职责进一步收敛为“提交/校验/刷新编排”，弹窗状态不再与保存逻辑混杂。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts
npm run typecheck
```

### 已完成：Batch V 子项（竞赛详情 Flag 提交流程拆分）
- 新增 `features/contest-detail/model/useContestFlagSubmission.ts`，承接：
  - 题目选择后提交流程状态（`selectedChallenge`、`flagInput`、`submitting`、`submitResult`）
  - Flag 字段校验与提交流程
  - 正确提交后的题目解出态同步
- `useContestDetailPage.ts` 改为组合调用该模块，移除内联提交命令式逻辑。
- `useContestDetailPage.ts` 行数下降（本批由 355 降至 299）。

验证：
```bash
npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts
npm run typecheck
```

### 已完成：Batch W 子项（实例操作动作流拆分）
- 新增 `features/instance-list/model/useInstanceOperations.ts`，承接：
  - 地址复制
  - 实例延时
  - 打开目标（含 TCP 命令复制）
  - 销毁实例确认与删除
- `useInstanceListPage.ts` 改为组合调用该模块，移除内联操作动作细节。
- `useInstanceListPage.ts` 行数下降（本批由 352 降至 270）。

验证：
```bash
npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts
npm run typecheck
```

### 已完成：Batch X 子项（平台赛事保存流程拆分）
- 新增 `features/platform-contests/model/useContestSaveFlow.ts`，承接：
  - 赛事创建/更新提交流程
  - 结束赛事确认分支
  - AWD 启动 gate 拦截与 override 弹层分支
  - 提交成功后的关闭与列表刷新编排
- `usePlatformContests.ts` 改为组合调用该模块，移除内联保存命令式实现。
- `usePlatformContests.ts` 行数下降（本批由 346 降至 301）。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts
npm run typecheck
```

### 已完成：Batch Y 子项（题目详情数据加载层拆分）
- 新增 `features/challenge-detail/model/useChallengeDetailDataLoader.ts`，承接：
  - 题目详情加载
  - 解出后推荐/社区题解并发加载
  - 请求竞态保护（最新请求 token）
  - 失败分支与题解列表清理
- `useChallengeDetailPage.ts` 改为组合调用该模块，移除内联 `loadChallenge/loadSolutions` 实现。
- `useChallengeDetailPage.ts` 行数下降（本批由 319 降至 277）。

验证：
```bash
npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts
npm run typecheck
```

### 已完成：Batch Z 子项（教师 AWD 复盘导出流程拆分）
- 新增 `features/teacher-awd-review/model/useTeacherAwdReviewExportFlow.ts`，承接：
  - 复盘归档/教师报告导出
  - 轮询状态跟进与失败分支
  - 生成后下载动作与提示
- `useTeacherAwdReviewDetail.ts` 改为组合调用该模块，移除内联导出与轮询实现。
- `useTeacherAwdReviewDetail.ts` 行数下降（本批由 352 降至 224）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
npm run typecheck
```

### 已完成：Batch AA 子项（题目详情个人题解流程拆分）
- 新增 `features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`，承接：
  - 个人题解加载
  - 题解草稿/发布保存
  - 题解表单状态 hydrate/reset
- `useChallengeDetailInteractions.ts` 改为组合调用该模块，移除内联个人题解加载与保存实现。
- `useChallengeDetailInteractions.ts` 行数下降（本批由 318 降至 260）。

验证：
```bash
npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts
npm run typecheck
```

### 已完成：Batch AB 子项（教师学员分析导航动作拆分）
- 新增 `features/teacher-student-analysis/model/useTeacherStudentAnalysisNavigation.ts`，承接：
  - 班级/学生切换跳转
  - 班级管理页跳转
  - 题目详情跳转
  - 复盘归档页面跳转
- `useTeacherStudentAnalysisPage.ts` 改为组合调用该模块，移除内联路由跳转逻辑。
- `useTeacherStudentAnalysisPage.ts` 行数下降（本批由 315 降至 280）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts
npm run typecheck
```

### 已完成：Batch AC 子项（AWD 快照刷新加载流程拆分）
- 新增 `features/contest-awd-admin/model/useAwdContestSnapshotLoader.ts`，承接：
  - 赛事维度快照并发刷新（rounds/teams/services/readiness/orchestration）
  - 非 AWD 赛事分支下的状态清理
  - round 选择跟随与请求竞态 token 防护
  - 刷新后 round detail 衔接调用
- `usePlatformContestAwd.ts` 改为组合调用该模块，移除内联快照刷新细节。
- `usePlatformContestAwd.ts` 行数下降（本批由 317 降至 252）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts
npm run typecheck
```

### 已完成：Batch AD 子项（AWD 配置页预览流程拆分）
- 升级 `features/contest-awd-config/model/useAwdCheckerPreview.ts`，新增 `useAwdCheckerPreviewFlow`，承接：
  - 试跑状态与表单状态（`previewing/previewResult/previewError/previewToken/previewForm`）
  - 预览 token 与配置签名关联校验（`canAttachPreviewToken`）
  - 签名变化后的 token 失效处理（`handleSignatureChange`）
  - 试跑请求构造与错误分支处理
- `useContestAwdConfigPage.ts` 改为组合调用该模块，移除内联试跑状态和请求实现。
- `useContestAwdConfigPage.ts` 行数下降（本批由 347 降至 300）。
- 补充 source 边界断言：`ContestAwdConfig.test.ts` 增加 `useAwdCheckerPreviewFlow` 导入断言与 `runContestAWDCheckerPreview` 禁止断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts
npm run typecheck
```

### 已完成：Batch AE 子项（AWD 配置页保存提交流程拆分）
- 新增 `features/contest-awd-config/model/useAwdCheckerSaveFlow.ts`，承接：
  - 保存中状态控制（`saving`）
  - checker payload 组装与保存请求
  - 试跑 token 绑定保存分支
  - 保存成功后的提示与页面刷新
- `useContestAwdConfigPage.ts` 改为组合调用该模块，移除内联保存提交流程。
- `useContestAwdConfigPage.ts` 行数下降（本批由 300 降至 279）。
- 补充 source 边界断言：`ContestAwdConfig.test.ts` 增加 `useAwdCheckerSaveFlow` 导入断言与 `updateContestAWDService` 禁止断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts
npm run typecheck
```

### 已完成：Batch AF 子项（AWD 配置展示标签纯函数下沉）
- 新增 `features/contest-awd-config/model/awdCheckerLabels.ts`，承接：
  - checker 类型标签映射
  - 协议标签映射
  - 校验状态标签映射
  - 检查时间格式化
- `useContestAwdConfigPage.ts` 改为复用纯函数模块，移除内联展示映射函数。
- `useContestAwdConfigPage.ts` 行数下降（本批由 279 降至 229）。
- 补充 source 边界断言：`ContestAwdConfig.test.ts` 增加 `awdCheckerLabels` 导入断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts
npm run typecheck
```

### 已完成：Batch AG 子项（AWD 配置页数据加载流程拆分）
- 新增 `features/contest-awd-config/model/useContestAwdConfigDataLoader.ts`，承接：
  - 页面加载/刷新状态（`loading/refreshing`）
  - 配置加载错误态（`loadError`）
  - 赛事与服务列表并发加载与竞态版本控制
  - 面包屑标题同步与卸载清理
- `useContestAwdConfigPage.ts` 改为组合调用该模块，移除内联加载并发控制逻辑。
- `useContestAwdConfigPage.ts` 行数下降（本批由 229 降至 201）。
- 补充 source 边界断言：`ContestAwdConfig.test.ts` 增加 `useContestAwdConfigDataLoader` 导入断言与 `getContest/listContestAWDServices` 禁止断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts
npm run typecheck
```

### 已完成：Batch AH 子项（AWD 战场服务启停动作拆分）
- 新增 `features/contest-awd-workspace/model/useAwdWorkspaceServiceActions.ts`，承接：
  - 服务启动/重启动作
  - 服务动作中的并发互斥与 pending 标记
  - 刷新后 settled 状态清理
  - 动作成功/失败提示
- `useContestAWDWorkspace.ts` 改为组合调用该模块，主流程保留“战场数据刷新与整体编排”职责。
- `useContestAWDWorkspace.ts` 行数下降（本批由 347 降至 284）。

验证：
```bash
npm run test:run -- src/features/contest-awd-workspace/model/useContestAWDWorkspace.test.ts
npm run typecheck
```

### 已完成：Batch AI 子项（AWD 战场访问动作拆分）
- 新增 `features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`，承接：
  - 本队实例访问打开
  - 防守 SSH 连接生成与缓存
  - 目标服务访问打开
  - 访问动作互斥 key 状态管理
- `useContestAWDWorkspace.ts` 改为组合调用该模块，主流程保留刷新、自动轮询与攻击提交流程。
- `useContestAWDWorkspace.ts` 行数下降（本批由 284 降至 220）。

验证：
```bash
npm run test:run -- src/features/contest-awd-workspace/model/useContestAWDWorkspace.test.ts
npm run typecheck
```

### 已完成：Batch AJ 子项（AWD 战场攻击提交流程拆分）
- 新增 `features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`，承接：
  - 攻击提交流程状态（`submittingKey`、`submitResult`）
  - 提交参数标准化与互斥控制
  - 提交后刷新与结果提示
  - 自定义攻击结果 toast 文案透传
- `useContestAWDWorkspace.ts` 改为组合调用该模块，主流程进一步收敛为刷新、轮询与子动作编排。
- `useContestAWDWorkspace.ts` 行数下降（本批由 220 降至 185）。

验证：
```bash
npm run test:run -- src/features/contest-awd-workspace/model/useContestAWDWorkspace.test.ts
npm run typecheck
```

### 已完成：Batch AK 子项（教师看板洞察文案构建拆分）
- 新增 `features/teacher-dashboard/model/teacherDashboardInsightBuilders.ts`，承接：
  - 学生洞察行构建（`studentInsightRows`）
  - 画像摘要构建（`portraitSummaryNotes`）
  - 趋势信号构建（`trendSignals`）
- `useTeacherDashboardMetrics.ts` 改为组合调用该纯函数模块，移除内联的大段文案构建逻辑。
- `useTeacherDashboardMetrics.ts` 行数下降（本批由 336 降至 233）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherDashboard.test.ts
npm run typecheck
```

### 已完成：Batch AL 子项（拓扑组件 helper 对齐 feature model）
- `components/platform/topology/topologyDraft.ts` 改为转发导出 `features/challenge-topology-studio/model/topologyDraft.ts`。
- `components/platform/topology/topologyLayout.ts` 改为转发导出 `features/challenge-topology-studio/model/topologyLayout.ts`。
- 结果：拓扑组件层不再维护一份重复的 draft/layout 实现，后续变更统一在 feature model 演进，避免双份逻辑漂移。
- 结果：组件层 API 直连扫描中，`topologyDraft.ts` 的 `@/api/admin/authoring` 依赖已清除。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts
npm run typecheck
```

### 已完成：Batch AM 子项（镜像创建弹窗类型依赖下沉）
- 新增 `entities/image`：
  - `entities/image/model/createForm.ts`
  - `entities/image/model/index.ts`
  - `entities/image/index.ts`
- `ImageCreateModal.vue` 改为依赖 `entities/image` 的 `ImageCreateForm`，不再直接导入 `@/api/admin/authoring`。
- `useImageManagePage.ts` 改为复用 `ImageCreateForm` 和 `createEmptyImageCreateForm`，统一创建表单初始值与重置逻辑。
- 补充 source 边界断言：`ImageManage.test.ts` 增加 `ImageCreateModal` 不导入 `@/api/admin/authoring` 的断言。
- 结果：组件层 API 直连扫描仅剩测试断言源码引用，不再有运行时代码命中。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts
npm run typecheck
```

### 已完成：Batch AN 子项（镜像管理页面动作与轮询拆分）
- 新增 `features/image-management/model/useImageManageMutations.ts`，承接：
  - 创建镜像提交
  - 删除镜像确认与提交流程
  - 创建中状态控制（`creating`）
- 新增 `features/image-management/model/useImageManageAutoRefresh.ts`，承接：
  - 构建中镜像轮询启动/停止
  - 页面挂载初次刷新
  - 轮询提示文案派生（`refreshHint`）
- `useImageManagePage.ts` 改为组合调用上述模块，移除内联 mutation 与轮询细节。
- `useImageManagePage.ts` 行数下降（本批由 317 降至 238）。
- 补充 source 边界断言：`ImageManage.test.ts` 增加 `useImageManagePage` 已接入两段新 composable 且不再内联 `createImage/deleteImage` 的断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts
npm run typecheck
```

### 已完成：Batch AO 子项（平台赛事表单 support 层拆分）
- 新增 `features/platform-contests/model/contestFormSupport.ts`，承接：
  - 赛事表单 draft 类型（`ContestFormDraft`）
  - 状态与字段锁规则（`PlatformContestStatus`、`createFieldLocks`、`createContestStatusOptions`）
  - draft 与 payload 映射（`createDraftFromContest`、`buildContestUpdatePayload`）
  - 时间转换与状态确认规则（`toISOString`、`shouldConfirmContestTermination`）
- `usePlatformContests.ts` 改为复用 support 层，移除内联表单规则实现。
- `useContestDialogState.ts`、`useContestSaveFlow.ts` 改为直接依赖 `contestFormSupport.ts` 类型，消除对 `usePlatformContests` 的反向类型依赖。
- 新增边界测试：`platformContestsModelBoundary.test.ts`，锁定子模块不再导入 `usePlatformContests`。
- `usePlatformContests.ts` 行数下降（本批由 301 降至 178）。

验证：
```bash
npm run test:run -- src/features/platform-contests/model/platformContestsModelBoundary.test.ts src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts
npm run typecheck
```

### 已完成：Batch AP 子项（平台赛事列表查询状态拆分）
- 新增 `features/platform-contests/model/useContestListState.ts`，承接：
  - `statusFilter` 状态
  - 赛事列表分页请求与参数注入
  - 筛选切换后的分页重置（回到第一页）
- `usePlatformContests.ts` 改为组合调用该模块，移除内联 `usePagination + watch` 逻辑。
- 扩展边界测试：`platformContestsModelBoundary.test.ts` 增加 `usePlatformContests` 必须引用 `useContestListState` 的断言，并禁止再次内联分页请求实现。
- `usePlatformContests.ts` 行数下降（本批由 178 降至 165）。

验证：
```bash
npm run test:run -- src/features/platform-contests/model/platformContestsModelBoundary.test.ts src/views/platform/__tests__/ContestManage.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts
npm run typecheck
```

### 已完成：Batch AQ 子项（镜像管理展示派生纯函数拆分）
- 新增 `features/image-management/model/imageManagePresentation.ts`，承接：
  - 镜像列表筛选与排序（`filterAndSortImages`）
  - 状态摘要构建（`buildImageStatusSummary`）
  - 状态标签与样式映射（`getImageStatusLabel/getImageStatusStyle`）
  - 展示格式化（`formatImageSize/formatImageDateTime`）
- `useImageManagePage.ts` 改为组合调用该模块，移除内联筛选/排序与展示格式化细节。
- `useImageManagePage.ts` 行数下降（本批由 238 降至 130）。
- 补充 source 边界断言：`ImageManage.test.ts` 增加 `imageManagePresentation` 接入断言与核心派生函数调用断言。

验证：
```bash
npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts
npm run typecheck
```

### 已完成：Batch AR 子项（AWD 管理流量动作与状态标志拆分）
- 新增 `features/contest-awd-admin/model/useAwdTrafficActions.ts`，承接：
  - 流量筛选应用
  - 流量分页切换
  - 流量筛选重置
  - 统一衔接 `refreshTrafficEvents`
- 新增 `features/contest-awd-admin/model/useAwdContestStateFlags.ts`，承接：
  - `hasSelectedContest` 派生
  - `shouldAutoRefresh` 派生
- `usePlatformContestAwd.ts` 改为组合调用上述模块，移除内联动作与状态派生实现。
- 新增边界测试：`usePlatformContestAwdBoundary.test.ts`，锁定主组合器不再回退到内联流量动作实现。
- `usePlatformContestAwd.ts` 行数下降（本批由 252 降至 224）。

验证：
```bash
npm run test:run -- src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts
npm run typecheck
```

### 已完成：Batch AS 子项（教师看板概览派生构建拆分）
- 新增 `features/teacher-dashboard/model/teacherDashboardOverviewBuilders.ts`，承接：
  - `overviewDescription` 构建
  - `metaPills` 构建
  - `overviewMetrics` 构建
  - `interventionTips` 与 `teachingAdvice` 构建
- `useTeacherDashboardMetrics.ts` 改为组合调用该模块，移除内联概览与建议文案构建逻辑。
- 新增边界测试：`useTeacherDashboardMetricsBoundary.test.ts`，锁定主模块已接入 overview builders。
- `useTeacherDashboardMetrics.ts` 行数下降（本批由 233 降至 193）。

验证：
```bash
npm run test:run -- src/views/teacher/__tests__/TeacherDashboard.test.ts src/features/teacher-dashboard/model/useTeacherDashboardMetricsBoundary.test.ts
npm run typecheck
```

### 已完成：Batch AT 子项（平台 AWD 题目导入流程拆分）
- 新增 `features/platform-awd-challenges/model/useAwdChallengeImportFlow.ts`，承接：
  - 导入队列加载
  - 题目包预览上传
  - 预览提交导入
  - 上传结果轨迹记录与状态管理（`uploading/queueLoading/uploadResults`）
- `usePlatformAwdChallenges.ts` 改为组合调用该模块，移除内联导入流程实现。
- 新增边界测试：`usePlatformAwdChallengesBoundary.test.ts`，锁定主组合器不再回退到内联导入流程函数。
- `usePlatformAwdChallenges.ts` 行数下降（本批由 300 降至 238）。

验证：
```bash
npm run test:run -- src/features/platform-awd-challenges/model/usePlatformAwdChallenges.test.ts src/features/platform-awd-challenges/model/usePlatformAwdChallengesBoundary.test.ts
npm run typecheck
```

## 每批验证要求
1. 运行本批相关 vitest。
2. 运行 `npm run typecheck`。
3. 更新 source 边界断言，至少覆盖“不含 `onMounted`/业务依赖”或“不含被迁移的直接 import”。
4. 每批单独提交，commit message 使用中文并说明具体页面。
