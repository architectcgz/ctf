> 状态：Current
> 事实源：`useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 当前 role-aware access owner
> 替代：无

# AWD Review Role-Aware Access Owner Normalization Plan

## 目标

- 为 AWD review 共享 feature 增加中立 role-aware access owner
- 把 `useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 里的 `admin/teacher` API 分支统一收口
- 同步更新 teacher / platform AWD review 相关护栏和 backlog 记录

## 非目标

- 本轮不重命名 `TeacherAWDReview*` DTO、widget props 或 presentation helper
- 本轮不改 AWD review route name resolver
- 本轮不改 `widgets/awd-review-workspace` 的 UI 结构和视觉表现

## 输入依据

- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`

## 当前结论

- AWD review 的 route view 已经是中立 shared feature + role-specific route name 的模式
- 但 role-aware API 选择仍然分散在 3 个 feature model 文件里，各自双引 `@/api/admin` 和 `@/api/teacher`
- `api/admin/contests.ts` 本身只是 alias 到 `api/teaching/awd-reviews.ts`，说明 access owner 还有进一步收口空间

## 设计边界

### `api/awd-reviews.ts` 本轮负责

- list AWD reviews
- get AWD review detail
- export AWD review archive
- export AWD review report
- 按 role 选择 admin / teacher API owner

### AWD review feature model 本轮继续负责

- 路由、筛选、分页、导出 workflow
- loading / error / polling state owner

### 本轮不动

- AWD review route shell
- AWD review widget props 契约
- 教师命名前缀 DTO

## 任务切片

### Slice 1：role-aware access owner 收口

- 目标：
  - 新增 `api/awd-reviews.ts`
  - 统一提供 `listAwdReviewsByRole`、`getAwdReviewByRole`、`exportAwdReviewArchiveByRole`、`exportAwdReviewReportByRole`
  - `useAwdReviewIndex.ts`、`useAwdReviewExportFlow.ts`、`useAwdReviewDetailPage.ts` 改为只依赖这层中立 owner
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Review focus：
  - AWD review feature 内是否不再散落双 API import
  - role-aware access owner 是否集中成单点，而不是换个文件继续复制分支逻辑

### Slice 2：护栏与 backlog 同步

- 目标：
  - 更新 teacher / platform AWD review 测试源码断言
  - backlog 记录 AWD review 这条 P1 结构债的本轮进展
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 护栏是否明确要求 feature 走中立 access owner
  - 不把 DTO 命名和 access owner 收口混在同一刀

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 AWD review 共享 feature 的 access owner，不继续改 `TeacherAWDReview*` 合同命名；如果后续继续推进 admin / teacher 结构耦合，下一刀更适合专门处理 DTO / widget 的 teacher 语义残留
