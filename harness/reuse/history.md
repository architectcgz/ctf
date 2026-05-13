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
