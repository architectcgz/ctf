# Runtime / Instance 边界阶段 2 Slice 3 Implementation Plan

## Objective

开始把实例 owner 真正从底层 `runtime` 里挪出来，但把这轮控制在 application / contract 迁移，不直接做大搬家：

- 新增 `internal/module/instance` 作为实例 owner 的底层落点
- 把 `instance`、`proxy ticket`、`scheduler / maintenance` 相关 use case 从 `internal/module/runtime/application/*` 迁到 `internal/module/instance/application/*`
- 把这些 use case 依赖的实例 contract 从 `runtime/ports` 收口到 `instance/ports`
- 保留 `runtime` 侧兼容 facade，避免这一轮扩大到 handler、container adapter、practice provisioning 和运行时底层实现的全面迁移

## Non-goals

- 不在这一轮搬动 `internal/module/runtime/api/http/*`
- 不在这一轮搬动 `internal/module/runtime/infrastructure/*`
- 不重命名或删除 `internal/module/runtime/runtime/module.go`
- 不把 `ProvisioningService`、`RuntimeCleanupService`、container files、image runtime、stats provider 一起迁走
- 不改外部实例 API、AWD proxy、defense SSH、practice/contest/challenge 的公开契约

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice1-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice2-implementation-plan.md`
- `code/backend/internal/module/runtime/ports/http.go`
- `code/backend/internal/module/runtime/application/commands/instance_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`
- `code/backend/internal/module/runtime/application/queries/instance_service.go`
- `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/app/composition/instance_module.go`

## Current Baseline

- app/composition 层已经有 `InstanceModule` 和 `ContainerRuntimeModule` 两个视图
- 但底层 owner 仍没有真正落到 `internal/module/instance`：
  - `instance` 命令/查询服务还在 `runtime/application`
  - `proxy ticket` contract 和服务还在 `runtime/ports`、`runtime/application/queries`
  - `runtime_cleaner` 背后的 scheduler / maintenance 逻辑还在 `runtime/application/commands`
- 这意味着 app 层虽然已经把“实例 owner”讲清了，底层物理目录和 contract owner 仍然在说反话

## Chosen Direction

这轮先迁“owner 语义”，不急着迁“所有实现文件”：

1. 新建 `internal/module/instance/ports`
   - 承接实例查询 / 状态 / proxy ticket / teacher filter / AWD scope 这些 instance owner contract
2. 新建 `internal/module/instance/application/commands` 与 `queries`
   - 迁入实例命令服务、实例查询服务、proxy ticket 服务、实例维护调度服务
3. `internal/module/runtime/ports` 对已迁 contract 保留 type alias 兼容入口
4. `internal/module/runtime/application/*` 对已迁服务保留 constructor / type 兼容薄层
5. `runtime/runtime/module.go` 继续作为底层总 builder，但实例 owner 的实际实现来源改为 `instance` 包

这样能先把“谁拥有实例业务规则”从 `runtime` 里剥离出来，同时避免这一轮把 `runtime` 整包撕开。

## Ownership Boundary

- `instance/ports`
  - 负责：实例 owner 的稳定 contract，包括实例查询、状态写入、proxy ticket scope、教师实例筛选
  - 不负责：Docker engine、container file、image probe、runtime stats
- `instance/application`
  - 负责：实例命令/查询、proxy ticket、实例维护调度
  - 不负责：HTTP handler、container adapter、runtime cleaner cron、practice topology provisioning
- `runtime`
  - 负责：当前仍未迁走的 container-facing adapter、HTTP handler、兼容 facade
  - 不负责：继续作为实例 owner use case 的长期归属地

## Change Surface

- Create: `code/backend/internal/module/instance/doc.go`
- Create: `code/backend/internal/module/instance/ports/ports.go`
- Create: `code/backend/internal/module/instance/application/commands/*.go`
- Create: `code/backend/internal/module/instance/application/queries/*.go`
- Create: `code/backend/internal/module/instance/architecture_test.go`
- Modify: `code/backend/internal/module/runtime/ports/http.go`
- Modify: `code/backend/internal/module/runtime/application/commands/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service.go`
- Modify: `code/backend/internal/module/runtime/application/queries/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Create: `docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice3-review.md`

## Task Slices

### Slice 1: 为 instance owner 建立底层 home

目标：

- 新增 `instance/ports`
- 新增 `instance/application/commands|queries`
- 迁入 instance / proxy ticket / maintenance 的实际实现

Validation:

- `rg -n "package (commands|queries)|type ProxyTicketClaims|type TeacherInstanceFilter" code/backend/internal/module/instance code/backend/internal/module/runtime -g '*.go'`
- `cd code/backend && go test ./internal/module/instance/... ./internal/module/runtime/application/...`

Review focus:

- 新 `instance` 包是否只承接实例 owner 规则，没有顺手吃进 container adapter
- moved contract 是否只保留一份 owner 定义，而不是两边复制发散

### Slice 2: 保留 runtime 兼容薄层并接回现有 builder

目标：

- `runtime/ports` 对已迁 contract 保留 alias
- `runtime/application/*` 对已迁服务保留兼容 constructor / type
- `runtime/runtime/module.go` 仍能装配现有 router / composition，不引入行为变化

Validation:

- `rg -n "type .* = .*instance|return instance" code/backend/internal/module/runtime -g '*.go'`
- `cd code/backend && go test ./internal/module/runtime/... ./internal/app/...`

Review focus:

- runtime 兼容层是否足够薄，没有把新逻辑又写回 runtime
- `runtime_cleaner`、proxy ticket、instance handler 的现有调用链是否保持可用

### Slice 3: 文档、review 证据和最终验证

目标：

- 更新设计稿与架构事实，明确“底层 instance use case 已迁、runtime 仍保留 adapter/facade”
- 补一份独立 review 存档，修复 material findings 后再复验

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- 文档是否把 `runtime` 已经完全拆空写成既成事实
- review 存档是否明确当前仍残留的 runtime 兼容面，而不是只写“已完成迁移”

## Risks

- 如果 `instance/ports` 与 `runtime/ports` 双边定义漂移，会出现名义迁移、实际双 owner
- 如果 runtime 兼容层太厚，后续调用方仍会继续新增到 `runtime/*`，迁移就会停在半路
- 如果 `runtime/runtime/module.go` 继续直接 new 旧 runtime use case，而不是 new `instance` 实现，目录迁移会变成假动作
- 如果把 provisioning / cleanup 一起强行迁走，会把本轮改动面扩大到 container adapter，review 成本过高

## Verification Plan

1. `cd code/backend && go test ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...`
2. `cd code/backend && go test ./internal/module/...`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- target ownership explicit：实例 owner 的 contract 和 use case 开始落到 `internal/module/instance`
- shared landing zone explicit：container adapter、HTTP handler、cleaner cron 仍暂留 `runtime`，不在这一刀混迁
- not output-only：不只是改文档或改 composition 名字，而是把底层文件归属和 contract owner 真正改掉
- hidden redesign risk reduced：后续再拆 `runtime/infrastructure` 或 `runtime/api/http` 时，实例业务规则已经不需要继续从宽泛 runtime 包里找
