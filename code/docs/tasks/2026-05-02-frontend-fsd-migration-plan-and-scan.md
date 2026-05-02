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

## 每批验证要求
1. 运行本批相关 vitest。
2. 运行 `npm run typecheck`。
3. 更新 source 边界断言，至少覆盖“不含 `onMounted`/业务依赖”或“不含被迁移的直接 import”。
4. 每批单独提交，commit message 使用中文并说明具体页面。
