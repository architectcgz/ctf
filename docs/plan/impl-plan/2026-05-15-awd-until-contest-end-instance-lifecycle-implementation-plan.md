# AWD Until Contest End Instance Lifecycle Implementation Plan

## Objective

把 AWD 队伍服务实例的生命周期 owner 从通用 `container.default_ttl` 收口到比赛窗口：

- AWD 实例创建、复用和重启时，`expires_at` 统一跟随 `contest.end_time`
- 比赛进入 `ended` 时，主动清理该比赛下的 AWD 实例运行态、defense workspace companion container，并收口未完成的 AWD service operation，而不是只清 Redis live 状态缓存

## Non-goals

- 不实现赛前预热 / 一键预热 AWD 队伍服务
- 不改变普通练习实例、Jeopardy 竞赛实例和 preview 运行态的 TTL 语义
- 不新增 schema、状态枚举或新的实例生命周期字段
- 不改 AWD checker、轮次推进和分数计算逻辑

## Inputs

- `docs/architecture/backend/03-container-architecture.md`
- `docs/architecture/backend/05-key-flows.md`
- `docs/architecture/features/AWD开赛就绪门禁设计.md`
- `code/backend/internal/config/config.go`
- `code/backend/internal/model/contest.go`
- `code/backend/internal/model/instance.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/app/composition/contest_module.go`
- `code/backend/internal/app/composition/runtime_module.go`

## Brainstorming Summary

候选方向：

1. 继续复用 `container.default_ttl`，只把值调大
   - 拒绝：owner 仍在容器通用配置，AWD 语义依旧错误，也会误伤普通练习实例
2. AWD 改成无 TTL，仅依赖比赛结束时手工或自动清理
   - 拒绝：会把 `expires_at` 口径变成特例，侵入查询、剩余时间、cleaner 和边界判断
3. AWD 复用现有 `expires_at` 机制，但把值改为 `contest.end_time`
   - 采用：最小改动即可把 owner 挂回 `contest` 域，同时保留现有 cleaner / 查询 / 剩余时间链路

## Chosen Direction

- 普通实例继续使用 `time.Now() + container.default_ttl`
- AWD 队伍服务实例使用 `contest.end_time`
- AWD 实例若已存在且被复用，`expires_at` 同步到当前比赛 `end_time`
- AWD 实例重启时，新的 `expires_at` 也同步到当前比赛 `end_time`
- 比赛状态迁移到 `ended` 时，side effect 除清 Redis live 状态外，还主动清理该比赛下的 AWD 实例运行态、defense workspace companion container，并把活跃中的 AWD service operation 收口到失败终态，最后把实例状态标记为 `expired`

## Ownership Boundary

- `practice/application/commands`
  - 负责：在实例创建 / 复用 / 重启时为 AWD 计算正确的 `expires_at`
  - 不负责：决定比赛何时结束或主动扫描结束比赛实例
- `contest/application/statusmachine`
  - 负责：在 `ended` transition 上触发运行态收口
  - 不负责：知道 Docker 清理细节
- `contest/infrastructure` / composition adapter
  - 负责：把 ended 比赛的 AWD 实例枚举出来，调用 runtime 清理并写回实例、defense workspace 和 active operation 状态
  - 不负责：重写实例启动主链路

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-15-awd-until-contest-end-instance-lifecycle-implementation-plan.md`
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_scope.go` or same package support helpers
- Modify: `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- Modify: `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Add or Modify: composition/runtime adapter for ended AWD runtime cleanup
- Modify: related contest / runtime tests
- Modify: `docs/architecture/backend/03-container-architecture.md`
- Modify: `docs/architecture/backend/05-key-flows.md` if current facts mention ended runtime state cleanup only covers cache

## Task Slices

- [x] Slice 1: AWD `expires_at` owner 改为 `contest.end_time`
  - Goal
    - AWD 创建、复用、重启路径不再使用 `DefaultTTL`
  - Touched modules or boundaries
    - `practice/application/commands`
  - Dependencies
    - 依赖 `contestScope.FindContestByID` 已能读取比赛窗口
  - Validation
    - 覆盖 AWD start / restart / reuse 测试，断言 `expires_at == contest.end_time`
  - Review focus
    - 普通实例 TTL 语义不能回归
    - AWD 若比赛窗口被调整，复用路径是否能同步最新 `end_time`

- [x] Slice 2: `ended` side effect 主动清理 AWD 运行态
  - Goal
    - 比赛结束时主动清理该比赛 AWD 实例、defense workspace companion container 和 active AWD operation，不再只清 live cache
  - Touched modules or boundaries
    - `contest/application/statusmachine`
    - `contest/infrastructure`
    - composition runtime adapter
  - Dependencies
    - Slice 1 完成后，AWD 实例 `expires_at` 语义已统一
  - Validation
    - ended side effect 测试断言：实例运行态被清理、defense workspace runtime 被清理并收口、active operation 被结束、Redis live 状态仍被清除
  - Review focus
    - side effect 不应清理非 AWD 或非当前 contest 实例
    - runtime cleanup 失败时 side effect 必须显式失败，便于 replay

- [x] Slice 3: 文档与后续待办同步
  - Goal
    - 更新架构事实源，并把赛前预热记录到 backlog
  - Touched modules or boundaries
    - `docs/architecture/backend`
    - `docs/todos/`
  - Validation
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus
    - 文档是否把 AWD 生命周期 owner 明确写回 `contest` 域
    - todo 是否明确写成后续能力，不与本次改动混淆

## Risks

- 比赛 `end_time` 被人工修改后，历史已运行 AWD 实例如果没有经过复用/重启路径，仍可能保留旧 `expires_at`
- `ended` side effect 若引入错误的实例筛选条件，可能误清普通竞赛实例
- 运行时清理在实例部分缺失资源时必须保持幂等，否则 transition replay 会卡住

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/... -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/app/composition -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`

## Rollback / Recovery Notes

- 不引入 migration，可按代码提交独立回滚
- 若 ended side effect 出现问题，transition 失败会保留 side effect failed 记录，可在修复后重放
- 历史 AWD 实例若已写入短 TTL，代码修复后通过复用或重启路径可逐步校正到 `contest.end_time`

## Architecture-Fit Evaluation

- owner 明确：AWD 生命周期决策回到 `contest` 时间窗，不再挂在通用容器配置
- reuse point 明确：继续复用现有 `instances.expires_at`、runtime cleaner 和 contest status side effect
- 本次既修行为，也收口结构：不新增并行 TTL 体系，不引入特殊“无过期”实例类型
