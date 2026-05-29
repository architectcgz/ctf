> 状态：Current
> 事实源：challenge list page model、challenge directory UI、前端架构 allowlist
> 替代：无

# Challenge List Route Owner Cleanup Plan

## 目标

- 收掉 `features/challenge-list/model/useChallengeListPage.ts -> vue-router`

## 非目标

- 不重做 challenge list 的视觉结构和分页能力。
- 不把 search keyword 也同步进 route query。
- 不引入新的 feature route wrapper 中间态。

## 输入依据

- `useChallengeListPage.ts`
- `ChallengeList.vue`
- `ChallengeDirectoryPanel.vue`
- `ChallengeDirectoryRow.vue`
- `ChallengeList.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- route query sync 只覆盖 `category / difficulty` 两个筛选项，属于 challenge list page 自己的能力 owner。
- 返回仪表盘、进入能力画像和进入题目详情都属于薄导航，更适合改成 route target contract。
- 如果只是新建一个 `useChallengeListRoutePage.ts` 去继续持有 router，allowlist 只是换壳，不是真收口。

## 设计边界

### challenge list page model 本轮负责

- 保留列表加载、分页、空态、错误态、筛选状态和 query sync 策略
- 不再直接 import `vue-router`
- 暴露 header 与目录行需要的 route target builder

### shared transport 本轮负责

- 只提供当前 query 读取与 `replaceQuery()` transport
- 不承接 challenge-specific parse / normalize / refresh policy

### route view / UI 本轮负责

- `ChallengeList.vue` 直接消费 header route target
- `ChallengeDirectoryRow.vue` 直接消费题目详情 route target

## 任务切片

- [ ] Slice 1：route target 收口
  - 目标：
    - 返回仪表盘、能力画像、题目详情改成 route target contract
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts`

- [ ] Slice 2：query transport 收口
  - 目标：
    - `useChallengeListPage.ts` 去掉 `vue-router`
    - query sync 改成消费共享 query transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / backlog / review 收尾
  - 目标：
    - 更新 allowlist、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeList.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-list-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-list-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-list-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/routeQueryTransport.ts code/frontend/src/features/challenge-list/model/challengeListRoutes.ts code/frontend/src/features/challenge-list/model/index.ts code/frontend/src/features/challenge-list/model/useChallengeListPage.ts code/frontend/src/views/challenges/ChallengeList.vue code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 新的 shared query transport 必须保持足够薄，不能把 challenge-specific parse / normalize 平移出去。
- `ChallengeDirectoryRow.vue` 从 button 改成 route link 后，需要确认交互和选择器测试不依赖旧标签类型。
