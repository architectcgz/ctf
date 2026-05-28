> 状态：Current
> 事实源：`api/teaching/awd-reviews.ts`、`api/teacher/awd-reviews.ts`、`api/admin/contests.ts`
> 替代：无

# AWD Review API Implementation Owner Neutralization Plan

## 目标

- 把 AWD review 共享实现层 `api/teaching/awd-reviews.ts` 的 teacher 命名函数收口成中性 `AwdReview*` 实现 owner。
- 保持 `api/teacher` public API 不变，并让 `api/admin` 继续通过 platform 语义函数名暴露这组能力。
- 不修改后端 `/api/v1/teacher/awd/reviews*` 路径、权限语义和前端页面行为。

## 非目标

- 本轮不新增 `/api/v1/admin/awd/reviews*` 或其他平台专属 HTTP path。
- 本轮不改 `TeacherAWDReviewIndex` / `TeacherAWDReviewDetail` route name。
- 本轮不改 AWD review feature、widget、DTO 结构或导出交互。

## 输入依据

- `code/frontend/src/api/teaching/awd-reviews.ts`
- `code/frontend/src/api/teacher/awd-reviews.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/awd-reviews.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`
- `docs/reviews/frontend/2026-05-28-awd-review-role-aware-access-owner-normalization-review.md`
- `docs/reviews/frontend/2026-05-28-awd-review-shared-contract-naming-neutralization-review.md`

## 当前结论

- route view、shared feature access owner、shared contract naming 都已经收口过，当前剩余 teacher 语义残片主要停在：
  - `api/teaching/awd-reviews.ts` 的共享实现函数名
  - `api/admin/contests.ts` 对 teacher 函数的 alias re-export
- `api/teacher/awd-reviews.ts` 是更合适的 teacher public owner 落点，因此 teacher 语义应保留在这里，而不是继续渗进 `api/teaching` 共享实现层。

## 设计边界

### 本轮负责

- 中性化 AWD review 共享实现函数名
- 调整 teacher / admin public owner 的导出方式
- 补 raw-source 护栏、contract 记录和 backlog 进展

### 本轮不动

- 后端路由和 HTTP path
- AWD review 页面 / feature / widget 行为
- teacher / platform route name

## 任务切片

### Slice 1：中性化 teaching 层 AWD review 实现符号

- 目标：
  - 在 `api/teaching/awd-reviews.ts` 提供中性实现函数，例如 `listAwdReviews`、`getAwdReview`、`exportAwdReviewArchive`、`exportAwdReviewReport`
  - teacher 命名函数改成兼容桥或薄包装，不再充当共享实现 owner
- 预期改动：
  - `code/frontend/src/api/teaching/awd-reviews.ts`
- review focus：
  - HTTP path、normalize 行为、返回 DTO 完全不变

### Slice 2：收口 admin public owner

- 目标：
  - `api/admin/contests.ts` 改为显式 platform owner，不再 alias teacher 命名函数
- 预期改动：
  - `code/frontend/src/api/admin/contests.ts`
- review focus：
  - admin public owner 不再显式暴露 teacher 命名依赖
  - 现有 feature / test 调用面不回归

### Slice 3：测试与文档收尾

- 目标：
  - 补 API raw-source 护栏，记录 AWD review 本地 API owner 已进一步收口
- 预期改动：
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-awd-review-api-implementation-owner-neutralization-review.md`

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/api/__tests__/admin.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 完成后，AWD review 仍会保留 `/api/v1/teacher/awd/reviews*` 这组后端路径，以及 teacher route name；这是当前明确保留的 transport / route 语义，不属于前端本地共享 owner 漂移。
- `api/teaching` 目录内其他教学域函数是否仍有 teacher 命名残片，不在本轮范围；这刀只收 AWD review touched surface。
