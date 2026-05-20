# AWD Runtime Hardening Implementation Plan

## Objective

把 2026-05-16 两刀 AWD 恢复能力剩余的运行噪声和配置冻结缺口补齐，并把动态 Flag 全局密钥从“只靠部署层稳定注入”收口成平台内可自动持久化的能力。

本次收口目标：

- 开赛后冻结 AWD 单题 / 单服务配置，避免 `running / frozen / ended` 期间再改 service 编排
- 给 desired reconcile 增加失败 backoff / suppress，避免长期坏配置按固定周期重复打日志和生成 operation 噪声
- 让 `CTF_CONTAINER_FLAG_GLOBAL_SECRET` 支持“优先读 env、回写持久化文件、env 丢失时自动回读文件或首次生成”的平台内自动化
- 补一份真实宿主重启恢复演练 runbook / 脚本事实，明确如何做 end-to-end 回放；不把未实际执行的宿主重启说成已验证

## Non-goals

- 不实现退赛、队伍停用、服务停用或人工 suppress 控制面
- 不引入新的管理员 UI 按钮或比赛状态枚举
- 不在本次内完成一次真实宿主机重启；这一项只补演练入口、步骤和验收口径
- 不把 desired reconcile 漂移收敛扩展成通用容器 crash 自动重启子系统

## Inputs

- `docs/plan/impl-plan/2026-05-16-awd-desired-runtime-reconciliation-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-16-awd-host-reboot-recovery-pause-implementation-plan.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/architecture/backend/05-key-flows.md`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/config/config.go`
- `docker/ctf/docker-compose.dev.yml`

## Brainstorming Summary

候选方向：

1. desired reconcile 抑噪只看 `desired_reconcile_interval`
   - 拒绝：只能放慢重试频率，仍然没有 scope 级 suppress，也无法覆盖即时配置错误
2. 用数据库 migration 给 instance 或独立表持久化 failure state
   - 暂不采用：能更强持久化，但当前目标是先止噪；复用 Redis 状态存储可以更小改动接住 scope 级 backoff / suppress
3. 在 `config.Load()` 中引入 `flag_global_secret_file`
   - 采用：这是全局密钥唯一 owner，能同时覆盖 env 优先、文件回写、首次生成和重启恢复
4. 真实宿主重启演练直接在当前任务里执行
   - 拒绝：属于高风险运行操作，且当前会话环境不适合直接重启；本次只补 runbook / 脚本和验收步骤

## Chosen Direction

- `ContestAWDServiceService.ensureMutableAWDContest`
  - 收口到与 `ChallengeService.ensureMutableContest` 一致的“开赛后不可改配置”规则
- `practice` 新增 desired reconcile state store
  - 以 `contest × team × service` 为 key 记录 `failure_count / next_attempt_at / suppressed_until / last_error`
  - automatic failure 走指数退避，达到阈值后 suppress 一段时间
  - 成功恢复为 active instance 后自动清空状态
- `config.Load()`
  - 新增 `container.flag_global_secret_file`
  - 启动时按 `env -> persisted file -> generated file` 顺序解析，并保证结果仍通过现有长度校验
- `docs/operations`
  - 新增 AWD 宿主重启恢复演练 runbook，明确前置条件、真实重启步骤、恢复后验收点和需要保留的证据

## Ownership Boundary

- `contest/application/commands`
  - 负责：冻结 AWD service 编排修改
  - 不负责：扩展比赛标题、描述或通用基础字段的修改策略
- `practice/application/commands`
  - 负责：scope 级 desired reconcile backoff / suppress、生效与清理时机
  - 不负责：人工 suppress、退赛或 scope 停用控制面
- `config`
  - 负责：动态 Flag 全局密钥的 env / file 自动化解析与持久化
  - 不负责：声明生产部署必须把秘密放在哪种外部 secret manager
- `docs/operations`
  - 负责：真实宿主重启恢复演练步骤与验收口径
  - 不负责：把未执行的重启演练写成通过结论

## Change Surface

- Add: `.harness/reuse-decisions/awd-runtime-hardening.md`
- Add: `docs/plan/impl-plan/2026-05-16-awd-runtime-hardening-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go`
- Add: `docs/operations/awd-host-reboot-recovery-drill.md`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/{service.go,awd_desired_runtime_reconciler.go,awd_desired_runtime_reconciler_test.go,instance_provisioning.go}`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/config/{config.go,config_test.go}`
- Modify: `code/backend/internal/pkg/redis/keys.go`
- Modify: `code/backend/configs/{config.yaml,config.prod.yaml}`
- Modify: `docker/ctf/docker-compose.dev.yml`
- Modify: `docs/architecture/backend/{03-container-architecture.md,05-key-flows.md}`

## Task Slices

- [ ] Slice 1: AWD 开赛后配置冻结
  - Goal
    - `running / frozen / ended` 的 AWD contest 不再允许 create / update / delete service 配置
  - Validation
    - `cd code/backend && go test ./internal/module/contest/application/commands -run 'TestContestAWDServiceService.*Immutable' -count=1 -timeout 5m`
  - Review focus
    - 是否与现有 contest challenge 冻结规则一致
    - 是否没有误伤 draft / registration 的赛前编排

- [ ] Slice 2: desired reconcile failure backoff / suppress
  - Goal
    - scope 长期失败时停止按固定 interval 产生日志和自动 operation 噪声
  - Validation
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances.*(Backoff|Suppress)|TestProvisionInstance.*DesiredReconcile' -count=1 -timeout 5m`
  - Review focus
    - suppress 是否只影响 automatic desired reconcile，不影响显式人工操作
    - 成功恢复后是否会清空 suppress 状态

- [ ] Slice 3: FLAG_GLOBAL_SECRET 自动持久化
  - Goal
    - env 存在时回写持久化文件；env 不存在时可从持久化文件恢复；首次无 secret 时自动生成
  - Validation
    - `cd code/backend && go test ./internal/config -run 'TestLoad.*FlagGlobalSecret' -count=1 -timeout 5m`
  - Review focus
    - 是否仍保持 env 优先
    - 是否避免生成过短或空 secret

- [ ] Slice 4: 运行事实同步
  - Goal
    - 架构文档和恢复演练 runbook 同步当前行为与待人工执行的步骤
  - Validation
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus
    - 是否明确区分“平台已具备恢复能力”和“真实宿主重启演练尚需人工执行”

## Risks

- 如果 suppress 状态只在内存里，API 重启后会回到高频噪声
- 如果 suppress 误伤显式人工操作，会让管理员难以主动试错恢复
- 如果 secret 自动生成但没有稳定持久化路径，容器重建后仍会丢失已有动态 Flag 语义
- 如果 runbook 写成“已验证”，会把真实运行风险掩盖掉

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'TestContestAWDServiceService.*Immutable' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances.*(Backoff|Suppress)|TestProvisionInstance.*DesiredReconcile' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/config -run 'TestLoad.*FlagGlobalSecret' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module/practice/application/commands -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`

## Rollback / Recovery Notes

- desired reconcile 抑噪如果出现误判，可先把相关 backoff / suppress 配置调回最小值，相当于回退到“只按 interval 重试”
- `flag_global_secret_file` 如需停用，可继续通过环境变量显式覆盖，并移除持久化文件路径配置
- 文档 / runbook 只影响事实表达，不影响线上行为

## Architecture-Fit Evaluation

- AWD 赛中配置冻结继续落在 `contest command` owner，没有把规则散到 handler 或前端
- desired reconcile 仍是 automatic runtime gap repair，不会伪装成更宽的人工控制面
- global secret 自动持久化收口在配置加载 owner，避免实例 / flag 生成路径各自补文件逻辑
- 真实宿主重启演练仍保留为运行侧步骤，不会把“代码具备恢复能力”和“已完成真实回放”混为一谈
