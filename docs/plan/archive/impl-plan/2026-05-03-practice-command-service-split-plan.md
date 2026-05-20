# Practice Command Service 拆分方案

## 背景

`code/backend/internal/module/practice/application/commands/service.go` 当前约 2200 行，对应测试文件 `service_test.go` 约 3600 行。这个文件现在同时承载实例启动、AWD 服务启动/重启、异步 provisioning 调度、容器创建、拓扑请求构建、Flag 提交、人工审核、赛事 scope 解析和异步事件等职责。

这类文件继续增长会带来几个实际问题：

- 修改 AWD 运行时规则时，需要同时理解普通练习实例、赛事实例、拓扑实例和调度链路。
- 测试很难按业务职责定位，新增用例会继续集中到 `service_test.go`。
- 私有 helper 的语义边界不清晰，例如 AWD 网络命名、拓扑构建、实例探活和 Flag 生成都混在同一个文件中。
- 后续如果直接抽新 service 或新 package，容易一次改动过大，影响当前已经比较复杂的启动链路。

第一轮拆分目标应当是“等价拆文件”，而不是重构业务模型。

## 当前职责分布

### 1. Service 装配与生命周期

当前位置：

- `Service` struct
- `NewService`
- `SetEventBus`
- `StartBackgroundTasks`
- `Close`
- `runAsyncTask`
- `publishWeakEvent`
- `triggerAssessmentUpdate`
- `triggerScoreUpdate`

主要依赖：

- `practiceCommandRepository`
- `practiceInstanceCommandRepository`
- `RuntimeInstanceService`
- `ScoreUpdater`
- `AssessmentService`
- `redis`
- `config`
- `logger`
- `eventBus`

这部分是命令服务的依赖装配和后台任务生命周期，不应和具体业务流程混在一起。

### 2. 实例启动入口

当前位置：

- `StartChallenge`
- `StartContestChallenge`
- `StartContestAWDService`
- `RestartContestAWDService`
- `StartAdminContestAWDTeamService`
- `startPersonalChallenge`
- `startChallengeWithScope`
- `requiresPublishedHostPort`
- `instanceRespForScope`

主要职责：

- 处理公开入口。
- 解析 scope。
- 判断复用实例或创建实例。
- 预留端口。
- 根据调度器开关决定同步启动还是进入 pending。

这里是实例启动主流程，后续应保持为启动 use case 的主入口。

### 3. AWD 操作记录与编排查询

当前位置：

- `recordAWDServiceOperation`
- `createAWDServiceOperation`
- `awdOperationStatusForInstanceStatus`
- `isFinishedAWDServiceOperationStatus`
- `restartCleanupRuntimeView`
- `GetContestAWDInstanceOrchestration`

主要职责：

- 记录 AWD 服务启动/重启操作。
- 查询管理员 AWD 实例编排视图。
- 将实例状态映射到 AWD operation 状态。

这部分属于 AWD 运维视角，不应和普通实例启动混放。

### 4. 异步实例启动调度

当前位置：

- `RunProvisioningLoop`
- `dispatchPendingInstances`
- `availableProvisioningSlots`
- `processPendingInstance`
- `schedulerEnabled`
- `schedulerPollInterval`
- `schedulerBatchSize`
- `schedulerMaxConcurrentStarts`
- `schedulerMaxActiveInstances`

主要职责：

- 扫描 pending 实例。
- 抢占 creating 状态。
- 控制并发启动数量和总活跃实例数量。
- 调用 provisioning 主流程。

这里和具体容器创建有调用关系，但职责上是调度器，不应放在容器创建文件里。

### 5. Provisioning 与探活

当前位置：

- `provisionInstance`
- `markInstanceFailed`
- `waitForInstanceReadiness`
- `probeInstanceAccessURL`
- `probeTCPAccessURL`
- `buildProvisioningFlag`
- `startProbeTimeout`
- `startProbeInterval`
- `startProbeAttempts`

主要职责：

- 为实例创建运行时资源。
- 探测访问地址是否就绪。
- 成功后回写 running 状态。
- 失败时清理运行时资源、释放端口、标记 AWD 操作失败。

这部分是实例生命周期中最关键的副作用边界。拆文件时只移动代码，不改变失败处理顺序。

### 6. 容器与拓扑创建

当前位置：

- `createContainer`
- `createSingleContainer`
- `normalizeChallengeTargetProtocol`
- `buildTopologyCreateRequest`
- `resolveAvailableImageRef`

主要职责：

- 单容器题目创建。
- 多容器拓扑题目创建。
- 将 `model.TopologySpec` 转为 `practiceports.TopologyCreateRequest`。
- 解析运行时镜像引用。

这里是 `practice` 模块和 `runtime` 模块的交界，应单独放置，便于后续继续收敛拓扑构建规则。

### 7. AWD 运行时规则

当前位置：

- `isAWDInstance`
- `buildAWDContestNetworkName`
- `buildAWDServiceAlias`
- `applyAWDStableNetworkToTopologyRequest`
- `appendUniqueString`
- `usesAWDStableNetworkAlias`
- `buildRuntimeContainerName`
- `resolveRuntimeChallengeName`
- `sanitizeRuntimeContainerSegment`

主要职责：

- 给 AWD 实例注入稳定 contest 网络。
- 给 AWD 服务注入稳定网络 alias。
- 给 AWD 容器生成稳定 Docker 名称。
- 判断是否跳过宿主机探活。

这部分应该从容器创建文件里独立出来，因为它是 AWD 运行时契约，后续还可能继续加 SSH、防守入口、checker 内网访问规则。

### 8. 赛事 scope 与运行时 subject 解析

当前位置：

- `resolveContestChallengeInstanceScope`
- `resolveContestAWDServiceInstanceScope`
- `resolveAdminContestAWDServiceInstanceScope`
- `resolveContestBaseInstanceScope`
- `resolveEffectiveInstanceScope`
- `loadRuntimeSubjectWithScope`
- `loadRuntimeSubjectForInstance`
- `loadContestAWDServiceRuntimeSubject`

主要职责：

- 校验比赛状态、报名状态、队伍归属。
- 将用户/赛事/team/service 解析为 `InstanceScope`。
- 按 scope 加载普通题目或 AWD 虚拟题目。

这部分是启动流程前置的身份和业务范围解析，应从 provisioning 和容器创建中分离。

### 9. AWD 虚拟题目构建

当前位置：

- `buildContestAWDServiceVirtualChallenge`
- `buildContestAWDServiceVirtualTopology`
- `parseContestAWDServiceSnapshotPoints`
- `parseContestAWDServiceSnapshotImageID`
- `parseContestAWDServiceSnapshotInstanceSharing`
- `parseContestAWDServiceSnapshotFlagType`
- `parseContestAWDServiceSnapshotFlagPrefix`
- `parseContestAWDServiceSnapshotInt`
- `firstRuntimeValue`

主要职责：

- 从赛事服务快照生成运行时使用的 `Challenge`。
- 从赛事服务快照生成运行时使用的 `ChallengeTopology`。

这部分不直接执行副作用，适合和 AWD scope/subject 解析放在相邻文件。

### 10. Flag 提交与人工审核

当前位置：

- `SubmitFlag`
- `applySolveGracePeriod`
- `formatSolveGracePeriod`
- `buildInstanceFlag`
- `validateSubmittedFlag`
- `ReviewManualReviewSubmission`
- `ListTeacherManualReviewSubmissions`
- `GetTeacherManualReviewSubmission`
- `ListMyChallengeSubmissions`
- `ensureTeacherCanAccessManualReviewSubmission`
- `normalizeTeacherManualReviewQuery`
- `ensureManualReviewRequesterRole`
- `ensureManualReviewDecisionStatus`
- `ensureManualReviewQuery`
- `manualReviewDetailRespFromRecord`
- `manualReviewListItemRespFromRecord`
- `challengeSubmissionRecordRespFromModel`

主要职责：

- 校验提交频率。
- 校验静态/动态/人工审核 Flag。
- 创建提交记录。
- 更新分数、进度缓存和事件。
- 教师/管理员处理人工审核。

这块和实例启动是不同 use case，应从 `service.go` 中移出。后续如果继续做模块边界收敛，Flag 提交可以成为独立 command service，但第一轮只拆文件。

## 推荐目标文件划分

第一轮保持同一个 Go package：`internal/module/practice/application/commands`。

不改外部 API，不改 `Service` 类型，不改 `NewService` 参数，不新建 package，不抽接口。

建议文件如下：

- `service.go`
  - 只保留 `Service` struct、依赖接口、`NewService`、`SetEventBus`。
- `service_lifecycle.go`
  - `StartBackgroundTasks`
  - `Close`
  - `runAsyncTask`
  - `publishWeakEvent`
  - `triggerAssessmentUpdate`
  - `triggerScoreUpdate`
- `instance_start_service.go`
  - `StartChallenge`
  - `StartContestChallenge`
  - `StartContestAWDService`
  - `RestartContestAWDService`
  - `StartAdminContestAWDTeamService`
  - `startPersonalChallenge`
  - `startChallengeWithScope`
  - `requiresPublishedHostPort`
  - `instanceRespForScope`
- `contest_awd_operations.go`
  - `recordAWDServiceOperation`
  - `createAWDServiceOperation`
  - `awdOperationStatusForInstanceStatus`
  - `isFinishedAWDServiceOperationStatus`
  - `restartCleanupRuntimeView`
  - `GetContestAWDInstanceOrchestration`
- `instance_provisioning_scheduler.go`
  - `RunProvisioningLoop`
  - `dispatchPendingInstances`
  - `availableProvisioningSlots`
  - `processPendingInstance`
  - scheduler config helpers
- `instance_provisioning.go`
  - `provisionInstance`
  - `markInstanceFailed`
  - readiness probe helpers
  - `buildProvisioningFlag`
- `runtime_container_create.go`
  - `createContainer`
  - `createSingleContainer`
  - `normalizeChallengeTargetProtocol`
  - `buildTopologyCreateRequest`
  - `resolveAvailableImageRef`
- `awd_runtime_rules.go`
  - AWD 网络名、alias、容器名、稳定访问判断等 helper
- `contest_instance_scope.go`
  - 赛事 scope 解析、运行时 subject 加载、effective scope 计算
- `contest_awd_runtime_subject.go`
  - AWD 虚拟 challenge/topology 构造与 snapshot 解析 helper
- `submission_service.go`
  - `SubmitFlag`
  - `applySolveGracePeriod`
  - `formatSolveGracePeriod`
  - `buildInstanceFlag`
  - `validateSubmittedFlag`
- `manual_review_service.go`
  - 人工审核相关 command、query normalization、DTO mapper helper

测试文件第一轮可以先不拆，避免同时移动实现和测试造成评审成本过高。第二轮再把 `service_test.go` 拆成：

- `instance_start_service_test.go`
- `instance_provisioning_test.go`
- `runtime_container_create_test.go`
- `awd_runtime_rules_test.go`
- `submission_service_test.go`
- `manual_review_service_test.go`

## 执行顺序

### 阶段 1：等价拆实现文件

目标：

- 只移动函数，不改函数签名。
- 不改变 `Service` 字段。
- 不改变外部调用方。
- 保证 `go test` 结果不变。

建议顺序：

1. 先拆无副作用 helper：
   - AWD 运行时规则
   - AWD 虚拟题目构建
   - manual review DTO mapper helper
2. 再拆独立 use case：
   - submission
   - manual review
   - contest scope
3. 最后拆启动链路：
   - container create
   - provisioning
   - scheduler
   - instance start
4. 收尾缩小 `service.go`，只保留构造和依赖定义。

验收：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/practice/application/commands
```

### 阶段 2：拆测试文件

目标：

- 测试按业务职责定位。
- 共享 stub 保留在现有 `repository_stub_test.go` 或新建 `service_test_support_test.go`。
- 不改测试语义，只移动用例。

建议顺序：

1. 先移动 AWD runtime 和 topology builder 的测试。
2. 再移动 provisioning / scheduler 测试。
3. 最后移动 submission / manual review 测试。

验收：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/practice/application/commands
```

### 阶段 3：评估是否拆 service 类型

这一阶段暂不建议直接做。只有当阶段 1、2 已稳定后，再评估是否需要进一步拆为多个 command owner：

- `InstanceStartService`
- `InstanceProvisioningService`
- `SubmissionService`
- `ManualReviewService`
- `ContestAWDOperationService`

如果要进入这一阶段，需要先确认 handler、app 装配、接口 contracts 和测试替身的影响面。否则会把简单的文件整理升级成跨模块重构。

## 行为边界

拆分过程中必须保持以下行为不变：

- 普通练习实例仍使用现有端口发布策略。
- AWD 实例仍不向宿主机发布入口端口。
- AWD 实例仍使用稳定 contest 网络和 service alias。
- AWD 容器命名仍使用 `ctf-instance-<challenge-name>-c<contest-id>-t<team-id>`。
- provisioning 失败时仍清理运行时资源、释放端口、标记实例 failed、结束 AWD operation。
- scheduler 开启时，实例仍先进入 pending，再由调度循环抢占为 creating。
- Flag 提交、人工审核、进度缓存清理、分数刷新和事件发布行为不变。

## 风险点

- Go 同包拆文件本身风险低，但 `service.go` 中函数互相引用很多，移动时容易漏 import 或留下未使用 import。
- `service_test.go` 现在覆盖了多个职责，第一轮不拆测试可以降低风险，但会保留测试文件过大的问题。
- `requiresPublishedHostPort` 当前固定返回 `true`，AWD 是否发布端口由 `createSingleContainer` / `CreateTopology` 的 `DisableEntryPortPublishing` 处理。拆文件时不要顺手改这个逻辑。
- `markInstanceFailed` 同时处理 runtime cleanup、实例状态、端口释放和 AWD operation 结束，是 provisioning 失败路径的关键副作用顺序，第一轮不要抽象。
- `buildInstanceFlag` 和 `buildProvisioningFlag` 名字相近但服务于不同阶段，一个用于创建实例时生成 nonce/flag，一个用于 pending 实例实际 provisioning 时重建 flag，移动时要保留二者语义。

## 最小验证集合

第一轮实现文件拆分后，至少运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/practice/application/commands
```

如果拆分时触碰 `practiceports`、`runtime` 或 handler 装配，再追加：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/practice/... ./internal/module/runtime/...
```

## 完成标准

- `service.go` 降到只保留依赖定义、构造和少量 service 级生命周期入口。
- 各业务职责文件名能直接表达 owner。
- 第一轮不改变任何公开 API 和行为。
- `go test ./internal/module/practice/application/commands` 通过。
- 文档中的目标文件划分和实际文件基本一致。
