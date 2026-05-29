> 状态：Current
> 事实源：platform user management page model、user governance route helper、route query transport、前端架构 allowlist
> 替代：无

# Platform User Management Route Owner Cleanup Plan

## 目标

- 收掉 `features/platform-user-management/model/usePlatformUserManagePage.ts -> vue-router`

## 非目标

- 不改用户治理页的列表、导入、创建、编辑、删除和分页 workflow。
- 不把 `panel` query owner 再抽成新的 feature route wrapper。
- 不改 `usePlatformUsers.ts` 的加载、筛选和节流 owner。

## 输入依据

- `usePlatformUserManagePage.ts`
- `useUserGovernancePanelRoute.ts`
- `routeQueryTransport.ts`
- `UserManage.vue`
- `UserManage.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- `platform-user-management` 这条的 route owner 只有 `panel` query hydrate 与 `replace`，不涉及 route params、角色跳转或复杂导航 contract。
- `routeQueryTransport.ts` 已经能承接“读当前 query + replace query”这层 transport，本轮直接复用比新增 wrapper 更小。
- `usePlatformUserManagePage.ts` 继续保留 panel normalize、mounted refresh 与用户治理 workflow owner，不把这些业务语义平移进 shared composable。

## 设计边界

### platform user management page model 本轮负责

- 保留 `activePanel` 的 panel normalize 与 `switchPanel()` owner
- 保留 mounted refresh、筛选、分页、删除确认和表单 workflow
- 不再直接 import `vue-router`

### shared transport 本轮负责

- 提供当前 query 读取与 `replaceQuery()`
- 不承接 user-management-specific panel schema、默认值或 mounted policy

### user governance route helper 本轮负责

- 保留 `panel` 解析与 query 构建 helper
- 不承接 router transport

## 任务切片

- [ ] Slice 1：page model 改用 query transport
  - 目标：
    - `usePlatformUserManagePage.ts` 去掉 `vue-router`
    - `activePanel` 与 `switchPanel()` 改为消费 `routeQueryTransport.ts`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/UserManage.test.ts`

- [ ] Slice 2：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、源码护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/UserManage.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/UserManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-user-management-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-user-management-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-user-management-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `routeQueryTransport.ts` 仍然应该只做 transport；如果后续 user governance 出现更复杂的 query canonicalization，不应继续把 normalize/default 堆进 shared。
- 本轮只收 transport owner，不改变 `usePlatformUsers.ts` 的 mounted refresh 和筛选节流行为，需要通过现有页面测试确认交互不变。
