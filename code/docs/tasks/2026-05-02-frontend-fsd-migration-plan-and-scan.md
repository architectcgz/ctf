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
rg -n "onMounted\(|watch\(" code/frontend/src/views/**/*.vue
```
当前命中：
- `src/views/auth/LoginView.vue`（控制台提示与彩蛋提示时序）
- `src/views/platform/ContestManage.vue`
- `src/views/platform/UserManage.vue`
- `src/views/platform/AWDChallengeLibrary.vue`
- `src/views/platform/AWDChallengeImport.vue`
- `src/views/instances/InstanceList.vue`
- `src/views/scoreboard/ScoreboardView.vue`（本地分页重置）

## 当前未完成项

### P0：业务生命周期仍在 route view 中
- `src/views/platform/ContestManage.vue`
  - 当前仍由 route view 执行 `onMounted(() => refresh())`。
  - route view 还持有创建后切回列表、公告抽屉状态、状态筛选 setter 等页面流程。
  - 建议新增 `features/platform-contests/model/useContestManagePage.ts`，由它组合 `usePlatformContests`、`requestedPanel`、公告抽屉状态和创建后跳转逻辑。
- `src/views/platform/UserManage.vue`
  - 当前仍由 route view 执行 `onMounted(() => refresh())`。
  - 删除确认流程仍在 route view 中，且直接使用 `confirmDestructiveAction`。
  - 建议新增 `features/platform-users/model/usePlatformUserManagePage.ts`，由它组合 `usePlatformUsers`、筛选 setter、弹窗关闭和删除确认流程。
- `src/views/platform/AWDChallengeLibrary.vue`
  - 当前仍由 route view 执行 `onMounted(() => refresh())`。
  - 筛选 setter 和弹窗关闭逻辑仍在 route view 中。
  - 建议扩展现有 `useAwdChallengeLibraryPage`，让它组合 `usePlatformAwdChallenges`、导入页跳转、列表刷新和筛选 setter。
- `src/views/platform/AWDChallengeImport.vue`
  - 当前仍由 route view 执行 `onMounted(() => refreshImportQueue())`。
  - 建议新增 `useAwdChallengeImportPage` 或扩展 AWD 题库 page composable，把导入队列加载和导入动作下沉。

### P1：展示状态仍在 route view 中，需要评估后再迁移
- `src/views/scoreboard/ScoreboardView.vue`
  - `watch(sections)` 和 `watch(contestTotalPages)` 只处理本地分页复位与页码裁剪。
  - 目前可判定为 UI 展示状态，优先级低；若继续收口，建议新增 `useScoreboardContestDirectoryPage`，承接竞赛榜分页、汇总指标、日期格式化和 CSS class 派生。
- `src/views/instances/InstanceList.vue`
  - `watch(showWarning)` 用于 warning dialog 焦点管理，属于 UI 可访问性状态。
  - 可保留在 view，或下沉为 `useInstanceWarningFocus` 这类局部 UI composable；不建议放入业务 feature 核心模型。
- `src/views/auth/LoginView.vue`
  - `onMounted` 只负责控制台提示；不是业务数据加载。
  - 可保留，除非后续要把登录页彩蛋/控制台提示整体抽成 `useLoginProbeConsolePage`。

### P2：类型与展示 helper 仍留在 route view
- `src/views/platform/ContestManage.vue`
  - 仍从 `@/api/contracts` 导入 `ContestDetailData` 作为公告抽屉状态类型。
  - 若 P0 新增 `useContestManagePage`，应一并把该类型依赖从 route view 移到 feature。
- `src/views/scoreboard/ScoreboardView.vue`
  - 仍从 `@/api/contracts` 导入 `ContestStatus`，并在 route view 中持有 `formatDateTime`、`formatContestWindow`、`sectionAccentStyle`、`getRowClass`、`getRankPillClass`、`getCardDescription`。
  - 若 P1 迁移 scoreboard 展示状态，应一并把这些 helper 放入 feature model 或局部 presentation helper。

## 后续迁移批次

### Batch A：平台竞赛管理页
- 目标文件：`ContestManage.vue`、`features/platform-contests/model/index.ts`、新增 `useContestManagePage.ts`、对应 `ContestManage.test.ts` 或 source 边界测试。
- 迁移内容：`refresh onMounted`、状态筛选 setter、创建后切回列表、公告抽屉状态、dialog open change。
- 验收：`ContestManage.vue` 不再直接 `onMounted`，不再直接持有 `ContestDetailData`，只消费 `useContestManagePage`。

### Batch B：平台用户管理页
- 目标文件：`UserManage.vue`、`features/platform-users/model/index.ts`、新增 `usePlatformUserManagePage.ts`、对应测试。
- 迁移内容：`refresh onMounted`、筛选 setter、删除确认流程、dialog open change。
- 验收：`UserManage.vue` 不再直接 `onMounted`，不再直接 import `confirmDestructiveAction`。

### Batch C：AWD 题库与导入页
- 目标文件：`AWDChallengeLibrary.vue`、`AWDChallengeImport.vue`、`features/platform-awd-challenges/model/*`、对应测试。
- 迁移内容：列表页刷新、导入队列刷新、筛选 setter、dialog open change、导入页 page composable。
- 验收：两个 route view 不再直接 `onMounted`，只消费 `useAwdChallengeLibraryPage` / `useAwdChallengeImportPage`。

### Batch D：低优先级展示收口
- 目标文件：`ScoreboardView.vue`、`InstanceList.vue`、必要时 `LoginView.vue`。
- 迁移内容：scoreboard 本地分页和展示 helper；instance warning 焦点管理；login console/probe 副作用。
- 验收：只迁移能明确降低 route view 复杂度的部分；纯局部 UI 状态可保留并在文档中标记为 accepted exception。

## 每批验证要求
1. 运行本批相关 vitest。
2. 运行 `npm run typecheck`。
3. 更新 source 边界断言，至少覆盖“不含 `onMounted`/业务依赖”或“不含被迁移的直接 import”。
4. 每批单独提交，commit message 使用中文并说明具体页面。
