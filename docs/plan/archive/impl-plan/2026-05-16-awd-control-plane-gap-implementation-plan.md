# AWD Control Plane Gap Implementation Plan

## Objective

补齐 AWD runtime follow-up backlog 里剩余的三类控制面能力，并保证控制动作真正能收口到运行时入口，而不是只改某一层状态：

- 退赛：把某支队伍从 AWD runtime desired set 和访问入口里移除
- 停用某队服务：按 `contest × team × service` scope 禁止该服务继续被自动或显式拉起
- 人工 suppress 某些 scope：只抑制 automatic desired reconcile，不误伤显式人工操作

本次目标不是只补一个 admin 开关，而是让控制状态同时作用于：

- desired runtime reconcile
- 用户 / 管理员显式启动与重启
- 用户实例可见性与访问 URL
- AWD 攻击代理与防守 SSH / workbench 入口

## Non-goals

- 不把“退赛”回写成 `contest_registrations` 的新状态流转，也不改报名审核产品语义
- 不扩展 scoreboard、round service check 或战报统计口径
- 不补新的前端页面；本次只提供后端 API 和 orchestration 数据面
- 不把人工 suppress 设计成新的通用 runtime retry 子系统

## Inputs

- `docs/todos/2026-05-16-awd-runtime-followup.md`
- `docs/plan/impl-plan/2026-05-16-awd-desired-runtime-reconciliation-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-16-awd-runtime-hardening-implementation-plan.md`
- `code/backend/internal/module/practice/application/commands/{contest_instance_scope.go,instance_start_service.go,awd_desired_runtime_reconciler.go,contest_awd_operations.go}`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/runtime/infrastructure/{repository.go,awd_target_proxy_repository.go}`
- `code/backend/internal/module/instance/application/{queries/proxy_ticket_service.go,awd_defense_workbench_service.go}`
- `code/backend/internal/module/contest/application/commands/{participation_register_commands.go,participation_review_commands.go}`

## Brainstorming Summary

候选方向：

1. 直接把“退赛”建模成 `contest_registrations.status`
   - 拒绝：当前运行时访问和 AWD 代理入口对 team scope 主要靠 `team_members` 放行，不会随报名状态自动收紧；只改报名状态会留下旧实例与代理访问漏口
2. 在 `practice` / `runtime` owner 下新增 AWD scope control 持久化
   - 采用：控制语义本来就是 team / team-service scope，能直接作用到 desired set、显式启停和运行时访问
3. 让人工 suppress 复用 automatic desired reconcile Redis state
   - 部分采用：automatic failure state 仍留在现有 Redis store；manual suppress 用独立持久化 control state 表达，避免成功恢复时误删人工控制
4. 退赛 / 停用后仅阻止未来重建，不处理当前运行中的实例
   - 拒绝：控制面会变成“名义禁用、实例仍可用”；至少要停掉当前 scope 的 active runtime，并让运行时访问查询同步收口

## Chosen Direction

- 新增 `awd_scope_controls` 持久化表，owner 放在 runtime / practice 交界的基础模型层
  - 行级表达一条控制，不用多布尔字段混合多种语义
  - `scope_type` 只支持：
    - `team`
    - `team_service`
  - `control_type` 只支持：
    - `retired`
    - `service_disabled`
    - `desired_reconcile_suppressed`
- “退赛”在本次里定义为 AWD team runtime retirement
  - 作用对象是 `contest × team`
  - 不改报名审核记录；只改变 AWD runtime desired set 和访问面
- “停用某队服务”与“人工 suppress”都作用于 `contest × team × service`
  - `service_disabled`：阻止 desired reconcile、用户显式启停、管理员显式启停，并收紧相关访问入口
  - `desired_reconcile_suppressed`：只阻止 automatic desired reconcile；显式管理员 / 用户操作仍允许
- 当 `retired` 或 `service_disabled` 被打开时：
  - 清理对应 scope 的 automatic desired reconcile failure state
  - 停掉该 scope 当前 active instance，并把实例状态转为 `stopped`
- 运行时访问查询收口到同一套 control join
  - 实例列表 / 实例访问
  - AWD target proxy
  - AWD defense SSH / workbench

## Ownership Boundary

- `contest participation`
  - 负责：报名创建、审核与查询
  - 不负责：AWD runtime team/service scope 控制
- `practice application`
  - 负责：admin AWD 控制命令、显式启停前 gate、desired reconcile desired set 过滤、orchestration 查询出 control 状态
  - 不负责：contest 报名产品状态机
- `runtime / instance infrastructure`
  - 负责：把 runtime access / proxy scope 查询收口到 control state
  - 不负责：解释 control 的业务来源

## Change Surface

- Add: `.harness/reuse-decisions/awd-control-plane-gap.md`
- Add: `docs/plan/impl-plan/2026-05-16-awd-control-plane-gap-implementation-plan.md`
- Add: `code/backend/internal/model/awd_scope_control.go`
- Add: `code/backend/migrations/000008_create_awd_scope_controls.{up,down}.sql`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`
- Modify: `code/backend/internal/module/practice/application/commands/{service.go,contest_awd_operations.go,contest_instance_scope.go,instance_start_service.go,awd_desired_runtime_reconciler.go}`
- Modify: `code/backend/internal/module/practice/api/http/handler.go`
- Modify: `code/backend/internal/dto/contest_awd_instance.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/{repository.go,awd_target_proxy_repository.go}`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: 相关 `practice / runtime / app` 测试

## Task Slices

- [ ] Slice 1: AWD scope control persistence 与聚合读模型
  - Goal
    - 有一套可持久化、可枚举、可按 scope 查询的 control state
  - Validation
    - `cd code/backend && go test ./internal/module/practice/infrastructure -count=1 -timeout 5m`
  - Review focus
    - 控制语义是否足够显式
    - 是否避免把 team / service / suppress 混成一个 owner 不清的布尔对象

- [ ] Slice 2: practice command gate 与 desired reconcile 收口
  - Goal
    - `retired / service_disabled / desired_reconcile_suppressed` 分别按预期作用于显式启停与 automatic reconcile
  - Validation
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'Test(Service|ReconcileDesiredAWD).*Control' -count=1 -timeout 5m`
  - Review focus
    - manual suppress 是否只影响 automatic reconcile
    - retire / disable 是否会清理 active runtime 与 stale desired state

- [ ] Slice 3: runtime / instance access 查询收口
  - Goal
    - 退赛或停用后，不再能通过旧实例列表、proxy、defense SSH 继续访问
  - Validation
    - `cd code/backend && go test ./internal/module/runtime -run 'Test.*AWD.*Control' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/instance/... -run 'Test.*AWD.*Control' -count=1 -timeout 5m`
  - Review focus
    - team member join 是否仍存在绕过口
    - attacker / victim 双侧 scope 是否都被正确过滤

- [ ] Slice 4: admin API 与 orchestration 数据面
  - Goal
    - 管理端能查看当前 control rows，并通过独立 endpoint 切换 retired / disabled / suppressed
  - Validation
    - `cd code/backend && go test ./internal/app -run 'TestFullRouter.*AWD.*Control' -count=1 -timeout 5m`
  - Review focus
    - handler / service / repository 的 owner 是否清楚
    - API 是否避免一个 generic payload 混装多种语义

## Risks

- 如果 retire / disable 只挡 desired reconcile，不挡 runtime access query，旧实例仍可继续被访问
- 如果 manual suppress 和 automatic failure state 共用同一删除逻辑，恢复成功时会误清人工 suppress
- 如果 retire 定义落在报名状态而不是 team/service runtime scope，AWD team member 入口会继续旁路放行
- 如果停用后不主动停掉 active runtime，管理员会看到“已停用但实例仍在工作”的假收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'Test(Service|ReconcileDesiredAWD).*Control' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/runtime -run 'Test.*AWD.*Control' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/instance/... -run 'Test.*AWD.*Control' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/app -run 'TestFullRouter.*AWD.*Control' -count=1 -timeout 5m`
5. `cd code/backend && go test ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/instance/... -count=1 -timeout 10m`

## Rollback / Recovery Notes

- 如果 control join 误挡访问，可先删除对应 `awd_scope_controls` 行并重试显式管理员启动
- 如果 retire / disable 停机动作出现问题，删除控制行后可以用现有 admin AWD start endpoint 手工恢复
- 如果 manual suppress 误配，只需删除对应 suppress control，不影响 automatic failure state store 的独立结构

## Architecture-Fit Evaluation

- 这版把 AWD 控制面继续收口在 runtime / practice scope owner，没有把 runtime 语义塞回报名审核模块
- desired reconcile 的 automatic failure backoff 仍然是 automatic runtime repair；manual suppress 只是上层 control gate，不会反向污染 automatic state owner
- runtime access query、attack proxy、defense SSH 都用同一张 control 表过滤，避免每条入口各写一套特例
