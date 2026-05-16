# AWD Host Reboot Recovery Pause Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 宿主机整机重启后，AWD 比赛的 round / freeze / end 时间应视为暂停，等实例恢复完成后再继续推进。

**Architecture:** 不先扩展通用 `paused/interrupted` 赛事状态机，也不引入前端新状态枚举；本次围绕“宿主机重启”这一具体事故场景，给 `contests` 持久化累计暂停秒数 `paused_seconds`，后台统一用 `effectiveNow = now - pausedDuration` 和 `effectiveEnd = end_time + pausedDuration` 解释 AWD 比赛时间窗。API 启动期同步检测主机重启、恢复活跃实例，并把停机与恢复耗时累计进 `paused_seconds`，同时用 `runtime_recovery_key + runtime_recovery_applied_seconds` 做同一次 outage 的幂等补差，再刷新 AWD 实例 `expires_at`。这样能在不扩散前端契约的前提下，把 round、checker、封榜和 `until_contest_end` 的时间口径一起拉齐。

**Tech Stack:** Go, Gin, GORM, Redis, Docker runtime maintenance jobs, PostgreSQL contest/instance persistence

---

## Objective

把当前“宿主机停机后，AWD round 仍按墙钟推进”的错误语义收口成恢复前阻塞、恢复后顺延的实际行为。

本次目标收口为：

- API 启动时识别“宿主机发生过重启”，避免把普通 API 容器重启误判成停机事故
- 在 contest 状态任务和 AWD round 任务启动前，同步完成活跃实例恢复
- 对处于 `running / frozen` 的 AWD 比赛累计 `paused_seconds`，并让 `freeze_time / end_time` 通过有效时间窗派生顺延语义
- 同步刷新活跃 AWD 实例 `expires_at`，让 `until_contest_end` 跟随新的比赛结束时间

## Non-goals

- 不实现通用管理员手动 `paused/interrupted` 状态
- 不修改前端 `ContestStatus` 枚举，也不新增比赛暂停 UI
- 不改变 Jeopardy 比赛的时间推进语义
- 不处理“仅 API 重启但 Docker / 宿主机未重启”这类更宽的控制面故障

## Inputs

- `docs/architecture/backend/design/contest-status-state-machine.md`
- `docs/architecture/backend/03-container-architecture.md`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_plan.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/http_server.go`
- `code/backend/internal/pkg/redis/keys.go`

## Brainstorming Summary

候选方向：

1. 新增通用 `paused/interrupted` 赛事状态
   - 拒绝：语义完整，但会扩散到状态机、OpenAPI、前端枚举、筛选和展示，本刀超出“先把宿主机重启场景收住”的最小边界
2. 只在恢复后给 stopped 容器做自动重启，不改比赛时间
   - 拒绝：实例恢复窗口里 round/checker 仍会继续推进，平台事故仍会被记成队伍失守
3. API 启动期识别宿主机重启，先同步恢复实例，再把停机与恢复窗口累计到活跃 AWD 比赛的 `paused_seconds`
   - 采用：能把宿主机停机与恢复窗口一起计入暂停时间，同时不改前端状态契约，也不直接重写比赛原始 `start_time / freeze_time / end_time`

## Chosen Direction

- 新增一个 API 启动期的“宿主机重启恢复”后台 job，注册在 `instance` 模块并先于 contest jobs 启动
- 通过持久化的 `boot_id + last_heartbeat_at` 判断“当前 API 启动是否跨过了宿主机重启”
- 检测到宿主机重启时：
  - 先给活跃 AWD 比赛补一段“停机至当前”的暂停时长，并刷新这些比赛下活跃实例的 `expires_at`
  - 再同步执行一次 `ReconcileLostActiveRuntimes`
  - 以 `recovery_finished_at - recovery_started_at` 把恢复耗时继续累计到同一批活跃 AWD 比赛；同一次 outage 通过 `runtime_recovery_key + runtime_recovery_applied_seconds` 只补差值，不重复累计
- 恢复完成后再允许 `contest_status_updater` 与 `awd_round_updater` 启动
- contest / scoreboard / flag / instance 侧统一改为读取 `effectiveNow / effectiveEnd`，不在各处各自重写暂停语义

## API Contract Owner

- `instance/application/commands`
  - 负责：宿主机重启检测、启动期恢复编排、比赛时间顺延触发和活跃 AWD 实例过期时间刷新
  - 不负责：实现 AWD round 计算、checker 执行或前端展示
- `contest/repository`
  - 负责：查询活跃 AWD 比赛并原子累加 `paused_seconds`
  - 不负责：判断是否属于宿主机重启，或决定何时触发恢复
- `runtime/repository`
  - 负责：批量刷新活跃 AWD 实例 `expires_at`
  - 不负责：重新解释比赛时间窗或决定顺延幅度
- `composition`
  - 负责：把启动期恢复 job 放到正确的后台任务顺序里
  - 不负责：承载恢复业务规则

## Ownership Boundary

- `instance maintenance`
  - 负责：实例运行态恢复和启动期宿主机恢复编排
  - 不负责：改写 AWD round 推进逻辑
- `contest status / round jobs`
  - 负责：继续基于比赛时间窗推进状态和 round
  - 不负责：自行推断宿主机是否停机过
- `frontend`
  - 不负责：本刀不需要感知新的比赛状态

## Change Surface

- Add: `.harness/reuse-decisions/awd-host-reboot-recovery-pause.md`
- Add: `docs/plan/impl-plan/2026-05-16-awd-host-reboot-recovery-pause-implementation-plan.md`
- Add: `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- Add: `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
- Add: `code/backend/internal/module/runtime/infrastructure/platform_runtime_state_store.go`
- Add: `code/backend/internal/module/contest/domain/contest_timing.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_runtime_recovery_repository.go`
- Add: `code/backend/internal/module/contest/infrastructure/contest_awd_runtime_recovery_repository_test.go`
- Add: `code/backend/internal/module/runtime/infrastructure/repository_awd_expiry_refresh_test.go`
- Add: `code/backend/internal/app/contest_paused_seconds_migration_test.go`
- Add: `code/backend/migrations/000007_add_contest_paused_seconds.up.sql`
- Add: `code/backend/migrations/000007_add_contest_paused_seconds.down.sql`
- Modify: `code/backend/internal/module/runtime/infrastructure/repository.go`
- Modify: `code/backend/internal/module/contest/application/commands/*`
- Modify: `code/backend/internal/module/contest/application/jobs/*`
- Modify: `code/backend/internal/module/contest/application/queries/*`
- Modify: `code/backend/internal/module/contest/infrastructure/contest_status_repository.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/pkg/redis/keys.go`
- Modify: `docs/architecture/backend/03-container-architecture.md`

## Task Slices

- [x] Slice 1: 启动期宿主机重启检测与恢复编排
  - Goal
    - 在 contest jobs 启动前完成一次宿主机重启检测与同步实例恢复
  - Touched modules or boundaries
    - `instance/application/commands`
    - `runtime/infrastructure`
    - `app/composition`
  - Validation
    - `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryService' -count=1 -timeout 5m`
  - Review focus
    - 是否只在宿主机重启时触发，不会把普通 API 重启误判成停机事故
    - contest jobs 是否确实在恢复完成后才启动

- [x] Slice 2: AWD 比赛时间窗顺延与实例过期时间刷新
  - Goal
    - 对活跃 AWD 比赛累计 `paused_seconds`，并刷新活跃 AWD 实例 `expires_at`
  - Touched modules or boundaries
    - `contest/ports`
    - `contest/infrastructure`
    - `runtime/infrastructure`
  - Validation
    - `cd code/backend && go test ./internal/module/contest/infrastructure -run 'Test.*Shift.*AWD.*Contest|Test.*ActiveAWD.*Expiry' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryServiceExtendsActiveAWDContests' -count=1 -timeout 5m`
  - Review focus
    - 是否只影响 `running / frozen` 的 AWD 比赛
    - `effectiveNow / effectiveEnd` 是否在 round、封榜和实例过期时间上保持一致
    - 实例 `expires_at` 是否和新的比赛有效结束时间对齐

- [x] Slice 3: 架构事实同步
  - Goal
    - 把“宿主机重启恢复会顺延活跃 AWD 比赛时间窗”写回容器编排事实源
  - Touched modules or boundaries
    - `docs/architecture/backend`
  - Validation
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus
    - 文档是否明确区分“宿主机重启恢复顺延”与“通用人工暂停”这两个不同层次

## Risks

- 如果宿主机重启检测只依赖“上次心跳”而不区分 boot identity，会把普通 API 重启误判成比赛暂停
- 如果启动期恢复 job 不阻塞 contest jobs，round/checker 仍可能在实例恢复完成前提前推进
- 如果只累计 `paused_seconds` 但仍有逻辑直接读取原始 `end_time` / `freeze_time` / `time.Now()`，round、封榜和实例过期时间会再次分叉
- 如果顺延范围扩到了非 AWD 比赛，会改变现有 Jeopardy 时间窗语义

## Verification Plan

1. `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/infrastructure -run 'TestAddPausedDurationToActiveAWDContests' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/runtime/infrastructure -run 'Test.*ActiveAWD.*Expiry' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
5. `cd code/backend && go test ./internal/module/contest/application/commands -count=1 -timeout 5m`
6. `cd code/backend && go test ./internal/module/contest/application/jobs -count=1 -timeout 5m`
7. `python3 scripts/check-docs-consistency.py`
8. `bash scripts/check-consistency.sh`

## Rollback / Recovery Notes

- 涉及 `contests.paused_seconds` migration；回滚时需要先确认线上是否已经依赖该字段
- 若宿主机重启检测出现误判，可先回退启动期恢复 job，保留既有 runtime cleaner 自动恢复
- 若有效时间窗逻辑有问题，可按受影响比赛回调 `paused_seconds`，不需要重写原始 `start_time / end_time`

## Architecture-Fit Evaluation

- owner 明确：宿主机重启恢复由 `instance` 启动期恢复 service 负责，比赛时间顺延由 `contest repository` 负责，实例过期时间刷新由 `runtime repository` 负责
- reuse point 明确：继续复用现有 `maintenance_service`、runtime repository 和 contest 时间窗 owner，不新增并行 pause 子系统
- 这刀解决的是“宿主机整机重启”这一具体事故，不假装已经落地通用手动暂停状态机
- 若后续要支持管理员显式暂停或更广义控制面故障暂停，应作为独立方案进入 `contest` 状态机扩展，而不是在本刀里继续叠加
