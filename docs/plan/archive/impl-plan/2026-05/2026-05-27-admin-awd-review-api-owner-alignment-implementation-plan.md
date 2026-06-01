> 状态：Current
> 事实源：`awd-review-workspace`、`awd-review-detail-workspace`、`api/admin/contests.ts`、`api/teacher/*`、architecture review 与 backlog
> 替代：无

# Admin AWD Review API Owner Alignment Implementation Plan

## 目标

- 让共享 AWD review workflow 按角色走对应的 API owner，避免 `/platform/*` 继续通过共享 feature 依赖 `@/api/teaching` 里的 teacher 命名函数。
- 收口 `PlatformAwdReviewIndex` 与 `PlatformAwdReviewDetail` 当前最直接的 admin / teacher API owner 耦合。
- 保持底层 HTTP contract、route view 模板结构和共享 widget 行为不变，只调整 API owner 边界。

## 非目标

- 本轮不处理 `PlatformClassWorkspaceSection` 的 route redirect 命名残留。
- 本轮不处理 `ChallengeWriteupManagePanel`、`student-analysis`、`student-directory` 等其他仍依赖 `@/api/teaching` 的 surface。
- 本轮不重命名 `TeacherAWDReview*` DTO 名称，也不改后端接口路径。
- 本轮不新建平台专属 AWD review feature 或复制 teacher / admin 两套 page workflow。

## 输入依据

- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/teacher/index.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/views/platform/AWDReviewIndex.vue`
- `code/frontend/src/views/platform/PlatformAwdReviewDetail.vue`
- `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`

## 当前结论

- AWD review 的 route view 和 workspace owner 已经中立，但共享 feature 仍直接 import `listTeacherAWDReviews`、`getTeacherAWDReview`、`exportTeacherAWDReviewArchive`、`exportTeacherAWDReviewReport`。
- 这意味着 admin 侧虽然不再挂 teacher route view，仍在共享 workflow 下间接依赖 teacher 语义 API owner。
- 最小安全切片是复用现有 `api/admin/contests.ts` 加一组 AWD review wrapper，并让 `awd-review-workspace` / `awd-review-detail-workspace` 按角色选择 API owner。

## 任务切片

### Slice 1：补 admin AWD review wrapper owner

- 目标：
  - 在现有 `api/admin/contests.ts` 暴露 admin 语义的 AWD review query / export wrapper。
- 预期改动：
  - `code/frontend/src/api/admin/contests.ts`
- review focus：
  - 只做薄 wrapper / re-export
  - 不引入新的请求实现或 contract 分叉

### Slice 2：让共享 AWD review feature 按角色选择 API owner

- 目标：
  - 让 `useAwdReviewIndex`、`useAwdReviewExportFlow`、`useAwdReviewDetailPage` 不再直接绑定 `@/api/teaching` teacher 函数。
- 预期改动：
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
  - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
  - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- review focus：
  - teacher / admin 都继续复用同一条 workflow
  - role-based API owner 选择不漂移到 route view
  - 导出轮询、错误处理、round query 行为不回归

### Slice 3：同步测试与事实文档

- 目标：
  - 让 platform / teacher AWD review 测试断言新的 API owner 选择。
  - 更新 AWD review 架构事实与 backlog 进展。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退粒度按 API owner 切：可以整体回退 `api/admin/contests.ts` 的 wrapper 和 AWD review feature 内的 role-based import 选择，不涉及接口、数据或页面结构回退。

## 残余风险

- `PlatformClassWorkspaceSection` 的 redirect 命名残留、本轮未覆盖的 `ChallengeWriteupManagePanel` 和其他 `@/api/teaching` 依赖仍需后续继续收口。
- `TeacherAWDReview*` DTO / contract 命名依然保留 teacher 前缀；本轮只先把共享 workflow 的 API owner 从 teacher 语义上摘下来。
