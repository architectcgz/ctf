# Admin AWD Review API Owner Alignment 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-admin-awd-review-api-owner-alignment-implementation-plan.md`
  - files reviewed：
    - `code/frontend/src/api/admin/contests.ts`
    - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
    - `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
    - `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
    - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
    - `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
    - `docs/contracts/api-contract-v1.md`
    - `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意本轮属于前端 P1 级结构收口，范围限定在 AWD review 共享 workflow 的 admin / teacher API owner alignment。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 本轮没有再复制平台专属 AWD review workflow，而是复用现有 `awd-review-workspace` / `awd-review-detail-workspace`，只把 API owner 选择收回共享 feature，本身就是这类角色共享页面更低风险的实现方式。
- 把 admin wrapper 放进现有 `api/admin/contests.ts` 而不是新增一层 `api/admin/awd-reviews.ts`，避免了多一个 owner 文件和额外的 check 噪音，也更符合 AWD review 本身隶属 contest/AWD 管理面的事实。
- 同步补契约文档是必要的：虽然 HTTP contract 没变，但本仓库的 completion gate 明确把 API owner 变化也视作需要说明的 contract-level 证据。

## Required re-validation

- `npm run test:run -- src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/api/admin/contests.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts docs/contracts/api-contract-v1.md docs/architecture/features/AWD教师复盘归档与报告导出设计.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/admin-awd-review-api-owner-alignment.md docs/plan/impl-plan/2026-05-27-admin-awd-review-api-owner-alignment-implementation-plan.md docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`

## Residual risk

- `PlatformClassWorkspaceSection`、`ChallengeWriteupManagePanel` 以及更深层 `Teacher*` DTO / contract 命名仍然是 backlog 里的残余耦合面，本轮没有覆盖。
- 这次 review 在同一主会话里完成，没有独立子 agent 复核；因此这里只作为同上下文 gate review 证据，不表述为独立审查。

## Touched known-debt status

- 本轮 touched 的已知结构债是 AWD review 共享 workflow 继续直接依赖 `@/api/teaching` 的 teacher 命名函数。
- 在本轮 touched surface 上，这条债务已经完成收口：`PlatformAwdReviewIndex`、`PlatformAwdReviewDetail` 对应的共享 feature 已按角色切到 `api/admin` / `api/teacher` owner；剩余 admin / teacher 结构耦合已明确收敛到其他未 touched surface。
