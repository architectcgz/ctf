> 状态：Current
> 事实源：awd review detail page model、route transports、前端架构 allowlist
> 替代：无

# AWD Review Detail Route Owner Cleanup Plan

## 目标

- 收掉 `features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts -> vue-router`

## 非目标

- 不改 `api/awd-reviews.ts` 的 role-aware API owner。
- 不重构 export polling、team drawer 或 summary 聚合逻辑。
- 不调整 teacher / platform AWD review detail route 名称解析规则。

## 输入依据

- `useAwdReviewDetailPage.ts`
- `PlatformAwdReviewDetail.test.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `teachingWorkspaceRouting.ts`
- `architectureAllowlist.ts`

## 当前结论

- 这条 route owner 比 `challenge-detail` 更轻，主要是 params/query 读取和返回列表导航。
- `setRound()` 改为 `replaceQuery()` 后，当前页 query owner 仍留在 detail page，本地业务规则不需要继续沉到 shared。

## 设计边界

### awd review detail page model 本轮负责

- 保留详情加载、summary 聚合、team drawer、export polling、breadcrumb 和错误处理 owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `params / query` 读取
- 提供 query 写回与 `push()` 导航
- 不承接 AWD review detail 的 round normalize 或 role 语义

### awd review detail routes 本轮负责

- 描述返回 AWD review index 的 route target
- 不承接 round query 构建、详情加载或 export workflow

## 任务切片

- [ ] Slice 1：抽本地 back route target
  - 目标：
    - 新增 `awdReviewDetailRoutes.ts`
    - 返回列表导航不再在 page model 内直接拼 route object
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`

- [ ] Slice 2：page model 改用 shared route transports
  - 目标：
    - `contestId` / `round` 改用 `routeQueryTransport`
    - `setRound()` 改用 `replaceQuery()`
    - `openReviewIndex()` 改用 `routeNavigationTransport`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-detail-workspace/model/awdReviewDetailRoutes.ts code/frontend/src/features/awd-review-detail-workspace/model/index.ts code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 当前 AWD review index/detail route 名称解析仍依赖 `teachingWorkspaceRouting.ts`；本轮只是把 page model 的 transport owner 收回，不处理 role-route 命名来源本身。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
