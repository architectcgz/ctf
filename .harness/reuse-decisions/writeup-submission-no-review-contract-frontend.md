# Reuse Decision

## Change type

- page
- component
- hook
- api
- schema

## Existing code searched

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPresentation.ts`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1/`

## Similar implementations found

- `code/frontend/src/features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`
- `code/frontend/src/api/challenge.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `code/frontend/src/components/teacher/review-archive/reviewArchiveCases.ts`

## Decision

extend_existing

## Reason

这次不是新增一条题解审核流程，而是把现有题解提交/展示/教师列表的状态语义收口到后端已经保留的 `draft / published + visibility_status + published_at` 事实。

最小可复用路径是延续现有的：

- challenge detail 的题解保存/发布 flow
- teacher writeup 列表与复盘展示
- OpenAPI source/bundle 同步链

只把前端状态映射和契约字段对齐到当前后端模型，去掉 writeup 专属的 `submitted` / `review_status` 分支，不引入新的 workflow owner。

## Files to modify

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPresentation.ts`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1/components/schemas/teacher.yaml`
- `docs/contracts/openapi-v1/paths/teacher.yaml`
- `docs/contracts/openapi-v1.yaml`

## After implementation

- 后续同类改动应继续复用现有 challenge-detail 与 teacher writeup 展示链路。
- 如果 writeup 领域再次出现“前端审核态”文案，优先检查契约是否又漂回旧模型，而不是新建第二套状态。
