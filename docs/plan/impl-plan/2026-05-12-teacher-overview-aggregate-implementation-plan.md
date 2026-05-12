# 2026-05-12 teacher overview aggregate implementation plan

## Objective

把 `/academy/overview` 从“默认班级详情页”改成真正的教师教学概览工作区，并为后续教师多班级范围扩展保留稳定的 overview contract。

## Non-goals

- 不在本次任务里重做教师权限模型或数据库 `class_name` 结构
- 不改班级详情页 `/academy/classes/:className` 的完整趋势、复盘、洞察、介入 owner
- 不重写 `internal/teaching/advice` 规则内核

## Source architecture / design docs

- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/frontend/01-architecture-overview.md`

## Ordered task slices

### Slice 1: backend overview contract

- Goal: 新增 `GET /api/v1/teacher/overview`，让 overview 拥有独立 DTO、handler 和 query service 入口
- Touched modules:
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/dto/teacher.go`
  - `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
  - `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- Validation:
  - `go test ./internal/app/... ./internal/module/teaching_readmodel/api/http/...`
- Review focus:
  - route 与 handler contract 是否明确，overview 是否仍错误依赖 `class_name`

### Slice 2: teaching readmodel aggregate query

- Goal: 在 `teaching_readmodel` 内新增 overview scope query，输出 summary / trend / focus classes / focus students / weak dimensions
- Touched modules:
  - `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
  - `code/backend/internal/module/teaching_readmodel/ports/query.go`
  - `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
- Validation:
  - `go test ./internal/module/teaching_readmodel/...`
- Review focus:
  - overview owner 是否与 class detail owner 分离
  - 当前单班权限现状下 contract 是否仍保持 scope 化
  - repository 是否没有把 class detail 文案/规则耦合回 overview

### Slice 3: frontend overview owner split

- Goal: 前端新增 overview 专属 loader/workspace，停止 `TeacherDashboardPage` 直接消费 class detail DTO
- Touched modules:
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/classes.ts`
  - `code/frontend/src/api/teacher/index.ts`
  - `code/frontend/src/features/teacher-dashboard/model/*`
  - `code/frontend/src/views/teacher/TeacherDashboard.vue`
  - `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- Validation:
  - `cd code/frontend && npm run test -- src/views/teacher/__tests__/TeacherDashboard.test.ts src/api/__tests__/teacher.test.ts`
- Review focus:
  - route view 是否继续只做壳
  - page/composable 是否仍把 class detail 逻辑混回 overview
  - overview 页面是否不再直接挂载 class-level panel

### Slice 4: docs and guards

- Goal: 同步 API 文档、专题索引和必要测试断言
- Touched modules:
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/features/专题架构索引.md`
  - tests touched by slices above
- Validation:
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/check-consistency.sh`
- Review focus:
  - 文档事实是否与实际 contract 一致
  - reuse decision 是否覆盖受保护文件

## Expected files / boundaries

- Backend:
  - `internal/module/teaching_readmodel` 新增 overview query capability
  - `dto/teacher.go` 新增 overview DTO
  - `router_routes.go` 新增 teacher overview route
- Frontend:
  - `teacher-dashboard` feature 改为 overview scope owner
  - `TeacherDashboardPage.vue` 改为 overview summary workspace

## Data / API / compatibility impact

- 新增 `GET /api/v1/teacher/overview`
- 旧 `GET /api/v1/teacher/classes/:name/*` 保持不变，继续服务班级详情页
- 当前教师权限仍只落到单个 `class_name`，overview contract 先按可访问 scope 聚合，不把单班限制固化到前端

## Validation commands

- `go test ./internal/app/... ./internal/module/teaching_readmodel/...`
- `cd code/frontend && npm run test -- src/api/__tests__/teacher.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Rollback / recovery

- 若 overview endpoint 或前端 contract 不稳定，可先回退 `/academy/overview` 到旧 loader，但不能保留“overview 直接复用 class detail panel”这一长期边界。
- 若 teaching_readmodel aggregate query 结果不可靠，优先修正 query owner，不改由前端临时并发拼 class detail 接口兜底。
