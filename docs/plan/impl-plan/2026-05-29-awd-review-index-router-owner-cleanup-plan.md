> 状态：Current
> 事实源：AWD review index feature model、teacher/platform route views、前端架构 allowlist
> 替代：无

# AWD Review Index Route Target Cleanup Plan

## 目标

- 把 AWD 复盘目录页剩余的薄导航 owner 从 `useAwdReviewIndexPage.ts` 收口成显式 route target contract。
- 删除 `useAwdReviewIndexPage.ts -> vue-router` 这条 allowlist。

## 非目标

- 不修改 AWD 复盘目录的数据加载、筛选、分页和展示结构。
- 不处理 AWD review export flow、detail page 或其它 AWD allowlist 条目。
- 不改变 teacher / platform 现有的返回入口和详情跳转目标。

## 输入依据

- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useAwdReviewIndex.ts` 已经是 teacher / platform 共享的数据与筛选 owner，本轮不再动它。
- `useAwdReviewIndexPage.ts` 现在只剩薄导航，但继续持有 `useRouter()` 没有必要；这类能力应收口成 route target contract，再由 view/widget 直接消费。
- 这轮收口后，teacher / platform AWD review index 的薄导航会从 feature page wrapper 进一步下沉到显式 route target 文件。

## 设计边界

### `useAwdReviewIndex.ts` 本轮负责

- 数据加载、筛选、分页
- 目录摘要、状态文案、行数据构建
- 错误态、debounce、abort cleanup

### `useAwdReviewIndexPage(scope)` 本轮负责

- 组合 `useAwdReviewIndex()`
- 暴露 `homeRoute`
- 暴露 `buildContestRoute(contestId)`

### `awdReviewIndexRoutes.ts` 本轮负责

- 按 `teacher / platform` scope 解析返回概览 route target
- 按 `teacher / platform` scope 构建 AWD 复盘详情 route target

### `TeacherAWDReviewIndex.vue` / `AWDReviewIndex.vue` 本轮负责

- 继续作为薄 route shell 组合 page wrapper 与展示组件
- 把 route target 传给 teacher workspace 与 platform hero / directory panel

## 任务切片

### Slice 1：page wrapper 去掉 router

- 目标：
  - 删除 `useAwdReviewIndexPage.ts` 里的 `useRouter()` 和导航函数
  - 新增 `awdReviewIndexRoutes.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
- Review focus：
  - page wrapper 是否不再直接持有 route name / router.push

### Slice 2：view / widget 直接消费 route target

- 目标：
  - teacher / platform route view、teacher workspace、platform hero / directory panel 改成 `AppRouteLink`
  - 原有用户可见行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
- Review focus：
  - 返回和详情跳转是否已经经由显式 route target，而不是继续停留在 `router.push`

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 allowlist、backlog 和 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useAwdReviewIndexPage.ts` 是否已经从 feature router allowlist 移除

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-review-index-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-review-index-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-review-index-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts code/frontend/src/features/awd-review-workspace/model/awdReviewIndexRoutes.ts code/frontend/src/features/awd-review-workspace/model/index.ts code/frontend/src/features/awd-review-workspace/index.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.test.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.vue code/frontend/src/widgets/awd-review-workspace/AwdReviewContestDirectory.test.ts code/frontend/src/widgets/awd-review-workspace/AwdReviewContestRow.vue code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue code/frontend/src/views/platform/AWDReviewIndex.vue code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 AWD review index 这一条 route target 债，不继续改 AWD export flow 或 detail 页。
- review 仍默认按同上下文 self-review 记录，独立 reviewer gate 仍未满足。
