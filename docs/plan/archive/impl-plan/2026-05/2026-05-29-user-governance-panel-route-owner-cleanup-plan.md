> 状态：Current
> 事实源：UserManage route view、platform user management page owner、panel query helper、前端架构 allowlist
> 替代：无

# User Governance Panel Route Owner Cleanup Plan

## 目标

- 把 `panel=overview/import` 的 route/query owner 从 `useUserGovernancePanelRoute.ts` 收回 `usePlatformUserManagePage.ts`。
- 让 `UserGovernancePage.vue` 变回纯 props / emits UI shell。
- 去掉 `useUserGovernancePanelRoute.ts -> vue-router` 这条非 page helper allowlist。

## 非目标

- 不改变用户列表、导入、创建、编辑、删除和分页的业务逻辑。
- 不进一步拆 `usePlatformUserManagePage.ts` 或 `usePlatformUsers.ts`。
- 不处理其它 router allowlist。

## 输入依据

- `code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts`
- `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useUserGovernancePanelRoute.ts` 只是 panel query helper，不应继续作为 router owner。
- `usePlatformUserManagePage.ts` 已经是 page owner，承接 `useRoute()` / `useRouter()` 更合理。
- 这次收口的核心不是减少所有 router import，而是把 router import 放到对的位置。

## 设计边界

### `usePlatformUserManagePage.ts` 本轮负责

- 读取 `route.query.panel`
- 归一化当前 `activePanel`
- 执行 `switchPanel()` 时的 `router.replace()`
- 继续承接用户治理页面的 page-level workflow

### `useUserGovernancePanelRoute.ts` 本轮负责

- 仅保留 `panel` 解析 / query 构建 helper
- 不直接依赖 `vue-router`

### `UserGovernancePage.vue` 本轮负责

- 消费 `activePanel` prop
- 通过 emit 请求外部切换 panel
- 不直接接触 router

## 任务切片

### Slice 1：panel helper 去掉 `vue-router`

- 目标：
  - 把 `useUserGovernancePanelRoute.ts` 改为纯 helper
  - 保留 `UserPanelKey`、panel 解析和 query 构建能力
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts`
- Review focus：
  - helper source 是否不再出现 `useRoute` / `useRouter` / `router.replace`

### Slice 2：page owner 承接 query owner

- 目标：
  - 在 `usePlatformUserManagePage.ts` 内承接 `activePanel` / `switchPanel`
  - `UserGovernancePage.vue` 改为纯 props / emits
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/UserManage.test.ts`
- Review focus：
  - owner 是否明确回到 page model
  - UI shell 是否不再直接依赖 router helper

### Slice 3：allowlist / backlog 收尾

- 目标：
  - 删除旧 allowlist，新增 page owner allowlist
  - 更新 raw-source 护栏与 backlog 说明
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/UserManage.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - allowlist owner 是否从 helper 正确迁到 page owner
  - 不应留下 UI shell 级别的 router 漂移

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts src/views/platform/__tests__/UserManage.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/user-governance-panel-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-user-governance-panel-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-user-governance-panel-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.test.ts code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts code/frontend/src/features/platform-user-management/model/index.ts code/frontend/src/features/platform-user-management/ui/UserGovernancePage.vue code/frontend/src/views/platform/UserManage.vue code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮更多是在纠正 router owner 落点，`featureRouterImportAllowlist` 总数未必下降。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
