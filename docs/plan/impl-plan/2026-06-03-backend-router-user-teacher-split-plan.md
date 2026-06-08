# 2026-06-03 backend router user / teacher split plan

## Objective

- 把 `code/backend/internal/app/router_routes.go` 中的 `registerUserRoutes` 从单个大函数拆成稳定的 registrar 组合。
- 让学生 / 自助侧路由与教师侧路由分别落到独立文件，收口当前 user surface 的 owner 混杂。
- 保持现有 URL、权限、中间件顺序、handler owner 和测试语义不变。

## Non-goals

- 不在本轮改动 `registerTeacherAuthoringRoutes`。
- 不在本轮把 teacher/session/report 逻辑继续下沉到各 module runtime。
- 不在本轮修复既有 runtime / practice 集成测试里的环境性失败。

## Inputs

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-admin-identity-ops-split-review.md`

## Current problem

- `registerUserRoutes` 仍把以下责任堆在同一函数内：
  - 公共赛事浏览与 contest proxy
  - 学生 contest gameplay / challenge practice / instance lifecycle
  - `users/me/*` 自助画像与推荐
  - 教师 overview / class / student insight / AWD review / writeup moderation / manual review
  - personal / class report 出口
- 这已经不是“单个角色的一组路由”，而是多个角色和多个业务域的混合 owner。

## Working design

### Target structure

- 保留 `registerUserRoutes(apiV1, protected, teacherOrAbove, deps)` 作为入口。
- 在 `internal/app` 内新增：
  - `router_user_self_routes.go`
  - `router_user_teacher_routes.go`
- `registerUserRoutes` 只做两次委托：
  - `registerUserSelfRoutes(apiV1, protected, deps)`
  - `registerTeacherRoutes(protected, teacherOrAbove, deps)`

### Boundary choice

- `user self` 文件负责：
  - `/api/v1/contests` 公共浏览
  - `protected` 下的 contest gameplay / challenge practice / instances
  - `/users/me/*`
  - `/reports/personal`
  - `/reports/class` 上的 teacher-guard 版本
- `teacher` 文件负责：
  - `protected /users/:id/skill-profile`
  - `teacherOrAbove` 全部 teacher 路由
  - `/teacher/reports/class`

### Grill check

- 术语与代码一致：这里的 “user self” 指自助 student / logged-in user surface，不等于匿名 public route。
- 不再用“student routes”命名文件，因为里面还包含 teacher 可访问但不绑定 `/teacher/*` 的自助与共享入口。
- `users/:id/skill-profile` 明确归到 teacher 侧 owner，因为它受 `RequireRole(teacher)` 保护，行为上属于教师查看学生画像，而不是用户自助。

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补 implementation plan 与 reuse decision。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-user-teacher-split`

### Slice 2：先补结构测试

- Goal：用结构测试约束 `registerUserRoutes` 中的学生/教师路由必须迁到独立 registrar 文件。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile|TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 3：拆出 user self / teacher registrar

- Goal：把 `registerUserRoutes` 拆成两个独立 registrar 文件，并收窄 deps。
- Touched files：
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/app/router_user_self_routes.go`
  - `code/backend/internal/app/router_user_teacher_routes.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile|TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 4：最小充分回归

- Goal：确认路径矩阵与 teacher access matrix 不漂移。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile|TestTeacherRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
  - `bash scripts/check-consistency.sh`

### Slice 5：review 归档

- Goal：记录本轮 self-review 与独立 reviewer gate 状态。
- Touched files：
  - `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-user-teacher-split-review.md`

## Expected change surface

- `internal/app` user/teacher route registrar 拆分
- 局部 deps struct
- 结构测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径、method、权限变更。
- 风险主要在：
  - `apiV1` 与 `protected` 路由归属切换时漏掉 proxy / report / users 路径
  - `teacherOrAbove` 保护的 teacher routes 中间件顺序漂移

## Review fit check

- Owner 清晰：本轮完成后 `registerUserRoutes` 将不再承载多域混合 owner。
- Reuse 清晰：沿用前三刀已经建立的 registrar 拆分模式。
- 结构收敛：一旦触碰 `registerUserRoutes`，本轮直接完成该 surface 的 owner 收口，而不是只拆一半。

## Rollback / recovery

- 纯代码组织调整，可直接回退新增 registrar 文件与分发调用。
