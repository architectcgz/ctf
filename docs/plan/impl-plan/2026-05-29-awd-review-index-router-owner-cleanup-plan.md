> 状态：Current
> 事实源：AWD review index feature model、teacher/platform route views、前端架构 allowlist
> 替代：无

# AWD Review Index Router Owner Cleanup Plan

## 目标

- 把 AWD 复盘目录页的 router owner 从共享 feature model `useAwdReviewIndex.ts` 收回显式的 route-aware page wrapper。
- 删除 `useAwdReviewIndex.ts -> vue-router` 这条 allowlist。

## 非目标

- 不修改 AWD 复盘目录的数据加载、筛选、分页和展示结构。
- 不处理 AWD review export flow、detail page 或其它 AWD allowlist 条目。
- 不改变 teacher / platform 现有的返回入口和详情跳转目标。

## 输入依据

- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useAwdReviewIndex.ts` 是 teacher / platform 共享的数据与筛选 owner，不是单一角色 route owner。
- 返回各自首页和进入详情的导航应落在 route-aware page wrapper，而不是混在共享 feature model 里，也不应直接回退到 `.vue` view。
- 这轮收口后，shared feature model 应只保留数据与 workflow，route shell 通过 page wrapper 接住角色路由。

## 设计边界

### `useAwdReviewIndex.ts` 本轮负责

- 数据加载、筛选、分页
- 目录摘要、状态文案、行数据构建
- 错误态、debounce、abort cleanup

### `useAwdReviewIndexPage(scope)` 本轮负责

- 按 `teacher / platform` scope 持有 router.push
- 返回对应首页
- 打开对应 AWD 复盘详情

### `TeacherAWDReviewIndex.vue` / `AWDReviewIndex.vue` 本轮负责

- 继续作为薄 route shell 组合 page wrapper 与展示组件

## 任务切片

### Slice 1：共享 feature model 去掉 router

- 目标：
  - 删除 `useAwdReviewIndex.ts` 里的 `useRouter()` 和导航函数
  - 保持加载 / 筛选 / 分页行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
- Review focus：
  - feature model 是否不再直接持有 route name / router.push

### Slice 2：page wrapper 承接角色导航

- 目标：
  - 新增 `useAwdReviewIndexPage(scope)` 承接角色导航
  - teacher / platform route view 退回薄壳
  - 原有用户可见行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
- Review focus：
  - 返回和详情跳转 owner 是否已落到对应 route view

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useAwdReviewIndex.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-index-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts code/frontend/src/features/awd-review-workspace/model/index.ts code/frontend/src/features/awd-review-workspace/index.ts code/frontend/src/views/platform/AWDReviewIndex.vue code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 AWD review index 这一条 router owner，不继续改 AWD export flow 或 detail 页。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
