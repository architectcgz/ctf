# Reuse Decision

## Change type

service / component / test / plan

## Existing code searched

- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveObservationStrip.vue`
- `code/frontend/src/widgets/teacher-review-archive/model/presentation.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `docs/architecture/features/教学复盘建议生成架构.md`

## Similar implementations found

- `code/backend/internal/teaching/advice/advice.go`
  - 已经是个人复盘观察、班级复盘建议和推荐理由的共享规则 owner
- `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 已经负责装配 review archive 事实快照并调用共享规则层
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveObservationStrip.vue`
  - 已经是教师复盘归档观察卡片的唯一展示组件
- `code/frontend/src/widgets/teacher-review-archive/model/presentation.ts`
  - 已经是复盘归档摘要区文案出口

## Decision

extend_existing

## Reason

这次不是新增新的教学复盘模块，而是在现有复盘归档链路中把“事实快照 -> 复杂观察项”收口成“三层复盘结构”。最小正确改法是继续扩展现有 `internal/teaching/advice` 规则 owner、继续复用 `report_service` 的事实装配职责，并在既有 review archive 前端组件上同步展示口径。这样可以保持后端 owner 和前端展示入口不分叉，也避免再起一套并行复盘实现。

## Files to modify

- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveObservationStrip.vue`
- `code/frontend/src/widgets/teacher-review-archive/model/presentation.ts`
- `code/frontend/src/widgets/teacher-review-archive/model/presentation.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `docs/plan/impl-plan/2026-05-21-review-archive-three-layer-restructure-implementation-plan.md`

## After implementation

- 后续 review archive 规则调整应继续复用 `internal/teaching/advice` 与 `report_service` 这条链，不再在前端或其他 service 中复制一套观察逻辑。
- 如果复盘归档页继续调整文案或结构，优先复用 `teacher-review-archive` 组件和 presentation model，不再增加新的页面级文案出口。
