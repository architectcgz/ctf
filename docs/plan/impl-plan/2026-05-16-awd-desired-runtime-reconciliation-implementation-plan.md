# AWD Desired Runtime Reconciliation Implementation Plan

## Objective

把 AWD 宿主恢复与赛中漂移修复，从“只恢复当前活跃实例”扩展为“持续收敛应该活着的队伍服务集合”：

- 启动恢复时，不只修 stopped / lost active runtime，还要补齐 `running / frozen` 赛事里缺失的 `team × visible service`
- 赛中持续周期性对账，覆盖宿主恢复之外的缺失实例、失败实例和未成功入队场景
- 优先复用已有实例行与 nonce，避免无必要地生成新 flag；只有没有可复用历史实例时才新建实例

## Non-goals

- 不实现队伍退赛、服务停用、人工 suppress 或赛中改单题配置
- 不把 registration 阶段 prewarm 扩展成比赛中的批量恢复
- 不新增前端按钮或新的管理员批量重启入口
- 不改 AWD 攻击、checker、计分规则本身

## Inputs

- `docs/architecture/features/校园级CTF-AWD模式完整设计.md`
- `docs/architecture/features/竞赛题目编排工作台设计.md`
- `docs/plan/impl-plan/2026-05-16-awd-host-reboot-recovery-pause-implementation-plan.md`
- `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/app/composition/practice_module.go`

## Brainstorming Summary

候选方向：

1. 继续扩 `maintenance_service`
   - 拒绝：它的 owner 是“已有活跃实例运行态修复”，不是 `contest × team × service` 期望集合；把差集推导塞进去会让实例 owner 混入赛事编排语义
2. 把“期望调和”继续并入 prewarm
   - 拒绝：文档已明确 prewarm 只覆盖 `registration`，赛中恢复不是供给预热问题
3. 新增 AWD desired runtime reconciler，由 `practice` 负责期望集合推导，并在启动恢复与周期 loop 中复用
   - 采用：`practice` 已拥有 teams / services / instances / start-restart 语义，最接近 `service_id` 维度的运行态编排 owner

## Chosen Direction

- `startup_runtime_recovery_service`
  - 继续作为平台停机检测与恢复触发器
  - 启动恢复时按顺序执行：
    1. 累计 AWD 比赛 `paused_seconds`
    2. `ReconcileLostActiveRuntimes`
    3. `ReconcileDesiredAWDInstances`
    4. 完成恢复后补齐恢复耗时并保存 heartbeat
- `practice` 新增 `ReconcileDesiredAWDInstances(ctx)`：
  - 查询 `running / frozen` 且未到有效结束时间的 AWD contests
  - 对每个 contest 计算 `teams × visible services`
  - 先看是否已有 active instance（`pending / creating / running`）
  - 如果没有 active instance：
    - 优先复用该 scope 下最近的 restartable instance（含 `failed`），重置为 `pending / creating`
    - 若不存在历史实例，再走新建实例路径
  - 所有自动补齐统一入现有 pending / provisioning 流程，不新增并行 provisioning runner
- 周期性调和不单独起新 job，而是并入 `practice_instance_scheduler` loop，但使用独立节流间隔，避免每秒全量扫描所有 AWD contests

## Ownership Boundary

- `instance/application/commands/startup_runtime_recovery_service`
  - 负责：启动期停机检测、恢复顺序编排、恢复窗口暂停累加
  - 不负责：自行推导 `contest × team × service` 期望集合
- `instance/application/commands/maintenance_service`
  - 负责：活跃实例视角的 runtime 修复
  - 不负责：补齐从未存在 active row 的队伍服务
- `practice/application/commands`
  - 负责：AWD 队伍服务期望态收敛、failed 历史实例复用、缺失实例创建入队
  - 不负责：Docker 容器健康检查、stopped container 直接拉起
- `practice/application/commands/instance_provisioning_scheduler`
  - 负责：承接调和器补出的 pending 实例，并按现有限流规则 provisioning
  - 不负责：决定哪些队伍服务“应该活着”
- `composition`
  - 负责：把 startup recovery 与 practice reconciler 接上
  - 不负责：承载恢复业务规则

## Change Surface

- Add: `.harness/reuse-decisions/awd-desired-runtime-reconciliation.md`
- Add: `docs/plan/impl-plan/2026-05-16-awd-desired-runtime-reconciliation-implementation-plan.md`
- Add: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- Add: `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/infrastructure/repository.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`
- Modify: `code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service_test.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`
- Modify: `code/backend/internal/app/composition/practice_module.go`
- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/configs/config.yaml`
- Modify: `code/backend/configs/config.prod.yaml`
- Modify: `docs/architecture/backend/03-container-architecture.md`
- Modify: `docs/architecture/backend/05-key-flows.md`

## Task Slices

- [ ] Slice 1: `practice` 期望调和 owner 与查询口径
  - Goal
    - 让 `practice` 拥有“列出应收敛 AWD contests + 计算队伍服务差集”的独立入口
  - Touched modules or boundaries
    - `practice/ports`
    - `practice/infrastructure`
    - `practice/application/commands`
  - Validation
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances' -count=1 -timeout 5m`
  - Review focus
    - 是否只覆盖 `running / frozen` 且未到 effective end 的 AWD contests
    - 是否只把 `visible service` 纳入期望集合

- [ ] Slice 2: 缺失实例补齐与 failed 历史实例复用
  - Goal
    - 缺少 active instance 时优先复用 restartable 历史实例，否则新建 pending 实例
  - Touched modules or boundaries
    - `practice/application/commands`
    - `practice/infrastructure`
  - Validation
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances(ReusesFailedInstance|CreatesMissingInstance)' -count=1 -timeout 5m`
  - Review focus
    - 是否保留旧 nonce / old flag 语义
    - 是否复用现有 pending / provisioning owner，而不是直接绕过 scheduler 启容器

- [ ] Slice 3: 启动恢复与周期 loop 接线
  - Goal
    - 启动恢复时按“active recovery -> desired reconciliation”顺序执行；赛中持续周期性收敛
  - Touched modules or boundaries
    - `instance/application/commands`
    - `app/composition`
    - `practice/runtime`
  - Validation
    - `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryService.*Desired' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestRunProvisioningLoop.*Desired' -count=1 -timeout 5m`
  - Review focus
    - 是否没有引入 `instance -> practice` 的反向编译依赖
    - contest jobs 启动前的 startup recovery 是否已可拿到 desired reconciler

- [ ] Slice 4: 文档与配置事实同步
  - Goal
    - 把“AWD 期望调和”与现有 scheduler/config owner 写回事实源
  - Touched modules or boundaries
    - `docs/architecture/backend`
    - `config`
  - Validation
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus
    - 文档是否明确区分 prewarm、active runtime recovery、desired runtime reconciliation 三层职责

## Risks

- 如果差集补齐直接调用 `startChallengeWithScope`，会跳过 failed 实例复用，导致无必要的新 nonce / 新 flag
- 如果周期性调和不节流，`practice_instance_scheduler` 会退化成高频全量 AWD 扫描
- 如果 startup recovery 与 desired reconciler 接线顺序不对，contest jobs 可能在缺失实例补齐前先继续 round/checker
- 如果查询口径把 `registration` 或已结束 contest 带进来，会把 prewarm 和赛中恢复语义混淆

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances|TestRunProvisioningLoop.*Desired' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryService.*Desired|TestStartupRuntimeRecoveryService.*Stale' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/app/composition -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`

## Rollback / Recovery Notes

- 不涉及 migration，可按代码与配置改动整体回滚
- 若周期调和负载过高，可先把 `desired_reconcile_interval` 调大或临时关闭 scheduler，再单独回退该特性
- 若启动恢复接线有问题，可先保留 `ReconcileLostActiveRuntimes`，关闭 startup desired reconciliation，等待修复后再重新打开

## Architecture-Fit Evaluation

- 这刀新增的不是“管理员一键重启”，而是 `service_id` 维度的期望态收敛 owner
- `maintenance_service` 的实例视角不被扩成赛事编排器，结构边界保持清楚
- `practice` 继续作为 AWD 实例创建 / 重启 / 编排的 owner，期望调和与 provisioning 仍落在同一模块，不引入第二套实例生命周期
- 启动恢复只做触发与顺序控制，不承载业务推导，后续要支持退赛、停用或人工 suppress 时也有明确落点
