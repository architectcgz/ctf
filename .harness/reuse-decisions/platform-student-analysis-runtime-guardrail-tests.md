# Reuse Decision

## Change type
test / route-owner-guardrail

## Existing code searched
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/reviews/frontend/2026-05-25-student-analysis-page-prop-contract-convergence-review.md`

## Similar implementations found
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Decision
refactor_existing

## Reason
- `PlatformStudentAnalysis.vue` 已经有现成 route owner 与共享 page 组合面，缺的是运行时挂载护栏，不是新的页面结构。
- `PlatformClassStudents.test.ts` 已经证明平台 route view 可以通过挂载测试锁住平台命名空间导航；这次应沿用同类模式，而不是继续扩充字符串断言。
- `TeacherStudentAnalysis.test.ts` 已经覆盖共享 `StudentAnalysisPage` 的真实交互与 bridge 行为；平台侧只需要补平台 owner 独有的桥接保护，不需要再复制整套数据加载测试。

## Files to modify
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/plan/impl-plan/2026-05-25-platform-student-analysis-runtime-guardrail-tests-implementation-plan.md`
- `.harness/reuse-decisions/platform-student-analysis-runtime-guardrail-tests.md`

## After implementation
- 如果后续其它平台 route view 也出现“只有源码断言，没有挂载桥接护栏”的同类问题，可以复用这次的测试模式继续补齐。
