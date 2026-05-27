> 状态：Current
> 事实源：`ChallengeWriteupManagePanel`、`challenge-writeup-editor`、`api/admin/authoring.ts`、writeup submissions contract、fronted backlog
> 替代：无

# Challenge Writeup Manage API Owner Alignment Implementation Plan

## 目标

- 让 `ChallengeWriteupManagePanel` 对应的管理 workflow 不再通过 `@/api/teaching` 依赖 teacher 语义投稿目录查询。
- 在不改 UI 行为、不改底层 HTTP contract 的前提下，把平台题解管理面板消费的投稿目录查询收口到 platform/admin challenge authoring owner。
- 继续缩小 backlog 里 `/platform/*` 对 teacher 语义 owner 的残余依赖面。

## 非目标

- 本轮不拆 `ChallengeWriteupManagePanel` 的模板和样式结构。
- 本轮不处理 `teacher-student-analysis`、manual review、community writeup 审核等其他 teacher writeup 流程。
- 本轮不重命名 `TeacherSubmissionWriteupItemData` 等 DTO / contract 名称。
- 本轮不削弱教师查看题解的现有能力边界。

## 输入依据

- `code/frontend/src/components/platform/writeup/ChallengeWriteupManagePanel.vue`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/api/admin/authoring.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ChallengeWriteupManagePanel` 只在 platform challenge detail 里使用，当前最直接的结构问题不在模板层，而是 `useChallengeWriteupManagement` 继续直接调用 `getTeacherWriteupSubmissions`。
- 这与之前 AWD review 的 admin / teacher API owner 残留是同类问题：共享或 platform 侧 feature 已经独立，但仍透过 workflow 间接绑定 teacher 命名函数。
- 教师侧题解查看 / 评阅仍由 `TeacherStudentAnalysis` 与 `useSubmissionReviewFlows` 承接，所以这轮只能收 platform 面板的 owner，不能把能力边界误改成 admin-only。
- 最小安全切片是在 `api/admin/authoring.ts` 增加 platform 语义 wrapper，并把 `useChallengeWriteupManagement` 切过去。

## 任务切片

### Slice 1：补 platform writeup submissions wrapper

- 目标：
  - 在 `api/admin/authoring.ts` 暴露 platform 语义的学员题解投稿目录查询 wrapper。
- 预期改动：
  - `code/frontend/src/api/admin/authoring.ts`
- review focus：
  - 只做薄 wrapper / re-export
  - 不引入新的 HTTP path 或 normalize 分叉

### Slice 2：收口 challenge writeup management workflow owner

- 目标：
  - 让 `useChallengeWriteupManagement` 改从 platform/admin owner 读取学员题解投稿目录。
- 预期改动：
  - `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- review focus：
  - 加载、分页、删除后的空态行为不回归
  - 面板组件本身继续不直连 API owner

### Slice 3：同步测试与事实文档

- 目标：
  - 更新 `ChallengeWriteupManagePanel` 测试断言新的 admin API owner。
  - 在 contract doc / backlog 记录这次 owner 收口。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-challenge-writeup-manage-api-owner-alignment-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按 API owner 切：整体撤销 `api/admin/authoring.ts` 的 wrapper 与 `useChallengeWriteupManagement` 的 import 变更即可，不涉及模板、路由或后端接口回退。

## 残余风险

- `TeacherSubmissionWriteupItemData` 等 DTO 命名仍带 teacher 前缀；本轮只先收口 platform surface 的 API owner。
- 更深层 student analysis / manual review writeup 流程仍可能继续依赖 teacher 命名 owner，需要后续单独切片。
