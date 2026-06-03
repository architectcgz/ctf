# 2026-06-03 backend router admin contest / awd split plan

## Objective

- 把 `code/backend/internal/app/router_routes.go` 中 `registerAdminRoutes` 里的 contest / AWD 管理路由拆到独立 registrar 文件。
- 让 `internal/app` 的路由注册按访问域和业务域收口，避免继续把 admin contest、AWD、identity、ops 全堆在一个函数里。
- 在不改变现有路由、权限、中间件顺序和 handler owner 的前提下，为后续继续拆 `admin identity / ops`、`user`、`teacher` 路由提供稳定落点。

## Non-goals

- 不在本轮把整个 `router_routes.go` 一次性拆完。
- 不在本轮修改 URL、权限模型、审计语义或 handler 行为。
- 不把 route registration 直接下沉进 `internal/module/*/runtime`。
- 不顺手重构 contest / practice / assessment 的 handler 或 service。

## Inputs

- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/router_session_routes_test.go`
- `docs/requirements/opening-report-gap-analysis-2026-04-01.md`

## Current problem

- `router_routes.go` 当前有 `1065` 行，`registerAdminRoutes` 自身覆盖 admin ops、identity、contest、AWD、session revoke 等多类责任。
- admin contest / AWD 路由是其中最密集的一段，重复出现 `ParseInt64Param("id")`、`contest/:id/awd/*`、审计 builder 和相同 deps 访问模式。
- 继续在同一函数上叠加新路由，会让 touched surface 同时承载结构债和功能改动，后续 review 与回归验证成本都会继续变差。

## Working design

### Target structure for this slice

- 保留 `router.go` 作为 composition root。
- 保留 `registerAdminRoutes(adminOnly, deps)` 作为 admin 入口。
- 在 `internal/app` 内新增独立 registrar 文件，至少拆出：
  - `registerAdminContestRoutes`
  - `registerAdminContestAWDRoutes`
- `registerAdminRoutes` 只保留 admin 入口级分发，以及本轮未拆部分。

### Dependency boundary

- 不新增新的跨模块抽象。
- 通过更窄的 deps struct 收口 registrar 依赖，避免 contest / AWD registrar 继续直接依赖完整 `adminRouteDeps`。
- 路由层继续只负责：
  - 路径声明
  - router group 组织
  - 参数解析中间件
  - audit 中间件挂载
  - 调用既有 handler

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补齐当前切片的 implementation plan 和 reuse decision，明确 touched surface 与复用结论。
- Touched files：
  - `docs/plan/impl-plan/2026-06-03-backend-router-admin-contest-awd-split-plan.md`
  - `.harness/reuse-decisions/backend-router-admin-contest-awd-split.md`
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-admin-contest-awd-split`
- Review focus：
  - 切片边界是否足够小且不遗漏结构债 owner

### Slice 2：先补 registrar 级测试

- Goal：先为 admin contest / AWD registrar 的存在和路由落点加测试，确认拆分前后外部行为不变。
- Touched files：
  - `code/backend/internal/app/router_test.go`
  - 如有必要新增 admin contest registrar 单测文件
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
- Review focus：
  - 断言是否覆盖本轮拆分的主要 admin contest / AWD 路由

### Slice 3：抽出 admin contest / AWD registrar

- Goal：把 `registerAdminRoutes` 里的 contest / AWD 注册逻辑迁到独立文件，并收窄 deps。
- Touched files：
  - `code/backend/internal/app/router_routes.go`
  - 新增 `code/backend/internal/app/router_admin_contest_routes.go`
- Implementation notes：
  - 保持现有 route path 与 handler 不变。
  - contest 基础路由与 contest 下的 AWD 子路由可以在同一文件中拆成两个函数。
  - `awdReadinessAudit` 如仍只服务 admin contest 相关路径，可在 registrar 内透传 closure，不新增全局 helper。
- Validation：
  - `cd code/backend && go test ./internal/app -count=1`
- Review focus：
  - 中间件顺序、audit 绑定和 `ParseInt64Param` 顺序不能漂移

### Slice 4：独立 review 与回归验证

- Goal：在拆分完成后做一次 backend review，确认没有把结构债换成更多间接层。
- Touched files：
  - `docs/reviews/backend/2026-06-03-backend-router-admin-contest-awd-split-review.md`
- Validation：
  - `cd code/backend && go test ./internal/app -count=1`
- Review focus：
  - registrar owner 是否清晰
  - deps 是否真的变窄
  - 是否仍有本轮 touched surface 内必须一并收口的遗漏

## Expected change surface

- `internal/app` 路由注册文件拆分
- admin contest / AWD registrar 函数与局部 deps struct
- `internal/app` 路由测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径、HTTP method、权限语义变更。
- 无 handler contract 变更。
- 风险主要在中间件顺序、局部漏注册或 audit 绑定漂移。

## Validation matrix

- admin contest 基础路由仍存在并指向原 handler。
- admin AWD 路由仍存在并指向原 handler。
- runtime / practice 相关 admin AWD instance orchestration 路由不丢失。
- `internal/app` 全路由测试仍通过。

## Review fit check

- Owner 清晰：本轮只收口 `internal/app` 路由注册 owner，不扩大到 module runtime。
- Reuse 清晰：继续复用现有 handler、audit helper、middleware，不新造路由 DSL。
- 结构收敛：本轮 touched surface 直接消化 `registerAdminRoutes` 的一块 oversized owner，不把同一位置的结构债继续留到 follow-up。

## Rollback / recovery

- 这次是纯代码组织调整，无持久化副作用。
- 如拆分后测试或 review 发现中间件顺序漂移，可直接回退新增 registrar 文件并恢复 `registerAdminRoutes` 原有内联注册逻辑。
