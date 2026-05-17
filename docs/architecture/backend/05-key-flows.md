# 关键流程设计文档

> 状态：Current
> 事实源：`code/backend/internal/module/practice/application/commands/`、`code/backend/internal/module/contest/application/jobs/`、`code/backend/internal/module/challenge/application/commands/`
> 替代：无

## 定位

本文档说明当前已经采用的关键业务流程 owner、状态推进和主要副作用。

- 负责：说明实例编排、访问票据、竞赛状态机、AWD 轮次与题包导入这些主链路现在怎么跑。
- 不负责：把旧的候选方案、未启用队列或未来异步重构写成当前事实；细节流程图仅作为展开说明。

## 当前设计

- `code/backend/internal/module/practice/application/commands/instance_start_service.go`、`instance_provisioning_scheduler.go`、`runtime_container_create.go`、`awd_desired_runtime_reconciler.go`
  - 负责：处理练习 / 竞赛 / AWD 开题链路中的作用域锁、实例占位、`pending -> creating -> running` 状态推进、端口与运行时补偿；普通实例 `expires_at` 继续走 `container.default_ttl`，AWD 队伍服务实例 `expires_at` 改由 `contestdomain.ContestEffectiveEndTime(contest)` owner
  - 负责：按 `container.scheduler.desired_reconcile_interval` 周期性收敛 `running / frozen` AWD 比赛里“应该活着的 `team × visible service`”，缺失时优先复用已有实例行 / nonce，再交回现有 provisioning loop；同一 scope 连续遇到坏配置或 provisioning 失败时，会用 Redis 记录 backoff / suppress 状态，窗口内直接跳过自动 operation
  - 不负责：引入独立 Redis 队列或让 HTTP 请求直接等待完整 Docker 冷启动；当前排队事实仍落在 `instances` 表与调度器循环里

- `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`、`instance_service.go`
  - 负责：把实例访问、AWD 攻击访问和 AWD 防守 SSH 入口统一收敛到 proxy ticket / access URL 链路，避免直接暴露容器地址
  - 不负责：让前端或外部工具绕过平台鉴权直接使用底层容器网络信息

- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`、`status_transition_service.go`、`statusmachine/side_effects.go`
  - 负责：推进 `contests.status` 的时间窗状态机、调度锁续租、transition replay 和副作用重放，保证 `registration / running / frozen / ended` 口径一致；`ended` side effect 会同步清理 AWD live 状态、队伍服务实例 runtime、defense workspace companion container，并收口未完成的 AWD service operation
  - 不负责：让管理命令或定时任务直接改表而跳过 transition 记录与副作用协议

- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`、`challenge_support.go`、`code/backend/internal/module/contest/domain/contest.go`
  - 负责：把 AWD service 编排配置的可变窗口收口到 `draft / registration`。比赛进入 `running / frozen / ended` 后，`ContestAWDService` 的 create / update / delete 会统一返回 `ErrContestImmutable`，不再允许赛中改单题、改单 checker、改单服务快照或可见性
  - 不负责：提供赛中热更新 `service_snapshot`、`runtime_config`、`score_config` 的旁路入口

- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`、`awd_round_updater.go`、`awd_checks.go`、`awd_service_check_empty_result.go`
  - 负责：驱动 AWD 轮次同步、checker 执行、快照写入、空实例兜底结果和计分输入表更新
  - 不负责：把 AWD 裁判链路外包给独立引擎进程；当前主链路仍在 `contest` 模块内部 jobs

- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`、`awd_challenge_import_service.go`、`challenge_service.go`
  - 负责：处理题包 preview / commit / self-check、附件持久化、镜像构建源准备和题目自检的运行时探测
  - 不负责：在导入 preview 阶段启动正式学员实例，或把导入工作目录作为长期运行态的一部分

## 接口或数据影响

- 关键状态字段包括 `instances.status`、`instances.share_scope`、`contests.status`、`contest_status_transitions.side_effect_status`、`awd_rounds.status`、`awd_team_services.service_status`、`image_build_jobs.status`。
- 关键运行态降噪数据还包括 Redis `ctf:awd:desired_reconcile:state:<contest_id>:<team_id>:<service_id>`，用于记录自动补齐失败次数、下一次重试时间和 suppress 窗口；自动补齐成功或 scope 已经 active 时会清空。
- 关键流程入口包括实例启动 / 续期 / 访问、contest 更新与冻结、AWD service preview / readiness / rounds / attack logs，以及 challenge import / commit / self-check；契约以 `docs/contracts/openapi-v1.yaml` 为准。
- 主要副作用落在 PostgreSQL、Redis 锁与缓存、Docker runtime、附件目录和 registry / build source 目录；这些副作用的 owner 分别受 `practice`、`contest`、`runtime`、`challenge` 模块控制。

## Guardrail

- 练习 / 实例流程集成：`code/backend/internal/app/practice_flow_integration_test.go`
- 路由级状态矩阵与实例访问：`code/backend/internal/app/full_router_state_matrix_integration_test.go`
- 开题调度与补偿：`code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`、`instance_start_service_test.go`
- AWD desired reconcile backoff / suppress：`code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- 竞赛状态机与锁续租：`code/backend/internal/module/contest/application/jobs/status_updater_test.go`、`awd_round_scheduler_runtime_internal_test.go`
- AWD service 配置冻结：`code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- AWD 轮次与 checker：`code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`、`code/backend/internal/module/contest/application/commands/awd_service_test.go`
- 题包导入 / 附件 / 自检：`code/backend/internal/app/challenge_import_integration_test.go`、`code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`

## 历史迁移

- 当前事实已经从“单篇流程蓝图”收口为“owner service / scheduler / job + 契约 + integration tests”的组合事实源。
- 下文保留的长时序图、决策表和异常表仍可作为详细参考；若出现旧的同步启动、旧状态名或未启用方案描述，以前述当前设计和代码为准。

## 目录

1. [靶机实例启动流程](#1-靶机实例启动流程)
2. [Flag 提交与计分流程](#2-flag-提交与计分流程)
3. [动态计分流程](#3-动态计分流程)
4. [排行榜实时更新流程](#4-排行榜实时更新流程)
5. [容器生命周期回收流程](#5-容器生命周期回收流程)
6. [竞赛状态流转](#6-竞赛状态流转)
7. [AWD 轮次执行流程](#7-awd-轮次执行流程)
8. [作弊检测流程](#8-作弊检测流程)
9. [竞赛高峰排队机制](#9-竞赛高峰排队机制)

---

## 1. 靶机实例启动流程

### 1.1 流程描述

用户请求启动靶机实例时，系统需要完成前置校验、实例占位、资源调度、容器创建、Flag 注入等一系列操作。

当前实现（2026-03-31）已经切换为**两段式编排**：

- HTTP 请求阶段只负责校验、预留端口、写入 `instances(status=pending)`；
- 后台 `practice_instance_scheduler` 以受控并发推进 `pending -> creating -> running`；
- 同一个 scheduler 循环还会按独立节流间隔执行 AWD desired runtime reconciliation，补齐 `running / frozen` 比赛里缺失的 `team × visible service`；
- 同一 `contest_id + team_id + service_id` scope 如果连续遇到坏配置或 provisioning 失败，会写入 Redis backoff / suppress 状态，在 `next_attempt_at` 或 `suppressed_until` 之前不再重复创建自动 operation；
- 如果显式关闭 `container.scheduler.enabled`，仍可回退到同步启动路径。

这样可以把 Docker 冷启动从 HTTP 请求里剥离出来，避免 50 人同时开题时接口直接超时。

### 1.2 时序图

```mermaid
sequenceDiagram
    participant 用户
    participant API as API Gateway
    participant PS as PracticeService
    participant PG as PostgreSQL
    participant SCH as ProvisioningScheduler
    participant Docker

    用户->>API: 启动请求
    API->>PS: 鉴权+限流
    PS->>PG: 事务内锁用户/队伍作用域
    PS->>PG: 检查已有实例与用户配额
    PS->>PG: 预留 host_port + INSERT instance(status=pending)
    PG-->>PS: instance_id
    PS-->>API: 返回 instance(status=pending)
    API-->>用户: 排队中

    loop 后台轮询
        SCH->>PG: count(creating), count(creating+running)
        alt 有可用容量
            SCH->>PG: 查询最早 pending 实例
            SCH->>PG: 原子更新 pending -> creating
            SCH->>Docker: 创建网络/容器并启动
            alt 成功
                SCH->>PG: UPDATE instance(status=running, runtime/access_url)
            else 失败
                SCH->>PG: UPDATE instance(status=failed) + 释放端口
            end
        else 容量不足
            Note over SCH: 保持 pending，等待下一轮
        end
    end
```

### 1.3 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 用户/队伍并发控制 | 数据库事务内 `SELECT ... FOR UPDATE` 锁用户或队伍记录 | 不依赖进程内锁，多后端实例也能成立 |
| 实例数限制 | 数据库 `COUNT(*) WHERE status IN ('pending','creating','running')` | 排队态也计入占位，防止一个用户在高峰期刷满队列 |
| 启动节流 | `container.scheduler.max_concurrent_starts` | 显式限制同一时刻的 Docker 冷启动数 |
| 宿主机容量保护 | `container.scheduler.max_active_instances` 控制 `creating + running` | 不再默认宿主机可以无限接单 |
| AWD 期望调和节流 | `container.scheduler.desired_reconcile_interval` | 把“应该活着的队伍服务补齐”并到现有 scheduler 循环，同时避免高频全量扫描所有 AWD 比赛 |
| Flag 生成算法 | `HMAC-SHA256(global_secret, uid + ":" + challenge_id + ":" + instance_nonce)` | 当前训练/普通竞赛链路由全局密钥与实例随机 `nonce` 共同生成动态 Flag；AWD 轮次会把队伍、赛事、轮次等上下文拼接为 `nonce` 后复用同一 HMAC 逻辑 |
| 资源限制 | Docker `--memory`、`--cpus`、`--pids-limit` 从 challenge 配置读取 | 防止恶意容器耗尽宿主机资源 |
| 网络隔离 | 每个实例独立 Docker Network，仅暴露指定端口 | 防止实例间横向渗透 |
| 同步/异步切换 | `container.scheduler.enabled` 控制 | 测试环境或小规模环境可回退同步启动 |
| 实例有效期 | 普通实例 `expires_at = now() + container.default_ttl`；AWD 队伍服务实例 `expires_at = contest.end_time + paused_seconds` | 让练习实例继续按通用 TTL 回收，同时避免 AWD 比赛中的队伍服务在比赛未结束前被 2 小时 TTL 提前销毁，且宿主机停机期间会跟随比赛暂停一起顺延 |

### 1.4 异常处理

| 异常场景 | 处理策略 | 回滚操作 |
|----------|----------|----------|
| 用户并发实例数超限 | 返回 429，提示"已达最大实例数" | 不创建实例记录 |
| 竞赛不在 running 状态 | 返回 403，提示"竞赛未开始或已结束" | 不创建实例记录 |
| 端口预留失败 | 返回 500 | 回滚事务，不留下脏实例 |
| 调度器抢占失败 | 其他后端已领取该实例，当前 worker 跳过 | 无需回滚 |
| 容器创建/启动失败 | 返回 500，记录错误 | 删除已创建的 Network，更新实例状态为 failed |
| 宿主机容量达到上限 | 实例继续停留在 pending | 不回滚，等待下一轮调度 |
| 数据库写入运行态失败 | 返回 500 | 停止并删除容器、删除 Network，并释放端口 |

### 1.5 当前实现说明

当前代码没有单独引入 Redis 队列，而是直接复用 `instances` 表作为排队事实源：

- `StartChallengeWithContext` 在事务内完成用户/队伍作用域加锁、实例复用判断、用户配额判断、端口预留和 `status=pending` 的实例创建；
- `RunProvisioningLoop` 周期性扫描最早的 `pending` 实例，并通过原子状态推进把任务领取到当前 worker；
- 领取成功后再调用运行时编排能力创建容器，最终把实例更新为 `running` 或 `failed`。

### 1.6 并发与一致性考虑

- **同一用户并发启动**：事务内锁用户或队伍记录，保证检查已有实例、计数、端口预留和实例占位是串行的。
- **多后端实例抢任务**：调度器先查 `pending`，再执行 `UPDATE ... WHERE status='pending'` 抢占；只有一个实例能成功推进到 `creating`。
- **实例计数准确性**：用户配额按 `pending + creating + running` 统计；宿主机容量按 `creating + running` 统计。
- **资源泄漏防护**：容器创建失败时立即走补偿清理，运行中实例仍由既有过期清理器和孤儿资源清理器兜底。
- **同步回退**：当 `container.scheduler.enabled=false` 时，仍允许使用同步创建路径，便于测试和小规模部署。

### 1.7 实例访问与 Proxy Ticket

实例启动成功后，用户访问实例并不是直接连接容器地址，而是先通过平台访问入口换取一个短时 `proxy ticket`。

当前规则固定如下：

- `proxy ticket` 在“访问实例”时签发，不在“启动实例”时签发
- ticket 用于平台代理访问链路的短时鉴权
- ticket 当前携带 `user_id`、`instance_id`、`contest_id`、`share_scope`、`purpose` 以及 AWD 目标访问所需的队伍/服务上下文

它和最终提交内容不是同一类凭证：

- 所有题目统一提交 `flag`
- `proxy ticket` 只负责“实例访问”与“上下文传递”

因此，实例相关链路实际上分成两段：

1. `POST /api/v1/challenges/:id/instances`
   负责创建或复用实例
2. `POST /api/v1/instances/:id/access`
   负责校验访问权限、签发 `proxy ticket`、返回平台代理访问地址

---

## 2. Flag 提交与计分流程

### 2.1 流程描述

用户提交 Flag 后，系统需要完成身份校验、频率限制、Flag 验证、计分、幂等去重、事件发布等操作。
该流程是平台的核心交互路径，必须保证**高并发下的正确性**和**严格的幂等性**。

### 2.2 时序图

```mermaid
sequenceDiagram
    participant 用户
    participant API as API Gateway
    participant SS as SubmitService
    participant Redis
    participant PG as PostgreSQL
    participant EB as EventBus

    用户->>API: 提交 Flag
    API->>SS: 鉴权+限流

    SS->>Redis: 频率限制检查 INCR submit:{uid}:{cid} EXPIRE 60s
    Redis-->>SS: 当前次数

    Note over SS: 校验: 次数 <= 10次/分钟

    SS->>PG: 查询实例状态
    PG-->>SS: instance(status, flag, challenge)

    Note over SS: 校验: 实例 status=running
    Note over SS: 校验: 竞赛 status=running|frozen

    SS->>PG: 幂等检查: SELECT 1 FROM submissions WHERE scoring_unit(team_id/user_id)=? AND challenge_id=? AND contest_id=? AND result='correct'
    PG-->>SS: 存在/不存在

    Note over SS: 已解出则直接返回"已完成"

    SS->>Redis: Flag 验证 (静态: 直接比对 / 动态: HMAC 重新计算比对)
    Redis-->>SS: 匹配/不匹配

    Note over SS: 不匹配: 记录错误提交, 返回"Flag错误"

    rect rgb(200, 230, 200)
        Note over SS, PG: 匹配，进入计分事务
        SS->>PG: BEGIN TX
        SS->>PG: INSERT submission (result='correct', 唯一约束防并发重复)
        PG-->>SS: OK/DUPLICATE
        Note over SS: DUPLICATE: 幂等返回"已完成"
        SS->>PG: 计算得分(base × weight - hints)
        SS->>PG: UPDATE user_score SET score=score+?, solved_count=solved_count+1
        SS->>PG: COMMIT
        PG-->>SS: OK
    end

    Note over SS: 关键路径同步：计分+排行榜更新在同一请求内完成
    SS->>Redis: ZINCRBY leaderboard:real:{cid} score uid
    SS->>Redis: EXISTS contest:frozen:{cid}
    alt 未冻结
        SS->>Redis: ZINCRBY leaderboard:public:{cid} score uid
    end

    SS->>EB: 发布得分事件 ScoreEvent{uid, cid, score, timestamp}（仅用于 WebSocket 推送等非关键通知）
    SS-->>API: 返回结果
    API-->>用户: 解题成功
```

### 2.3 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 频率限制 | Redis `INCR` + `EXPIRE`，每用户每题 10 次/分钟 | 防止暴力枚举 Flag，使用滑动窗口更精确但此场景固定窗口足够 |
| Flag 验证方式 | 静态题：`SHA-256(submitted_flag + salt)` 与数据库存储的哈希比对；动态题：用 `instance.nonce` 重新 HMAC 计算后比对 | 静态 Flag 不存明文（仅存哈希+盐），动态 Flag 不存明文（只存 nonce），防止数据库泄露后 Flag 暴露 |
| 幂等保证 | 数据库部分唯一索引：竞赛 `UNIQUE(user_id, challenge_id, contest_id) WHERE result='correct' AND contest_id IS NOT NULL`；练习 `UNIQUE(user_id, challenge_id) WHERE result='correct' AND contest_id IS NULL` | 数据库层面兜底，即使并发请求同时通过应用层检查也不会重复计分 |
| 计分公式 | `score = base_score × difficulty_weight - hint_penalty` | 简单直观，hint_penalty 从已使用的提示累加 |
| 得分更新 | 原子累加 `UPDATE ... SET score=score+?, solved_count=solved_count+1` | 无需乐观锁版本校验，PostgreSQL 行锁保证原子性；同一用户不会高频得分，行锁持有时间极短 |
| 事件发布 | 事务提交后同步更新 Redis 排行榜，再异步发布 `ScoreEvent` 用于 WebSocket 推送 | 排行榜更新是关键路径，必须同步完成；WebSocket 推送等非关键通知走异步事件总线 |

### 2.4 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| 频率超限 | 返回 429，提示"提交过于频繁，请稍后再试"，不记录提交历史 |
| 实例不存在或已停止 | 返回 400，提示"靶机实例未运行" |
| 竞赛已结束 | 返回 403，提示"竞赛已结束，无法提交" |
| Flag 格式非法（不符合 `flag{...}` 格式） | 返回 400，前端校验 + 后端兜底，不消耗频率限制次数 |
| Flag 错误 | 返回 200（result='incorrect'），记录错误提交到 submissions 表 |
| Redis 排行榜更新失败 | 记录错误日志，主流程仍返回成功（数据库已持久化）；后台补偿任务定期从数据库重建排行榜 |
| 事件发布失败 | 记录错误日志，不影响主流程返回（事件仅用于 WebSocket 推送等非关键通知） |

### 2.5 并发与一致性考虑

- **同一用户并发提交同一题**：数据库部分唯一索引保证只有一条正确记录（竞赛/练习分别约束）。第二个并发请求会触发唯一约束冲突，应用层捕获后返回"已完成"。INSERT 和 UPDATE 在同一事务内，唯一约束冲突时整个事务回滚，不会出现"INSERT 成功但 UPDATE 未执行"的中间状态。
- **不同用户同时提交**：各自独立事务，互不影响。`score=score+?` 原子累加由 PostgreSQL 行锁保证，不同用户操作不同行，无冲突。
- **计分与排行榜的强一致性**：排行榜 Redis 更新在事务提交后同步执行（同一请求内），不依赖异步事件总线。如果 Redis 更新失败，数据库得分已持久化，后台补偿任务每 5 分钟从数据库全量重建排行榜。事件总线仅用于 WebSocket 推送等非关键异步通知。
- **frozen 状态下的提交**：竞赛进入 frozen 状态后，Flag 提交仍然正常处理并记录得分，同步更新 `leaderboard:real` 但不更新 `leaderboard:public`（排行榜冻结逻辑见流程 4）。

---

## 3. 动态计分流程

### 3.1 流程描述

动态计分（Dynamic Scoring）根据解题人数动态调整题目分值：解题人数越多，分值越低。
当有新的解题记录产生时，需要**重新计算该题分值**，并**回溯更新所有已解出该题的计分单元（队伍/个人）得分**，最终同步到排行榜。

计分公式：
```
score = max(min_score, base_score × (decay_factor ^ solve_count))
```

参数说明：
- `base_score`：题目初始分值（如 1000）
- `min_score`：最低分值下限（如 100）
- `decay_factor`：衰减因子（如 0.95）
- `solve_count`：当前解题总计分单元数（含本次；组队赛按队伍，未组队按个人）

### 3.2 时序图

```mermaid
sequenceDiagram
    participant SS as SubmitService
    participant SC as ScoreCalcService
    participant Redis
    participant PG as PostgreSQL
    participant EB as EventBus

    SS->>SC: ScoreEvent(新解题)

    SC->>Redis: 获取分布式锁 SETNX dynscore:lock:{challenge_id} TTL=10s
    Redis-->>SC: OK

    SC->>PG: 查询解题人数(含本次)
    PG-->>SC: solve_count

    Note over SC: 计算新分值: new_score = max(min_score, base × decay^count)

    SC->>PG: 查询旧分值
    PG-->>SC: old_score

    Note over SC: score_diff = new_score - old_score (通常为负值)

    rect rgb(200, 230, 200)
        Note over SC, PG: 事务
        SC->>PG: BEGIN TX
        SC->>PG: UPDATE challenge SET current_score=new_score, solve_count+1 WHERE id=? AND version=?
        SC->>PG: 查询所有已解出该题的用户列表
        PG-->>SC: user_ids[]
        SC->>PG: 批量更新用户得分: UPDATE user_score SET score=score+score_diff WHERE user_id IN (...)
        SC->>PG: COMMIT
        PG-->>SC: OK
    end

    loop 循环每个受影响用户
        SC->>Redis: ZINCRBY leaderboard:{contest_id} score_diff user_id
        Redis-->>SC: OK
    end

    SC->>Redis: 释放分布式锁

    SC->>EB: 发布分值变更事件 ScoreChangeEvent{challenge_id, new_score, affected_users[]}
```

### 3.3 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 并发控制 | Redis 分布式锁 `dynscore:lock:{challenge_id}`，TTL 30s + watchdog 续期（每 10s 续期一次） | 同一题目同时有多人解出时，必须串行计算分值；TTL 10s 不够（锁内含数据库事务+Redis 批量更新），改为 30s + watchdog 自动续期防超时 |
| 分值更新粒度 | 计算 `score_diff` 增量，批量 `score + diff` | 避免全量重算所有用户总分，只调整差值，性能更优 |
| 排行榜同步 | 循环 `ZINCRBY` 更新每个受影响用户 | ZINCRBY 是原子操作，单个用户的排行榜更新不需要额外加锁 |
| 乐观锁 | challenge 表 `version` 字段 | 防止并发更新 solve_count 导致计算错误 |
| 批量更新策略 | 校园级场景（解题人数通常 < 500）单事务批量 UPDATE；超过 500 人时分批，每批 200，使用断点续传机制 | 保证分值变更的原子性；分批时记录已完成批次，失败后从断点恢复而非全量重试 |

### 3.4 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| 分布式锁获取失败 | 重试 3 次，间隔 100ms；仍失败则将事件放入延迟队列，5s 后重新消费 |
| 乐观锁冲突（version 不匹配） | 重新读取最新 solve_count 和 version，重新计算后重试 |
| 批量更新部分失败 | 单事务场景：事务回滚，整体重试。分批场景：记录已完成批次到 Redis（`dynscore:progress:{challenge_id}`），失败后从断点恢复 |
| Redis 排行榜更新失败 | 记录错误日志，标记该题需要补偿；后台任务定期从数据库重建排行榜 |
| 受影响用户数过多（>500） | 分批更新，每批 200 个用户，每批一个事务；记录断点进度，支持失败后续传 |

### 3.5 并发与一致性考虑

- **同一题目并发解题**：Redis 分布式锁保证同一时刻只有一个计分任务在执行。后到的事件等待锁释放后重新获取最新 solve_count 计算。
- **数据库与 Redis 一致性**：数据库是权威数据源，Redis 排行榜是衍生视图。如果 Redis 更新失败，后台补偿任务会从数据库重建。
- **分值回溯的公平性**：所有已解出该题的用户统一调整分值，不存在"先解题得高分、后解题得低分但先解题的人分值不变"的不公平情况。

---

## 4. 排行榜实时更新流程

### 4.1 流程描述

排行榜是竞赛的核心展示组件，需要支持**实时更新**和**高频查询**。
采用 Redis Sorted Set 作为排行榜存储，数据库作为权威数据源，通过事件驱动保持两者同步。
竞赛末期支持**榜单冻结**（frozen），冻结后前端展示停止更新，但后端继续记录真实得分，赛后揭榜。

### 4.2 数据结构设计

```
Redis Key 设计：

# 公开排行榜（前端可见，frozen 后停止写入）
leaderboard:public:{contest_id}    -> Sorted Set (score -> user_id/team_id)

# 真实排行榜（始终更新，frozen 期间也写入）
leaderboard:real:{contest_id}      -> Sorted Set (score -> user_id/team_id)

# 冻结标记
contest:frozen:{contest_id}        -> "1" (存在即冻结)

# 用户解题详情缓存
user:solves:{contest_id}:{user_id} -> Hash (challenge_id -> score)
```

### 4.3 时序图 — 得分更新

```mermaid
sequenceDiagram
    participant EB as EventBus
    participant LS as LeaderboardService
    participant Redis
    participant WS as WebSocket Hub

    EB->>LS: ScoreEvent{uid, score_diff, contest_id}

    LS->>Redis: ZINCRBY leaderboard:real:{cid} score_diff uid
    Redis-->>LS: new_score

    LS->>Redis: EXISTS contest:frozen:{cid}
    Redis-->>LS: 0(未冻结)

    LS->>Redis: ZINCRBY leaderboard:public:{cid} score_diff uid
    Redis-->>LS: OK

    LS->>Redis: ZREVRANK leaderboard:public:{cid} uid
    Redis-->>LS: rank

    LS->>WS: 推送排名变更事件 {uid, new_score, rank, cid}
    Note over WS: 广播到订阅该竞赛的所有客户端
```

### 4.4 时序图 — 排行榜查询（分页）

```mermaid
sequenceDiagram
    participant 用户
    participant API as API Gateway
    participant LS as LeaderboardService
    participant Redis

    用户->>API: 查询排行榜 (page, size)
    API->>LS: 转发请求

    LS->>Redis: ZREVRANGE leaderboard:public:{cid} start, stop WITHSCORES
    Redis-->>LS: [(uid, score)...]

    LS->>Redis: ZCARD leaderboard:public:{cid}
    Redis-->>LS: total_count

    Note over LS: 批量查询用户信息 (昵称、头像，优先从缓存读取)

    LS-->>API: 排行榜数据
    API-->>用户: JSON 响应
```

### 4.5 榜单冻结与揭榜流程

```mermaid
sequenceDiagram
    participant 触发者 as 定时任务/管理员
    participant CS as ContestService
    participant Redis
    participant WS as WebSocket Hub

    rect rgb(200, 220, 240)
        Note over 触发者, WS: 冻结（竞赛结束前 1 小时自动触发）
        触发者->>CS: 触发冻结
        CS->>Redis: SET contest:frozen:{cid} "1"
        Redis-->>CS: OK
        CS->>WS: 推送冻结通知 {type:"frozen", cid, freeze_at}
        Note over WS: 广播
        Note over 触发者, WS: 此后 ScoreEvent 仍写入 real 排行榜，但不再写入 public 排行榜
    end

    rect rgb(220, 240, 200)
        Note over 触发者, WS: 揭榜（竞赛结束后管理员手动触发）
        触发者->>CS: 触发揭榜
        CS->>Redis: COPY leaderboard:real:{cid} -> leaderboard:public:{cid}
        Redis-->>CS: OK
        CS->>Redis: DEL contest:frozen:{cid}
        Redis-->>CS: OK
        CS->>WS: 推送揭榜事件 {type:"reveal", cid}
        Note over WS: 广播最终排名
    end
```

### 4.6 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 存储结构 | Redis Sorted Set，score 为复合编码值（`total_score * 1e10 + (MAX_TS - last_solve_ts)`），member 为 user_id | O(logN) 插入/更新，O(logN+M) 范围查询；复合编码解决同分排序问题（同分时最后提交更早的排名靠前） |
| 双排行榜 | public（前端可见）+ real（始终更新） | 冻结期间前端看到的是冻结时刻的快照，后端继续记录真实得分 |
| 得分更新 | ZINCRBY 增量更新 | 原子操作，无需读-改-写，天然并发安全 |
| 分页查询 | ZREVRANGE + ZCARD | 降序排列，支持 offset/limit 分页 |
| WebSocket 推送 | 只推送变更事件（delta），不推送全量排行榜 | 减少带宽消耗，前端增量更新 DOM |
| 揭榜实现 | COPY real -> public（Redis 6.2+ COPY 命令） | 原子替换，避免逐条同步的中间状态 |

### 4.7 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| Redis 不可用 | 降级为数据库查询排行榜（`ORDER BY score DESC LIMIT`），性能下降但功能可用 |
| WebSocket 推送失败 | 客户端定时轮询兜底（每 10s 拉取一次排行榜） |
| public 与 real 排行榜不一致 | 后台补偿任务每 5 分钟对比两个 Sorted Set，非冻结期间自动修复 |
| 排行榜数据丢失（Redis 重启） | 从数据库 `user_score` 表全量重建，启动时自动检测并恢复 |

### 4.8 并发与一致性考虑

- **ZINCRBY 原子性**：Redis 单线程模型保证 ZINCRBY 的原子性，多个 ScoreEvent 并发更新同一用户得分不会丢失。
- **冻结判断的竞态**：冻结标记写入和得分更新之间可能存在短暂竞态（冻结瞬间有得分写入 public）。可接受，因为冻结时间点本身就有 ±1s 的容差。如需严格一致，可用 Redis Lua 脚本将"检查冻结标记 + ZINCRBY"合并为原子操作。
- **排行榜与数据库的最终一致性**：Redis 排行榜是数据库的衍生视图，数据库是权威源。任何不一致都可通过重建修复。

---

## 5. 容器生命周期回收流程

### 5.1 流程描述

容器回收是平台资源管理的核心环节，需要处理三种场景：
1. **定时过期回收**：实例到达 `expires_at` 后自动清理
2. **容器崩溃检测**：Docker 事件监听，容器异常退出后通知用户
3. **孤儿资源清理**：数据库记录与 Docker 实际状态不一致时的全量对账

三种机制互为补充，确保不会出现资源泄漏。

### 5.2 时序图 — 定时过期回收

```mermaid
sequenceDiagram
    participant Cron as Cron(每30s)
    participant RS as RecycleService
    participant PG as PostgreSQL
    participant Docker
    participant WS as 用户(WebSocket)

    Cron->>RS: 触发扫描

    RS->>PG: SELECT instances WHERE status='running' AND expires_at < now() LIMIT 50 FOR UPDATE SKIP LOCKED
    PG-->>RS: expired_list

    rect rgb(200, 230, 200)
        Note over RS, Docker: 逐个回收（并发度限制为 5）
        RS->>PG: UPDATE status='destroying'
        RS->>Docker: StopContainer(timeout=10s)
        Docker-->>RS: OK
        RS->>Docker: RemoveContainer
        Docker-->>RS: OK
        RS->>Docker: RemoveNetwork
        Docker-->>RS: OK
        RS->>PG: UPDATE status='destroyed', destroyed_at=now()
    end

    RS->>WS: 推送过期通知 {type:"expired", instance_id}
```

### 5.3 时序图 — 容器崩溃检测

```mermaid
sequenceDiagram
    participant DD as Docker Daemon
    participant EL as EventListener
    participant PG as PostgreSQL
    participant WS as 用户(WebSocket)

    DD->>EL: container.die (event stream)

    EL->>EL: 解析 container_id

    EL->>PG: SELECT instance WHERE container_id=?
    PG-->>EL: instance

    Note over EL: instance 不存在或 status!=running: 忽略

    alt exit_code=0
        Note over EL: 正常退出, 忽略
    else exit_code=137
        Note over EL: OOM Killed
    else exit_code!=0
        Note over EL: 异常崩溃
    end

    EL->>PG: UPDATE instance SET status='crashed', exit_code=?

    EL->>WS: 推送崩溃通知 {type:"crashed", instance_id, exit_code, can_restart: true/false}
```

### 5.4 时序图 — 孤儿资源清理（全量对账）

```mermaid
sequenceDiagram
    participant Cron as Cron(每5分钟)
    participant OC as OrphanCleaner
    participant PG as PostgreSQL
    participant Docker

    Cron->>OC: 触发对账

    OC->>PG: SELECT container_id, network_id FROM instances WHERE status IN ('running', 'creating')
    PG-->>OC: db_set

    OC->>Docker: docker ps -a --filter label=ctf=true
    Docker-->>OC: docker_set

    Note over OC: 对比两个集合
    Note over OC: A = docker_set - db_set (Docker有, DB无: 孤儿容器)
    Note over OC: B = db_set - docker_set (DB有, Docker无: 幽灵记录)

    rect rgb(255, 220, 220)
        Note over OC, Docker: 清理孤儿容器 A
        OC->>Docker: StopContainer
        OC->>Docker: RemoveContainer
        OC->>Docker: RemoveNetwork
        Note over OC: 记录告警日志
    end

    rect rgb(220, 220, 255)
        Note over OC, PG: 修复运行时丢失 B
        OC->>PG: UPDATE status='pending', clear runtime fields
        Note over OC: 保留实例作用域与 nonce
        PG-->>OC: practice scheduler 后续重建容器
    end
```

### 5.4.1 运行时丢失恢复

Docker daemon 关闭、宿主机重启或 Docker 运行时异常后，数据库中仍处于 `running / creating` 的实例可能已经失去实际容器。平台通过 runtime 维护任务和 practice 期望调和做主动对账：

- `maintenance_service.ReconcileLostActiveRuntimes` 负责扫描未过期的 `running / creating` 实例。
- 它会根据 `container_id` 与 `runtime_details.containers[]` 检查入口容器和拓扑容器是否仍存在且处于运行状态；若单个实例 Docker inspect 失败，只记录日志并跳过该实例，本轮继续处理其他实例。
- 若 active instance 的容器缺失或已退出，就把该实例重新置为 `pending`，交由现有 `practice_instance_scheduler` 按 `pending -> creating -> running` 流程重建。
- 重新入队时保留 `user_id / contest_id / team_id / challenge_id / service_id / share_scope / nonce / host_port / expires_at`，只清空 `container_id / network_id / runtime_details / access_url` 这类运行时字段。
- `practice.ReconcileDesiredAWDInstances` 负责 `running / frozen` AWD 比赛的差集补齐：按 `teams × visible services` 推导应该活着的 scope，如果没有 active instance，就优先复用该 scope 下最近的 restartable / failed 实例，否则新建 `pending` 实例。
- desired reconcile 在 Redis `ctf:awd:desired_reconcile:state:<contest_id>:<team_id>:<service_id>` 中记录 scope 级失败退避状态；立即配置错误和异步 provisioning 失败都会推进 `failure_count`，并根据 `container.scheduler.desired_reconcile_failure_*` 与 `desired_reconcile_suppress_*` 计算 `next_attempt_at / suppressed_until`。
- 当某个 scope 已经有 `pending / creating / running` 实例，或重建最终回到 active 状态时，desired reconcile 会主动清掉对应 Redis 状态，避免长期 suppress 卡死已经恢复的 scope。
- 多容器拓扑中任一容器丢失或退出时，active runtime recovery 会把整条实例重新入队，避免局部恢复破坏拓扑一致性。
- 启动恢复顺序是：补 `paused_seconds`、刷新活跃实例 `expires_at`、执行 active runtime recovery、执行 desired runtime reconciliation、把恢复耗时继续累计到 `paused_seconds` 并保存 heartbeat。

这两层恢复都不直接创建容器。容器创建、动态 Flag 构造、端口复用、就绪探测与失败标记继续由 practice 实例调度器统一负责。

### 5.5 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 扫描频率 | 过期回收 30s，孤儿对账 5 分钟 | 过期回收需要及时释放资源；孤儿对账是兜底机制，频率可低 |
| AWD 期望补齐频率 | `container.scheduler.desired_reconcile_interval` | 赛中差集补齐与实例 provisioning 共用同一 loop，但使用独立节流 |
| 长期坏配置降噪 | scope 级指数 backoff + 超阈值 suppress | 避免同一 `team × service` 每 15s 固定重试并持续制造日志 / operation 噪声 |
| 批量回收并发度 | 限制为 5 个并发 goroutine | 避免瞬间大量 Docker API 调用导致 daemon 压力过大 |
| 行锁策略 | `SELECT ... FOR UPDATE SKIP LOCKED` | 多个回收 worker 不会争抢同一条记录，避免死锁 |
| 容器停止超时 | 10s graceful shutdown，超时后 SIGKILL | 给容器内进程合理的清理时间 |
| 崩溃重启策略 | 用户手动触发重启；active runtime 丢失由后台重新入队；缺失 scope 由 desired reconcile 补齐 | 容器进程自身崩溃不盲目重启；Docker daemon 重启或平台恢复后的差集补齐需要平台统一收敛 |
| 孤儿容器识别 | Docker label `ctf=true` 过滤 | 只清理平台创建的容器，不误删其他容器 |
| 恢复重建入口 | 复用 practice 实例调度器 | 避免绕过并发上限、动态 Flag、端口和就绪探测逻辑 |

### 5.6 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| 单个实例 Docker inspect 失败 | 记录错误日志，跳过该实例，下轮重试 |
| 数据库为运行中但 Docker 容器缺失 | 保留实例作用域，清空运行时字段并重新入队 |
| 数据库为运行中但 Docker 容器已退出 | 按运行时丢失处理，整条实例重新入队 |
| 同一 `team × service` 连续补齐失败 | 记录 Redis backoff / suppress 状态，窗口内跳过自动 operation；等到 `next_attempt_at` 或 `suppressed_until` 后再重试 |
| 容器停止超时（SIGKILL 也失败） | 标记为 `destroy_failed`，告警通知运维人工处理 |
| Network 删除失败（仍有容器挂载） | 先强制删除残留容器，再删除 Network |
| 数据库更新失败 | 不继续创建容器，下轮对账重试 |
| 事件监听断开 | 自动重连，重连后从断点继续消费；重连期间的崩溃事件由定时扫描兜底 |

### 5.7 并发与一致性考虑

- **多 worker 并发回收**：`SKIP LOCKED` 保证不同 worker 处理不同实例，无锁竞争。
- **回收与用户操作的竞态**：用户可能在回收过程中请求续期。通过数据库状态机保证：只有 `status=running` 的实例才能续期，回收时先将状态改为 `destroying`，续期请求会因状态不匹配而失败。
- **孤儿清理的安全性**：只清理带有 `ctf=true` label 的容器，且清理前再次确认数据库中确实没有对应记录。

---

## 6. 竞赛状态流转

### 6.1 流程描述

竞赛生命周期通过有限状态机管理，状态流转由**定时任务自动驱动**和**管理员手动操作**两种方式触发。
每个状态下有明确的行为约束，确保业务逻辑在正确的时间窗口内执行。

### 6.2 状态机定义

```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> published : 管理员发布
    published --> registering : 管理员开放报名
    registering --> running : 到达开始时间（自动）
    running --> frozen : 到达冻结时间（自动）
    running --> ended : 到达结束时间（自动）
    frozen --> ended : 到达结束时间（自动）
    ended --> archived : 管理员归档

    %% 管理员紧急操作（逆向/终止转换）
    published --> draft : 管理员撤回发布
    registering --> published : 管理员关闭报名
    running --> ended : 管理员紧急结束
    frozen --> ended : 管理员紧急结束

    %% 取消终态（任何非终态均可取消）
    draft --> cancelled : 管理员取消
    published --> cancelled : 管理员取消
    registering --> cancelled : 管理员取消
    running --> cancelled : 管理员取消（紧急）
    frozen --> cancelled : 管理员取消（紧急）

    cancelled --> [*]
    archived --> [*]
```

### 6.3 状态转换规则

| 当前状态 | 目标状态 | 触发方式 | 触发条件 |
|----------|----------|----------|----------|
| draft | published | 管理员手动 | 题目数 >= 1，时间配置完整 |
| published | registering | 管理员手动 | 开放报名 |
| registering | running | 定时任务 | `now() >= start_time` |
| running | frozen | 定时任务 | `now() >= freeze_time`（可选，未配置则跳过） |
| running | ended | 定时任务 | `now() >= end_time` |
| frozen | ended | 定时任务 | `now() >= end_time` |
| ended | archived | 管理员手动 | 成绩已确认，资源已回收 |

### 6.4 各状态行为约束

| 状态 | 允许的操作 | 禁止的操作 |
|------|-----------|-----------|
| draft | 编辑竞赛信息、添加/删除题目、配置时间 | 报名、启动靶机、提交 Flag |
| published | 查看竞赛信息 | 报名（未开放）、启动靶机、提交 Flag |
| registering | 报名/退出报名、查看竞赛信息 | 启动靶机、提交 Flag |
| running | 启动靶机、提交 Flag、查看排行榜、使用提示 | 编辑题目、报名 |
| frozen | 启动靶机、提交 Flag（得分不更新公开榜） | 编辑题目、报名 |
| ended | 查看最终排行榜、查看解题记录 | 启动靶机、提交 Flag、报名 |
| archived | 查看历史记录 | 所有写操作 |

### 6.5 定时任务驱动状态流转

```mermaid
sequenceDiagram
    participant Cron as Cron(每10s)
    participant CS as ContestScheduler
    participant PG as PostgreSQL
    participant EB as EventBus

    Cron->>CS: 触发检查

    CS->>PG: SELECT contests WHERE status IN ('registering','running','frozen') AND 存在待触发的时间点
    PG-->>CS: contest_list

    Note over CS: 逐个检查时间条件

    alt registering && now()>=start_time
        CS->>PG: UPDATE status='running' WHERE id=? AND status='registering'
        PG-->>CS: affected=1
        CS->>EB: 发布状态变更事件 ContestStateEvent{cid, old, new}
    end

    alt running && now()>=freeze_time && freeze_time IS NOT NULL
        CS->>PG: UPDATE status='frozen'
    end

    alt running/frozen && now()>=end_time
        CS->>PG: UPDATE status='ended'
        Note over CS: 触发 ended side effect：清空 AWD live 状态，并同步清理该比赛 AWD 队伍服务实例、defense workspace companion container 与未完成 operation
    end
```

### 6.6 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 状态流转原子性 | `UPDATE ... WHERE status=?` 条件更新 | CAS 语义，防止并发重复流转 |
| 扫描频率 | 每 10s 一次 | 状态流转对时间精度要求不高，10s 延迟可接受 |
| 行为约束实现 | 中间件/拦截器统一校验竞赛状态 | 避免每个接口重复校验，集中管理 |
| 赛后资源清理 | `ended` side effect 同步清理 AWD live 状态、队伍服务实例、defense workspace companion container，并把未完成的 AWD service operation 收口到终态；普通实例继续由 runtime cleaner 按 `expires_at` 回收 | 保证 AWD 比赛结束后不会残留可访问队伍运行态，也不会留下仍显示为 provisioning / recovering 的历史操作 |

### 6.7 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| 定时任务宕机 | 重启后自动补偿：扫描所有"应该已流转但未流转"的竞赛 |
| 状态流转失败（数据库异常） | 记录错误日志，下轮重试；不会跳过状态 |
| 管理员误操作（如提前结束） | 提供"回退到上一状态"功能，仅限 admin 角色 |

### 6.8 并发与一致性考虑

- **多实例部署下的重复流转**：`UPDATE ... WHERE status='registering'` 的 CAS 语义保证只有一个实例成功更新，`affected_rows=0` 的实例自动跳过。
- **状态校验与业务操作的竞态**：业务接口先读取竞赛状态，再执行操作。如果在读取后、操作前状态发生变化，通过数据库约束兜底（如 Flag 提交时再次校验竞赛状态）。

---

## 7. AWD 轮次执行流程

### 7.1 流程描述

AWD（Attack With Defense）模式下，竞赛按固定时间间隔（如每 5 分钟）划分为多个轮次。
每轮需要完成：Flag 轮换注入、服务存活检测、攻击得分收集、防守得分计算、排行榜更新。
整个轮次执行是一个**严格有序的编排流程**，各步骤之间有依赖关系。

当前代码落地边界（2026-05-07）：

- 已实现后台 `AWDRoundUpdater`，按 `contest.start_time/end_time` 与全局配置 `contest.awd.scheduler_interval`、`contest.awd.round_interval` 自动补齐历史轮次并推进当前轮次。
- 已实现 Redis 幂等锁 `ctf:awd:round:lock:{contest_id}:{round_number}`，避免多实例重复推进同一轮。
- 已实现当前轮 Flag 生成与 Redis 存储：Flag 仍按 `team_id + awd_challenge_id + round_number` 生成，但写入 `ctf:awd:{contest_id}:round:{round_id}:flags` 时，字段已经切到 `{team_id}:s:{service_id}`。
- 已实现 `AWDFlagInjector` 对运行中容器的文件注入：当前按 `contest_id + team_id + service_id` 匹配比赛服务实例，并将 Flag 写入 `/flag/flag.txt`。
- 已实现统一服务定义读取：Checker、readiness 和学生工作台都从 `contest_awd_services` 读取 runtime / score / validation 配置，不再以 `contest_challenges.awd_*` 作为运行态读契约。
- 已实现正式 Checker 执行链路：对当前轮每个 `team_id + service_id` 执行 `legacy_probe / http_standard / tcp_standard / script_checker`，结果写入 `awd_team_services` 与 Redis `ctf:awd:{contest_id}:service_status`。
- 已实现学员接口按“当前轮 + victim_team_id + service_id”校验攻击 Flag，并在 `contest.awd.previous_round_grace` 宽限期内兼容上一轮 Flag，复用攻击日志去重计分。
- 当前运行态实例、轮次结果、攻击日志和流量事件都已经显式持久化 `service_id`；`awd_challenge_id` 只保留题目资产与展示角色。

### 7.2 时序图

```mermaid
sequenceDiagram
    participant RT as RoundTimer
    participant AWD as AWDRoundService
    participant Docker as Docker(各队靶机)
    participant Redis
    participant PG as PostgreSQL

    RT->>AWD: 轮次触发 (round_number=N)

    AWD->>Redis: 获取轮次锁 SETNX awd:round:lock:{cid}:{N} TTL=轮次间隔×2
    Redis-->>AWD: OK

    AWD->>PG: INSERT round (status=running)
    PG-->>AWD: round_id

    rect rgb(200, 230, 200)
        Note over AWD, Docker: 阶段1: Flag 轮换注入
        AWD->>PG: 查询所有队伍与赛事服务定义
        PG-->>AWD: team_services[]

        loop 并发(限制10), 对每个 team_id + service_id
            AWD->>AWD: 生成新Flag(team_id, awd_challenge_id, round_number)
            AWD->>Docker: Docker API CopyToContainer 写入 /flag/flag.txt
            Docker-->>AWD: OK/FAIL
        end

        AWD->>Redis: HSET awd:flags:{cid}:{N} {team_id}:s:{service_id} {flag}
        Redis-->>AWD: OK
    end
```

接续时序图：

```mermaid
sequenceDiagram
    participant RT as RoundTimer
    participant AWD as AWDRoundService
    participant Docker as Docker(各队靶机)
    participant Redis
    participant PG as PostgreSQL

    rect rgb(200, 220, 240)
        Note over AWD, Docker: 阶段2: 服务存活检测
        loop 对每个 team_id + service_id
            AWD->>PG: 按 contest_id + team_id + service_id 查询运行中实例
            PG-->>AWD: instances[]
            AWD->>Docker: HTTP GET {access_url}{health_path} 失败后 TCP dial
            Docker-->>AWD: 200/timeout
        end
        AWD->>PG: UPSERT awd_team_services (round_id, team_id, service_id, awd_challenge_id, service_status, check_result, sla_score, defense_score)
        AWD->>Redis: HSET awd:{contest_id}:service_status {team_id}:s:{service_id} {status}
    end

    rect rgb(240, 230, 200)
        Note over AWD, PG: 阶段3: 收集攻击得分
        AWD->>PG: SELECT * FROM awd_attack_logs WHERE round_id=? AND is_success=true
        PG-->>AWD: attack_records
        Note over AWD: 攻击得分规则: 每成功提交一个其他队伍的Flag +attack_score分
    end

    rect rgb(220, 240, 220)
        Note over AWD: 阶段4: 计算防守得分
        Note over AWD: 服务存活: +alive_score
        Note over AWD: 服务宕机: +0
        Note over AWD: Flag被攻破: -breach_penalty × 攻破队伍数
    end
```

接续时序图（阶段5: 汇总与更新排行榜）：

```mermaid
sequenceDiagram
    participant RT as RoundTimer
    participant AWD as AWDRoundService
    participant Redis
    participant PG as PostgreSQL
    participant WS as WebSocket Hub

    rect rgb(230, 220, 240)
        Note over AWD, PG: 阶段5: 汇总与更新排行榜
        AWD->>PG: BEGIN TX
        AWD->>PG: INSERT round_scores (round_id, team_id, attack_score, defense_score, round_total)
        AWD->>PG: UPDATE team_score SET total=total+round_total
        AWD->>PG: COMMIT
        PG-->>AWD: OK
    end

    loop 循环每个队伍
        AWD->>Redis: ZINCRBY leaderboard:public:{cid} round_total team_id
        Redis-->>AWD: OK
    end

    AWD->>PG: UPDATE round SET status='completed'

    AWD->>WS: 推送轮次结果事件 {type:"round_result", round:N, scores:[...]}
    Note over WS: 广播

    AWD->>Redis: 释放轮次锁
```

### 7.3 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 轮次幂等 | Redis SETNX `awd:round:lock:{cid}:{N}`，TTL = 轮次间隔 × 2 | 防止定时器抖动导致同一轮次重复执行 |
| Flag 注入方式 | Docker API `CopyToContainer` 写入 `/flag/flag.txt` | 避免 `sh -c` 命令拼接注入风险；无需重启容器，实时生效；权限由容器内 entrypoint 控制 |
| Flag 存储 | Redis Hash `awd:flags:{cid}:{round}` | 验证时 O(1) 查找，轮次结束后可批量清理 |
| 服务存活检测 | 优先 HTTP GET `/health`，降级为 TCP dial | HTTP 能检测应用层健康，TCP 只能检测端口存活 |
| 检测超时 | 5s | 过短会误判网络抖动，过长会拖慢轮次执行 |
| 并发度控制 | Flag 注入和存活检测均限制 10 并发 goroutine | 避免瞬间大量 Docker exec 和网络请求 |

### 7.4 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| Flag 注入失败（CopyToContainer 超时） | 标记该队本轮 Flag 注入失败，跳过该队的攻击得分收集；记录告警 |
| 服务存活检测超时 | 标记为宕机（is_alive=false），计 0 分防守分 |
| 轮次执行超时（超过轮次间隔的 80%） | 强制终止当前轮次，标记为 `timeout`，已完成的步骤结果保留 |
| 数据库事务失败 | 回滚本轮所有得分变更，标记轮次为 `failed`，管理员可手动重跑 |
| Redis 排行榜更新失败 | 数据库已提交的得分不回滚，标记需要补偿，后台任务从数据库重建排行榜 |

### 7.5 并发与一致性考虑

- **轮次幂等执行**：Redis SETNX 锁以 `{contest_id}:{round_number}` 为粒度，保证同一轮次全局只执行一次。TTL 设为轮次间隔的 2 倍，兜底防死锁。
- **Flag 注入与提交的竞态**：新轮次 Flag 注入期间存在时间窗口，部分队伍已注入新 Flag、部分尚未注入。解决方案：先将新轮次 Flag 写入 Redis Hash（`awd:flags:{cid}:{N}`），再逐个注入容器；Flag 验证时同时检查当前轮次和上一轮次的 Flag（`round=N` 和 `round=N-1`），在注入全部完成后（标记 `awd:round:{cid}:{N}:injected=true`）才停止接受上一轮 Flag。这样即使注入过程中有提交，也不会因为新旧 Flag 混合而误判。
- **多实例部署**：轮次定时器在所有实例上运行，但 Redis 锁保证只有一个实例执行。其他实例获取锁失败后直接跳过。
- **轮次间隔与执行时间**：轮次执行时间必须小于轮次间隔的 80%，否则触发超时保护。建议轮次间隔 >= 3 分钟，给执行留足余量。

---

## 8. 作弊检测流程

### 8.1 流程描述

作弊检测主要针对**动态 Flag 共享**场景：不同队伍在短时间窗口内提交了相同的动态 Flag，
说明 Flag 可能通过非正常渠道传播。系统自动标记可疑记录并通知管理员，**不自动判罚**，
由管理员人工审核后决定处理方式。

### 8.2 检测触发点

作弊检测嵌入在 Flag 提交流程中，每次正确提交后同步执行检测逻辑。

### 8.3 时序图

```mermaid
sequenceDiagram
    participant SS as SubmitService
    participant CD as CheatDetector
    participant Redis
    participant PG as PostgreSQL
    participant 通知服务

    SS->>CD: Flag验证通过, 触发检测 {submitted_flag, team_id, challenge_id, submit_time}

    CD->>Redis: HGET ctf:cheat:flag:{contest_id}:{challenge_id} {flag_hash} 查询该Flag首次提交记录
    Redis-->>CD: first_submit (可能为nil)

    Note over CD: first_submit == nil: 首次提交

    CD->>Redis: HSET ctf:cheat:flag:{contest_id}:{challenge_id} {flag_hash} {team_id: submit_time} 记录首次提交
    Redis-->>CD: OK

    CD-->>SS: 无异常, 返回
```

当同一 Flag 被不同队伍提交时，触发可疑检测：

```mermaid
sequenceDiagram
    participant SS as SubmitService
    participant CD as CheatDetector
    participant Redis
    participant PG as PostgreSQL
    participant 通知服务

    SS->>CD: Flag验证通过, 触发检测

    CD->>Redis: HGET ctf:cheat:flag:{contest_id}:{challenge_id} {flag_hash}
    Redis-->>CD: first_submit {team_A, t1}

    Note over CD: first_submit != nil && team_id != first_team && |t2-t1| <= 30s

    rect rgb(255, 220, 220)
        Note over CD, 通知服务: 触发作弊告警
        CD->>PG: INSERT cheat_reports (contest_id, type='flag_sharing', status='pending', suspect_user_ids, evidence)
        PG-->>CD: OK
        CD->>通知服务: 通知管理员 {type:"cheat_alert", challenge, team_a, team_b, time_diff}
    end

    CD-->>SS: 正常返回(不阻塞)
```

### 8.4 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 时间窗口 | 30s | 动态 Flag 每个用户唯一，30s 内不同队伍提交相同 Flag 极大概率是共享行为 |
| Flag 索引 | Redis Hash，key=`ctf:cheat:flag:{contest_id}:{challenge_id}`，field=`SHA256(submitted_flag)` | 用 hash 而非明文存储，防止 Redis 泄露后暴露 Flag |
| 检测时机 | Flag 验证通过后同步执行 | 实时检测，不遗漏；检测逻辑轻量（1 次 Redis 读 + 条件判断），不影响主流程性能 |
| 判罚策略 | 不自动判罚，仅标记 + 通知管理员 | 避免误判（如同一队伍多设备登录），人工审核更可靠 |
| 告警通道 | WebSocket 推送到管理员面板 + 站内消息 | 实时感知，管理员可在竞赛进行中处理 |

### 8.5 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| Redis 查询失败 | 跳过本次检测，不阻塞 Flag 提交主流程；记录错误日志 |
| 数据库写入 cheat_reports 失败 | 记录错误日志，不影响主流程；后台补偿任务定期扫描 Redis 中的可疑记录 |
| 通知发送失败 | 记录到数据库即可，管理员可在面板中查看未处理的告警 |
| 时间窗口边界误判 | 管理员审核时可查看完整提交日志、IP 地址、提交间隔等辅助信息 |

### 8.6 并发与一致性考虑

- **并发提交同一 Flag**：两个队伍几乎同时提交同一 Flag 时，Redis HGET/HSET 不是原子的，可能两个都读到 nil。解决方案：使用 `HSETNX`（仅当 field 不存在时设置），只有第一个成功的才是"首次提交"，第二个 HSETNX 返回 0 后再 HGET 获取首次提交信息进行比对。
- **检测不阻塞主流程**：作弊检测的数据库写入和通知发送可以异步执行（goroutine），只有 Redis 查询是同步的（< 1ms）。
- **数据清理**：竞赛结束后，`ctf:cheat:flag:{contest_id}:*` 相关的 Redis key 由竞赛清理任务统一删除。

---

## 9. 竞赛高峰排队机制

### 9.1 流程描述

竞赛开始瞬间，大量用户同时请求启动靶机实例，最容易被打满的是 Docker 冷启动和宿主机资源，而不是 Go HTTP 协程。

当前实现采用**数据库持久化排队 + 后台调度器**：

- HTTP 线程只写入 `instances(status=pending)` 并立即返回；
- 后台调度器周期性扫描最早的 `pending` 实例；
- 调度器在推进前先检查全局启动并发上限和全局活跃实例上限；
- 只有拿到容量配额的实例才会被原子推进到 `creating` 并真正触发 Docker 编排。

### 9.2 当前实现的数据结构

```
instances.status:
  pending  -> 已占位，等待后台调度
  creating -> 已被某个 worker 抢占，正在创建容器
  running  -> 创建成功，用户可访问
  failed   -> 启动失败，端口已释放

container.scheduler:
  enabled                -> 是否启用异步启动流水线
  poll_interval          -> worker 扫描 pending 的周期
  batch_size             -> 单轮最多领取多少个 pending 实例
  max_concurrent_starts  -> 全局同时允许多少个 creating
  max_active_instances   -> 全局 creating + running 的硬上限
```

### 9.3 时序图 — 请求入队与后台消费

```mermaid
sequenceDiagram
    participant 用户
    participant API as API Gateway
    participant PS as PracticeService
    participant PG as PostgreSQL
    participant SCH as ProvisioningScheduler
    participant Docker

    用户->>API: 启动靶机
    API->>PS: 鉴权+限流
    PS->>PG: 事务内锁用户/队伍 + 写入 pending 实例
    PG-->>PS: instance_id
    PS-->>API: {instance_id, status=pending}
    API-->>用户: 已排队

    loop poll_interval
        SCH->>PG: count(creating), count(creating+running)
        alt 仍有容量
            SCH->>PG: 查询最早 pending 实例
            SCH->>PG: 原子更新 pending -> creating
            SCH->>Docker: 创建网络/容器并启动
            alt 成功
                SCH->>PG: UPDATE running + access_url
            else 失败
                SCH->>PG: UPDATE failed + 释放端口
            end
        else 容量不足
            Note over SCH: 保持 pending，等待下一轮
        end
    end
```

### 9.4 关键决策点

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 队列持久化 | 直接复用 `instances` 表的 `pending` 状态 | 不引入第二套排队事实源，进程重启后天然可恢复 |
| 抢占方式 | 查询 oldest pending + `UPDATE ... WHERE status='pending'` | 多后端实例下只有一个 worker 会成功领取任务 |
| 启动并发度 | `container.scheduler.max_concurrent_starts` | 把 Docker 冷启动并发限制从 API 层剥离成可配置阈值 |
| 宿主机总容量 | `container.scheduler.max_active_instances` | 让宿主机容量有硬上限，不再默认单机无限接单 |
| 前端状态感知 | 直接复用实例查询接口读取 `pending/creating/running/failed` | 先落地最小闭环，不额外引入 request_id/位置接口 |

### 9.5 异常处理

| 异常场景 | 处理策略 |
|----------|----------|
| 容量已满 | 新实例保持 `pending`，等待下一轮调度 |
| worker 宕机/进程重启 | 未领取的实例仍在 `pending`；`creating` 失败实例由补偿和清理任务收敛 |
| 实例创建失败 | 标记 status=failed，前端轮询到后提示用户重试 |
| 调度关闭 | 回退为同步启动路径，但保留同一套实例状态机 |

### 9.6 并发与一致性考虑

- **入队幂等**：用户或队伍作用域先加数据库锁，再检查 `pending/creating/running` 实例，所以重复点击不会产生多个排队实例。
- **多 worker 互斥**：`TryTransitionStatusWithContext(id, pending, creating)` 是真正的抢占点，只有一个后端实例会成功。
- **容量判断**：当前实现把 `creating` 视为正在消耗启动并发，把 `creating + running` 视为宿主机活跃容量；`pending` 只占用户配额，不占宿主机容量。
- **后续演进**：如果未来需要排队位置、超时取消、可重投递，可在保持 `pending` 状态机不变的前提下，把调度事实源从数据库扫描演进为 Redis Stream / MQ。

### 9.7 监控指标

| 指标 | 采集方式 | 告警阈值 |
|------|----------|----------|
| pending 数量 | `COUNT(status='pending')` | 持续增长且 5 分钟不下降 |
| creating 数量 | `COUNT(status='creating')` | 长时间等于 `max_concurrent_starts` |
| 活跃实例数 | `COUNT(status IN ('creating','running'))` | 接近 `max_active_instances` |
| 平均启动耗时 | `running.updated_at - pending.created_at` | 显著高于基线 |
| 创建失败率 | failed / (completed + failed) | > 10% |

---

## 附录：流程间依赖关系

```mermaid
flowchart TD
    F6["竞赛状态流转 (6)<br/>状态机驱动一切"]

    F1["靶机启动 (1)"]
    F2["Flag提交 (2)"]
    F7["AWD轮次 (7)"]
    F3["动态计分 (3)"]
    F4["排行榜 (4)"]
    F5["容器回收 (5)"]
    F8["作弊检测 (8)"]
    F9["高峰排队 (9)<br/>靶机启动的前置缓冲层"]

    F6 --> F1
    F6 --> F2
    F6 --> F7

    F2 --> F3
    F2 --> F8

    F3 --> F4
    F1 --> F4
    F7 --> F4

    F1 --> F5

    F9 -.->|前置缓冲| F1
```

> 说明：
> - 流程 6（竞赛状态流转）是全局控制器，决定其他流程是否可执行
> - 流程 9（排队机制）是流程 1（靶机启动）的前置缓冲层，高峰期自动启用
> - 流程 2（Flag 提交）触发流程 3（动态计分）和流程 8（作弊检测）
> - 流程 1/2/3/7 的结果最终汇聚到流程 4（排行榜更新）
> - 流程 5（容器回收）独立运行，清理流程 1 创建的资源
