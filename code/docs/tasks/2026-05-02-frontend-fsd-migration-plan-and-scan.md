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

## 每批验证要求
1. 运行本批相关 vitest。
2. 运行 `npm run typecheck`。
3. 更新 source 边界断言，至少覆盖“不含 `onMounted`/业务依赖”或“不含被迁移的直接 import”。
4. 每批单独提交，commit message 使用中文并说明具体页面。
