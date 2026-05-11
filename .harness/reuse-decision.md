# Reuse Decision

## Change type
- hook
- page

## Existing code searched
- code/frontend/src/views
- code/frontend/src/components
- code/frontend/src/features
- code/frontend/src/widgets
- code/frontend/src/composables
- code/frontend/src/api

## Similar implementations found
- code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts
- code/frontend/src/api/challenge.ts
- code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts

## Decision
- extend_existing

## Reason
- 这次不是新增下载能力，而是修正已有题目详情页与管理员题目详情页的同源附件下载实现。
- 后端附件接口已经能直接返回正确的 `Content-Disposition` 和 ZIP 字节，因此应复用现有 `/api/v1/challenges/attachments/...` 下载入口，而不是再保留一条 `Axios -> Blob -> objectURL` 的并行链路。
- 修复方式是在现有两个页面模型上直接收口为浏览器原生下载，避免创建新的下载 service 或新的附件页面流程。
- 受影响测试沿用现有详情页测试文件扩充，不新增独立测试 harness。

## Files to modify
- code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts
- code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts
- code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts
