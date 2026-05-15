# 题解提交取消审核语义实施计划

> 状态：Draft
> 输入：`docs/contracts/api-contract-v1.md`、`docs/contracts/openapi-v1/`、`docs/contracts/openapi-v1.yaml`、`code/frontend/src/api/contracts.ts`、`code/frontend/src/features/challenge-detail/**`、`code/frontend/src/components/challenge/**`、`code/frontend/src/components/teacher/**`

## 1. 目标

- 把学生题解提交链路收口为 `draft / published` 两态，不再向前端暴露“提交后等待教师审核”的语义。
- 让前端展示、类型定义和契约文档与当前后端事实一致。
- 保留 `manual_review` 题目审核链路不动。

## 2. 非目标

- 不改题目 Flag 的人工审核链路。
- 不新增新的题解审核状态、事件或通知。
- 不改后端实现，只同步前端与契约事实源。

## 3. 设计结论

- 选择方案：把 writeup submission 的状态空间直接收口到 `draft / published`，删除 `submitted` 和 writeup 专属 `review_status`。
- 取舍：前端会少一个分支，但契约和页面文案会和后端当前行为保持一致，避免学生侧继续看到不存在的审核态。

## 4. 影响范围

- 契约源：
  - `docs/contracts/api-contract-v1.md`
  - `docs/contracts/openapi-v1/components/schemas/teacher.yaml`
  - `docs/contracts/openapi-v1/paths/challenges.yaml`
  - `docs/contracts/openapi-v1/paths/teacher.yaml`
  - `docs/contracts/openapi-v1.yaml`
- 前端：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/challenge.ts`
  - `code/frontend/src/api/teacher/writeups.ts`
  - `code/frontend/src/features/challenge-detail/model/useChallengeDetailPresentation.ts`
  - `code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts`
  - `code/frontend/src/features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`
  - `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
  - `code/frontend/src/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue`
  - 相关测试断言

## 5. 切片

### Slice 1：契约源收口

- 目标：从 writeup contract 中移除 `submitted`、`SubmissionWriteupReviewStatus`、writeup review endpoint 和相关 query 条件。
- 验证：
  - `python3 scripts/sync_openapi_from_contract.py`
  - `python3 scripts/sync_openapi_from_contract.py --check`
- review focus：
  - openapi source 和 bundle 是否一致
  - teacher writeup 路径是否只保留列表、详情和 moderation 相关能力

### Slice 2：前端类型与展示收口

- 目标：更新 writeup DTO、状态文案和教师/学生页面的展示逻辑，彻底去掉 `submitted` 分支。
- 验证：
  - 前端定向测试或 typecheck
- review focus：
  - 是否还残留任何“等待教师审核”的可见文案
  - 是否有组件继续依赖已删除的 writeup review 字段

### Slice 3：测试与回归确认

- 目标：补齐或更新受影响测试，确认题解提交、题解展示和教师列表不会再走审核态。
- 验证：
  - 运行受影响的前端测试文件
- review focus：
  - 断言是否覆盖 draft / published 两态
  - 是否误伤 `manual_review` 链路

## 6. 风险与回退

- 风险：教师侧 writeup 展示和契约类型耦合较紧，若漏改一个 `submitted` 分支，会出现类型错误或旧文案残留。
- 风险：OpenAPI bundle 如果不通过脚本同步，契约源和稳定文件会漂移。
- 回退：如果前端收口后出现兼容问题，可先恢复 `submitted` 兼容分支，但不恢复 writeup review endpoint 的契约入口。

## 7. 完成清单

- [x] 收口 writeup 契约状态与教师侧查询字段
- [x] 同步前端题解展示与状态文案，移除等待审核语义
- [x] 运行 `python3 scripts/sync_openapi_from_contract.py --check`
- [x] 运行 harness 一致性检查并确认 reuse decision 已同步
