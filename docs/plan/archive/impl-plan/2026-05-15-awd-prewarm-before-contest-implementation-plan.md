# AWD Prewarm Before Contest Implementation Plan

## Objective

为管理员补一条明确的 AWD 赛前预热链路，解决宿主机重启、Docker runtime 清空或演示前实例未拉起时，需要在开赛前一次性把队伍服务实例预热起来的问题。

本次目标收口为：

- 保留现有单格启动接口，继续服务单个队伍服务实例的启动 / 重试
- 新增专用批量预热入口，负责“单队全部可见服务”或“整场全部队伍 × 可见服务”的赛前预热
- 预热链路复用现有实例启动 owner，已有且未过期实例直接复用，不重复创建
- 返回结果矩阵，明确区分 `started / reused / failed`，便于赛前排查

## Non-goals

- 不实现“管理员一键重启全部运行中实例”
- 不改 AWD 比赛中的重启语义、checker、轮次推进或评分逻辑
- 不新增实例生命周期字段、数据库表或后台预热任务
- 不把当前单格启动接口重写成多语义复合入口

## Inputs

- `docs/todos/2026-05-15-awd-prewarm-before-contest.md`
- `docs/architecture/features/竞赛题目编排工作台设计.md`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdServiceOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts`

## Brainstorming Summary

候选方向：

1. 继续让前端循环调用单格启动接口
   - 拒绝：没有批量 owner，没有结果矩阵，也无法把赛前预热和比赛中临时补启动区分开
2. 直接把 `POST /admin/contests/:id/awd/instances` 扩成单格 / 单队 / 全量三种复合语义
   - 拒绝：单接口语义会变混，handler 校验和前端契约都会变得不清楚
3. 保留单格启动接口，新增专用 `prewarm` 批量入口
   - 采用：单格启动继续保持简单，赛前预热获得独立 owner、独立状态限制和独立结果模型

## Chosen Direction

- 继续保留：
  - `GET /admin/contests/:id/awd/instances`
  - `POST /admin/contests/:id/awd/instances`
- 新增：
  - `POST /admin/contests/:id/awd/instances/prewarm`
- 语义拆分：
  - 单格启动：允许 `registration / running / frozen`，用于单点启动与失败重试
  - 批量预热：只允许 `registration`，避免把比赛中的批量补启动混成“赛前预热”
- 批量预热默认只覆盖 `is_visible=true` 的 AWD 服务
- `team_id` 为空时，预热整场所有队伍；有值时，只预热指定队伍的全部可见服务
- 每个目标单元继续复用现有 `startChallengeWithScope` 链路，并把结果映射成：
  - `started`
  - `reused`
  - `failed`

## API Contract Owner

- `handler`
  - 负责：解析 `contest_id`、绑定可选 `team_id`、拒绝非法 JSON
  - 不负责：决定允许哪些比赛状态、也不负责枚举服务矩阵
- `application`
  - 负责：校验 AWD 模式、状态窗口、团队目标、可见服务范围，并生成结果矩阵
  - 不负责：实现数据库排序或额外补默认服务列表
- `repository`
  - 负责：按既有顺序返回队伍 / 服务 / 现有实例数据
  - 不负责：预热状态判断、过滤规则或结果聚合

## Ownership Boundary

- `practice/application/commands`
  - 负责：批量预热 owner、状态窗口限制、结果矩阵聚合、单格启动 registration 放开
  - 不负责：改变实例创建主链路或新增并行实例体系
- `practice/api/http`
  - 负责：暴露批量预热入口并返回 DTO
  - 不负责：批量循环或错误收口
- `frontend/api + contest-awd-admin model`
  - 负责：在报名阶段调用批量预热接口并展示汇总提示
  - 不负责：决定后端哪些服务可被预热

## Change Surface

- Add: `.harness/reuse-decisions/awd-prewarm-before-contest.md`
- Add: `docs/plan/impl-plan/2026-05-15-awd-prewarm-before-contest-implementation-plan.md`
- Modify: `code/backend/internal/dto/contest_awd_instance.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- Modify: `code/backend/internal/module/practice/api/http/handler.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin/contests.ts`
- Modify: `code/frontend/src/api/__tests__/admin.test.ts`
- Modify: `code/frontend/src/features/contest-awd-admin/model/useAwdServiceOperations.ts`
- Modify: `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts`
- Modify: `docs/architecture/features/竞赛题目编排工作台设计.md`
- Modify: `docs/todos/2026-05-15-awd-prewarm-before-contest.md`

## Task Slices

- [x] Slice 1: 后端批量预热契约与应用行为
  - Goal
    - 新增批量预热 DTO、handler、route 和 application owner
  - Touched modules or boundaries
    - `practice/application/commands`
    - `practice/api/http`
    - `app/router`
  - Validation
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestService(StartAdminContestAWDPrewarm|StartAdminContestAWDTeamService)' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1 -timeout 5m`
  - Review focus
    - 单格启动与批量预热的状态窗口是否分离清楚
    - 失败项是否不会中断整个批量过程

- [x] Slice 2: 前端在报名阶段接入批量预热
  - Goal
    - `startTeamAllServices` / `startAllTeamServices` 在 `registering` 下走新接口，其余状态保留既有单格循环
  - Touched modules or boundaries
    - `frontend api`
    - `contest-awd-admin model`
  - Validation
    - `cd code/frontend && npx vitest run src/api/__tests__/admin.test.ts src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts --runInBand`
  - Review focus
    - 仅报名阶段切换到批量预热，避免比赛中行为回归
    - toast 汇总是否能反映 started / reused / failed

- [x] Slice 3: 架构事实与 backlog 同步
  - Goal
    - 把“赛前预热是独立批量入口”写回事实源，并把 todo 标成已完成或仅保留后续项
  - Touched modules or boundaries
    - `docs/architecture/features`
    - `docs/todos`
  - Validation
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus
    - 文档是否明确区分单格启动与赛前批量预热

## Risks

- 如果批量预热复用了单格入口但没有把 `reused` 暴露出来，前端仍无法区分真实新启动和复用
- 如果前端在 `running / frozen` 也无条件切到批量接口，会把既有比赛中批量补启动路径变成报错
- 如果 team captain 缺失、service snapshot 异常或 runtime 创建失败时没有按单元记录失败，赛前排查仍然会被整批错误淹没

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestService(StartAdminContestAWDPrewarm|StartAdminContestAWDTeamService)' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1 -timeout 5m`
3. `cd code/frontend && npx vitest run src/api/__tests__/admin.test.ts src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts --runInBand`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`

## Rollback / Recovery Notes

- 不涉及 migration，可按提交粒度直接回滚
- 若批量预热接口有问题，单格启动接口仍保留，可临时退回单点启动
- 若报名阶段批量预热失败，管理员仍可通过单格启动或再次触发预热重试

## Architecture-Fit Evaluation

- owner 明确：赛前预热由独立批量入口负责，单格启动不再背负多语义
- reuse point 明确：继续复用既有实例启动链路、实例编排读模型和前端 orchestration 面板
- 这刀只补赛前准备能力，不把运行中全量重启语义混入同一入口
