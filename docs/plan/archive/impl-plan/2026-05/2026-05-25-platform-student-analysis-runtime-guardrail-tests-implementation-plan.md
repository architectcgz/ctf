> 状态：Current
> 事实源：`PlatformStudentAnalysis.vue` 当前 route owner 装配面、`PlatformStudentAnalysis.test.ts` 现有护栏、`TeacherStudentAnalysis.test.ts` 既有挂载测试模式
> 替代：无

# Platform Student Analysis Runtime Guardrail Tests Implementation Plan

## 目标

- 给 `PlatformStudentAnalysis.vue` 补运行时挂载级测试，锁住平台 route view 对共享 `StudentAnalysisPage` workflow 的事件桥接。
- 保留现有源码 owner 断言，同时新增能在挂载后实际证明平台路由、导出弹窗和 challenge 跳转仍然接线正确的护栏。
- 尽量只动平台 route view 测试面；如果可测性不足，只做最小实现调整。

## 非目标

- 本轮不改 `TeacherStudentAnalysis.vue`、`StudentAnalysisPage.vue`、`useStudentAnalysisPage.ts` 的运行时逻辑。
- 本轮不重写整套学员分析真实 API 挂载测试，不扩大到 page model 数据加载路径。
- 本轮不收口其它平台 route view 的测试债。

## 输入依据

- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/reviews/frontend/2026-05-25-student-analysis-page-prop-contract-convergence-review.md`

## 当前结论

- `PlatformStudentAnalysis.test.ts` 现在只能证明源码 import / template contract，没有验证挂载后 `StudentAnalysisPage` emits 是否还正确桥接到平台 owner。
- 这条风险已经在前一轮 review 里被明确标出；如果后续 route view 改了事件名、漏传 handler、或平台路由解析退回 teacher 命名空间，现有字符串断言不会报警。
- 该风险最小收口方式是保留静态 owner 断言，并补一组 route-view-level mount tests：共享 page 发出事件后，平台路由与导出对话框行为要能在运行时成立。

## 任务切片

### Slice 1：补平台 route owner 运行时挂载测试

- 目标：
  - 按 TDD 给 `PlatformStudentAnalysis.test.ts` 增加运行时测试。
  - 用 stub `StudentAnalysisPage` 驱动 `retry`、`openClassStudents`、`openReviewArchive`、`openChallenge`、`openReportExport` 等桥接事件。
  - 用受控 mock `useStudentAnalysisPage()` 返回值验证 route view 是否把 handler、dialog 状态与 `selectedClassName` 正确接到模板上。
- 预期改动：
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - 如断言或挂载暴露可测性缺口，再决定是否需要最小实现调整
- Review focus：
  - 事件桥接是否真正发生在平台 route view，而不是只验证 stub 自己。
  - 是否锁住平台命名空间跳转，而不是笼统验证 “调用了 push”。

### Slice 2：如有必要，做最小可测性修正并回归

- 目标：
  - 只有当 Slice 1 暴露当前 route view 无法被稳定验证时，才做最小实现调整。
  - 保持 page model owner、共享 page contract 和用户可见行为不变。
- 预期改动：
  - 优先只动 `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - 如确有必要，再动 `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- 验证：
  - `git diff --check -- code/frontend/src/views/platform/PlatformStudentAnalysis.vue code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts docs/plan/impl-plan/2026-05-25-platform-student-analysis-runtime-guardrail-tests-implementation-plan.md .harness/reuse-decisions/platform-student-analysis-runtime-guardrail-tests.md`
  - `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Review focus：
  - 是否只增加 guardrail，没有把测试债修补扩大成实现层重构。
  - 挂载测试是否覆盖了 review 指出的残余风险，而不只是换一种形式继续读源码字符串。

## 风险

- 如果完全依赖真实 `useStudentAnalysisPage()`，平台 route view 测试会被 API mock 和共享 page 内部渲染细节放大，导致护栏不聚焦。
- 如果只 mock composable、却不验证 emits 与 props 在模板上的实际接线，测试仍可能遗漏 route owner 桥接错误。
- 如果为测试方便改动 route view 结构过多，会把“补护栏”变成不必要的实现噪音。

## 回退方式

- 本轮只涉及前端 route view 测试与可能的极小可测性调整；如有问题，可逐文件回退 `PlatformStudentAnalysis.test.ts` 与对应 route view 变更。
