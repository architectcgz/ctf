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

## 下一批迁移建议
1. 优先迁移 `ContestManage.vue`、`UserManage.vue` 的生命周期业务编排到对应 feature page composable。  
2. 评估 `ScoreboardView.vue`、`InstanceList.vue` 的 `watch` 是否属于纯展示状态：  
   - 若是纯展示（本地分页、UI 同步），可保留；  
   - 若牵涉业务流程，继续下沉。  
3. 每批保持最小可审阅提交，并同步增加 source 边界测试。
