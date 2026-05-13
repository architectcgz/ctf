# Teaching Readmodel Student Review Query Surface Phase 4 Slice 4 Implementation Plan

## Objective

继续完成 phase 4，把 `teaching_readmodel` 剩余宽 query surface 里的学生复盘查询收成独立 owner：

- 把 `GetStudentProgress`、`GetStudentRecommendations`、`GetStudentTimeline`、`GetStudentEvidence`、`GetStudentAttackSessions` 从剩余 `application/queries.Service` 拆到独立 `StudentReviewService`
- 让 `api/http.Handler` 不再通过同一个剩余宽接口承接 `/teacher/students/:id/*`
- 保持现有教师学生复盘相关路径、响应结构和权限口径不变

## Non-goals

- 不重写 `teaching_readmodel/infrastructure.Repository` 的 SQL
- 不新增新的 readmodel 模块、路由或 DTO
- 不改 `overview`、`class insight` 已完成的 owner
- 不进入 phase 5，不处理 Redis / GORM concrete allowlist

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/design/教学复盘建议优化方案.md`
- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,service.go,class_insight_service.go,overview_service.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`

## Current Baseline

- `GetOverview` 和 `GetClassSummary/GetClassTrend/GetClassReview` 已经拆出独立 service。
- 当前 `application/queries.Service` 还同时承接目录查询与学生复盘查询。
- `service.go` 里 `getAccessibleStudent`、evidence / attack session helper 和推荐题映射仍与目录查询混在同一物理文件。

## Chosen Direction

1. 保留 `teaching_readmodel` 作为物理模块，不新建模块或 HTTP contract。
2. 在 `application/queries` 内新增独立 `StudentReviewService` 与 `StudentReviewQueryService`，只负责学生复盘相关查询。
3. 现有 `Service` 接口收窄为目录查询：`ListClasses`、`ListStudents`、`ListClassStudents`。
4. `api/http.Handler` 改为分别依赖 `Service`、`OverviewService`、`ClassInsightService`、`StudentReviewService`。
5. `runtime/module.go` 在 wiring 边缘分别构造四类 query owner，并同步更新 guardrail / 集成测试 / 设计事实。

## Ownership Boundary

- `StudentReviewService`
  - 负责：学生进度、推荐题、时间线、证据链、攻击会话、学生访问校验后的聚合拼装
  - 不负责：目录分页与排序、overview 聚合、班级详情复盘
- `application/queries.Service`
  - 负责：班级目录、学生目录、班级成员列表
  - 不负责：继续持有学生复盘 query owner
- `api/http.Handler`
  - 负责：把 `/teacher/students/:id/*` 绑定到 `StudentReviewService`
  - 不负责：通过剩余目录 query service 间接进入学生复盘链路

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-teaching-readmodel-student-review-query-surface-phase4-slice4-implementation-plan.md`
- Add: `.harness/reuse-decisions/teaching-readmodel-student-review-query-surface-phase4-slice4.md`
- Add: `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service.go`
- Add: `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- Modify: `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/cmd/seed-teaching-review-data/main.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/architecture/features/教师教学概览聚合架构.md`

## Task Slices

### Slice 1: 抽出 student review owner

目标：

- `GetStudentProgress`、`GetStudentRecommendations`、`GetStudentTimeline`、`GetStudentEvidence`、`GetStudentAttackSessions` 与相关 helper 从 `service.go` 移到独立 `student_review_service.go`
- `contracts.go` 单独声明 `StudentReviewService`
- shared access helper 保持单点 owner，不在目录 service 与 student review service 间复制

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/application/queries -run 'StudentReview|ClassInsight|Overview' -count=1 -timeout 120s`

Review focus:

- 学生访问校验是否保持原口径
- evidence / attack session 聚合 helper 是否完整迁移，而不是残留半套逻辑在 `service.go`

### Slice 2: 收窄 handler、runtime 与调用点

目标：

- `api/http.Handler` 分别依赖四类 query owner
- `runtime/module.go` 不再把学生复盘挂在剩余目录 query service 上
- 直接 new handler 的集成测试、seed 命令与 guardrail 同步新依赖

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/... ./internal/app/... ./cmd/seed-teaching-review-data/... -count=1 -timeout 120s`

Review focus:

- `/teacher/students/:id/*` 是否都走独立 student review owner
- runtime wiring 与手工 new handler 的调用点是否没有偷留旧宽 surface

### Slice 3: 回收 phase 4 当前事实

目标：

- phase 4 当前状态更新到 slice 4
- API 与 feature 文档同步指向新的 student review owner

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否准确描述“phase 4 已把 teacher readmodel query surface 拆成目录 / overview / class insight / student review”

## Risks

- 如果 `getAccessibleStudent` 迁移不完整，教师与管理员对学生复盘的访问控制可能回归
- 如果 handler 仍通过剩余 `Service` 间接调用学生复盘，这刀只会改文件名不会真正收口
- 如果 evidence / attack session helper 仍散落在 `service.go`，phase 4 会继续留下同一块宽 surface

## Verification Plan

1. `cd code/backend && go test ./internal/module/teaching_readmodel/... ./internal/app/... ./cmd/seed-teaching-review-data/... -count=1 -timeout 120s`
2. `python3 scripts/check-docs-consistency.py`
3. `bash scripts/check-consistency.sh`

## Architecture-Fit Evaluation

- owner 明确：学生复盘查询有独立 query owner，不再挂在剩余目录 `Service` 上
- reuse point 明确：复用现有 repository、DTO、路由、recommendation provider 和访问校验 helper，不新增模块或并行 API
- 这刀同时解决行为与结构：外部 HTTP contract 不变，但 query owner、handler 依赖和 runtime wiring 一起收口
