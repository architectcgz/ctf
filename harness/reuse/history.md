# Reuse History

This file is append-only. Keep current task evidence in `.harness/reuse-decisions/`, then append a short durable summary here when the task creates a reusable decision.

## 2026-05-11 challenge attachment native download

### Change type
- hook
- page

### Similar implementations found
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts`
- `code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts`
- `code/frontend/src/api/challenge.ts`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`

### Decision
- `extend_existing`

### Reuse note
- 题目详情页和管理员题目详情页应复用已有附件下载入口。
- 后端附件接口已经返回正确 `Content-Disposition` 和 ZIP 字节，前端不应再保留并行的 `Axios -> Blob -> objectURL` 下载链路。

## 2026-05-14 openapi source and bundle split

### Change type
- schema
- docs
- script

### Similar implementations found
- `docs/contracts/openapi-v1.yaml`
- `scripts/check-code-changes.sh`
- `scripts/check-consistency.sh`

### Decision
- `extend_existing`

### Reuse note
- `docs/contracts/openapi-v1.yaml` 保持稳定 bundle 路径，不直接迁走或改名。
- 需要提升可维护性时，在同级新增拆分源目录，由脚本单向生成 bundle，而不是让消费方理解外部 `$ref`。
- 后续 OpenAPI 变更应先改 `docs/contracts/openapi-v1/`，再运行 `python3 scripts/sync_openapi_from_contract.py`。

## 2026-05-15 writeup submission no-review contract frontend

### Change type
- page
- component
- hook
- api
- schema

### Similar implementations found
- `code/frontend/src/features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPresentation.ts`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`

### Decision
- `extend_existing`

### Reuse note
- 学生题解提交链路继续复用现有 challenge-detail flow，教师题解展示继续复用现有 teacher writeup 列表和复盘面板。
- 题解提交状态应只保留 `draft / published`，契约与前端都不再暴露 writeup 专属审核态。
