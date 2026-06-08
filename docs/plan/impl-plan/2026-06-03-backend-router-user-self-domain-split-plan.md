# 2026-06-03 backend router user self domain split plan

## Objective

- 把 `code/backend/internal/app/router_user_self_routes.go` 中仍然混在一起的 user self route surface 继续按行为 owner 拆成稳定 registrar。
- 让 contest participation、practice/runtime、自助画像与报表各自落到独立文件，收口当前 user self surface 的多域混杂。
- 保持现有 URL、权限、中间件顺序、handler owner 和测试语义不变。

## Non-goals

- 不在本轮修改 `registerUserRoutes`、`registerTeacherRoutes`、`registerTeacherAuthoringRoutes` 或已拆好的 admin registrars。
- 不在本轮调整 `/api/v1/teacher/*`、`/api/v1/authoring/*` 或 admin 路由。
- 不在本轮修复既有 runtime / practice 集成测试里的环境性失败。

## Inputs

- `code/backend/internal/app/router_user_self_routes.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-authoring-split-review.md`

## Current problem

- `registerUserSelfRoutes` 当前仍同时承载：
  - public / protected contest browse、register、team、AWD participation、contest 内实例
  - 公开题练习、题解投稿、challenge instance、全局 instance lifecycle、排行榜
  - `users/me/*`
  - personal / class report 出口
- 虽然它已经从总入口移出，但从行为 owner 看，contest、practice/runtime、自助画像与报表不是同一块责任。

## Working design

### Target structure

- 保留 `registerUserSelfRoutes(apiV1, protected, deps)` 作为 user self 总入口。
- 在 `internal/app` 内新增：
  - `router_user_contest_routes.go`
  - `router_user_practice_routes.go`
  - `router_user_self_service_routes.go`
- `registerUserSelfRoutes` 只做三次委托：
  - `registerUserContestRoutes(apiV1, protected, deps)`
  - `registerUserPracticeRoutes(apiV1, protected, deps)`
  - `registerUserSelfServiceRoutes(protected, deps)`

### Boundary choice

- `user contest` 文件负责：
  - `/contests` public browse
  - contest register、challenge list、my-progress、AWD workspace / submission / target access / defense ssh
  - contest teams
  - contest 内 challenge / AWD instance 创建与重启
- `user practice` 文件负责：
  - `/challenges/*`
  - challenge writeup submission
  - standalone challenge instance / submit / ranking
  - `/instances/*` 全局实例生命周期与 proxy
- `user self service` 文件负责：
  - `/users/me/*`
  - `/reports/personal`
  - `/reports/:id`
  - `protected /reports/class` 上的 teacher-guard 版本

### Grill check

- 不按 handler 所属 module 切分，因为 contest prefix 下会调用 contest、practice、instance 多个 module；按模块切反而会把一个用户链路打散。
- contest 内 instance 虽然落到 practice / runtime handler，但从用户视角仍是 contest participation owner，应该跟 contest surface 保持一起。
- `/reports/class` 的 protected + teacher guard 版本继续留在 self service，而不是并回 teacher 文件，因为它仍属于用户侧通用报表出口，只是有显式角色约束。

## Task slices

### Slice 1：计划与 reuse-first 门禁

- Goal：补 implementation plan 与 reuse decision。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-router-user-self-domain-split`

### Slice 2：先补结构测试

- Goal：用结构测试约束 `registerUserSelfRoutes` 中的 contest / practice / self service 路由必须迁到独立 registrar 文件。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 3：拆出 user self 子 registrars

- Goal：把 `registerUserSelfRoutes` 拆成 contest / practice / self service 三个独立 registrar 文件，并收窄 deps。
- Touched files：
  - `code/backend/internal/app/router_user_self_routes.go`
  - `code/backend/internal/app/router_user_contest_routes.go`
  - `code/backend/internal/app/router_user_practice_routes.go`
  - `code/backend/internal/app/router_user_self_service_routes.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`

### Slice 4：最小充分回归

- Goal：确认 user self 路径存在性、runtime handler 接线和访问矩阵不漂移。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestFullRouter_AccessControlMatrix|TestUserContestRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserPracticeRoutesAreExtractedIntoDedicatedRegistrarFile|TestUserSelfServiceRoutesAreExtractedIntoDedicatedRegistrarFile' -count=1`
  - `bash scripts/check-consistency.sh`

### Slice 5：review 归档

- Goal：记录本轮 self-review 与独立 reviewer gate 状态。
- Touched files：
  - `docs/reviews/backend/archive/2026-06/2026-06-03-backend-router-user-self-domain-split-review.md`

## Expected change surface

- `internal/app` user self route registrar 继续拆分
- 局部 deps struct
- 结构测试
- review 记录

## Data / API / compatibility impact

- 无数据结构变更。
- 无 API 路径、method、权限变更。
- 风险主要在：
  - contest 内 instance / AWD proxy 误落到 practice owner
  - `/reports/class` 的 protected + teacher guard 版本被错误删掉或错误迁到 teacher registrar
  - `/instances/*` 和 `/contests/*` 的 runtime handler 接线漂移

## Review fit check

- Owner 清晰：本轮完成后 `registerUserSelfRoutes` 不再承载 contest、practice/runtime、自助画像与报表的具体注册细节。
- Reuse 清晰：沿用前几刀已经建立的 registrar 拆分模式。
- 结构收敛：一旦触碰 `registerUserSelfRoutes`，本轮直接完成该 surface 的 owner 收口，而不是只抽离 `/users/me/*` 之类的小块。

## Rollback / recovery

- 纯代码组织调整，可直接回退新增 registrar 文件与分发调用。
