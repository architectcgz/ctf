> 状态：Current
> 事实源：`api/contracts.ts`、AWD review shared widget / feature / api owner
> 替代：无

# AWD Review Shared Contract Naming Neutralization Plan

## 目标

- 把 AWD review 已经落在 shared owner 的 response DTO 命名从 `TeacherAWDReview*` 收口成中性 `AwdReview*`
- 同步收口 shared widget 里的 `TeacherAwdReview*` presentation helper / copy owner 命名
- 保持 teacher / platform AWD review 的请求路径、路由名、筛选分页和导出行为不变

## 非目标

- 本轮不改 `TeacherAWDReviewIndex`、`TeacherAWDReviewDetail` 这类 teacher route / view 名称
- 本轮不改 `listTeacherAWDReviews()`、`getTeacherAWDReview()`、`exportTeacherAWDReviewArchive()`、`exportTeacherAWDReviewReport()` 这些 teacher endpoint function 名
- 本轮不改 AWD review 的 HTTP path、query 参数、summary 结构或 UI 文案含义

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/api/awd-reviews.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/model/presentation.ts`
- `code/frontend/src/widgets/awd-review-workspace/model/presentation.test.ts`
- `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- AWD review 更深层的 teacher 残片已经不在 route / feature owner，而在共享 response DTO、API page data 和 shared widget presentation 命名
- `AwdReviewContestItemData` 已经证明这条命名线可以按共享 contract 中性化推进，不需要把 teacher endpoint / route 名一并迁移
- `AwdReviewWorkspace`、`AwdReviewIndexWorkspace` 这类 shared widget 继续依赖 `TeacherAWDReviewArchiveData`、`TeacherAwdReviewSummaryStats`、`TEACHER_AWD_REVIEW_WORKSPACE_COPY`，owner 语义已经漂移

## 设计边界

### 本轮负责

- `api/contracts.ts` 里的共享 AWD review response DTO 命名
- `api/teacher/awd-reviews.ts`、`api/teaching/awd-reviews.ts` 的 shared DTO / page data / normalize helper 命名
- `api/awd-reviews.ts`、detail feature model、shared widget、team drawer 的消费面
- AWD review 相关 raw-source / API 护栏和 backlog / review 记录

### 本轮不动

- teacher / platform 路由名
- teacher endpoint function 名
- 请求参数、分页、导出、轮询 owner
- AWD review 的展示结构和 UI 交互

## TDD 立场

- `TDD`：这是共享 contract / widget owner 命名重构，先用 raw-source / API 护栏制造红灯，再做实现迁移

## 任务切片

### Slice 1：共享 AWD review contract 命名收口

- 目标：
  - `TeacherAWDReviewScopeData`、`TeacherAWDReviewOverviewData`、`TeacherAWDReviewRoundItemData`、`TeacherAWDReviewTeamItemData`、`TeacherAWDReviewServiceItemData`、`TeacherAWDReviewAttackItemData`、`TeacherAWDReviewTrafficItemData`、`TeacherAWDReviewSelectedRoundData`、`TeacherAWDReviewArchiveData` 改成中性 `AwdReview*`
  - `TeacherAWDReviewContestPageData` 改成 `AwdReviewContestPageData`
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/awd-reviews.ts`
  - `code/frontend/src/api/teaching/awd-reviews.ts`
  - `code/frontend/src/api/awd-reviews.ts`
- review focus：
  - DTO 结构、字段名、normalize 行为、请求路径保持不变

### Slice 2：shared widget / feature presentation 命名收口

- 目标：
  - `TeacherAwdReviewSummaryStats`、`TeacherAwdReviewSummaryItem`、`TeacherAwdReviewIndexSummaryStats`
  - `buildTeacherAwdReviewSummaryItems`、`buildTeacherAwdReviewIndexSummaryItems`
  - `TEACHER_AWD_REVIEW_WORKSPACE_COPY`、`TEACHER_AWD_REVIEW_INDEX_WORKSPACE_COPY`
  - 全部改为中性 `AwdReview*`
- 预期改动：
  - `code/frontend/src/widgets/awd-review-workspace/model/presentation.ts`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue`
  - `code/frontend/src/widgets/awd-review-workspace/model/presentation.test.ts`
- review focus：
  - shared widget 不再保留 teacher owner 命名，但 copy 内容和交互不回归

### Slice 3：feature / component / test 消费面同步

- 目标：
  - detail feature model、team drawer、workspace test、teacher detail/index 护栏同步改为中性命名
- 预期改动：
  - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
  - `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
  - `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - `code/frontend/src/api/__tests__/teacher.test.ts`
- review focus：
  - shared feature / widget 的 raw-source 护栏是否明确要求中性 contract owner

### Slice 4：backlog 与 review 收尾

- 目标：
  - 记录 AWD review 共享 contract / shared widget naming 已进一步收口
- 预期改动：
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-awd-review-shared-contract-naming-neutralization-review.md`

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/widgets/awd-review-workspace/model/presentation.test.ts src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts src/views/teacher/__tests__/teacherAwdReviewIndexWorkspaceExtraction.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 本轮如果只收 contract 命名、不收 shared widget presentation 命名，touched surface 会继续留半截 owner 漂移，因此 review 默认不接受半收口状态
- teacher endpoint function 名和 route 名仍然保留 teacher 语义，这是本轮刻意保留的 transport / route owner，不视为共享 contract 漂移
