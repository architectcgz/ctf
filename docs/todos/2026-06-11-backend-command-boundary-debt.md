# 后端 application commands 边界清理待办

> 来源：2026-06-11 后端 `application/commands` 边界检查
> 状态：Open

本文记录 `internal/module/*/application/commands` 中仍把过宽 use case、基础设施实现或运行时细节放在 commands 包下的现象。已收口的项应在这里勾选并补一句处理位置。

## P1：明确应拆

- [x] `practice/application/commands` 的 `Service` 仍是过宽 facade：同一个 service 横跨 Flag 提交、实例启动、后台 provisioning、AWD desired reconcile、AWD scope control、人工评阅和竞赛 AWD 编排。应按 use case 拆成独立 application service / package，并让 runtime 显式装配。
  - 依据：`code/backend/internal/module/practice/application/commands/service.go`、`instance_start_service.go`、`submission_service.go`、`manual_review_service.go`、`awd_desired_runtime_reconciler.go`
  - 已处理：移除对外 `Service/NewService` 兼容入口，新增 `CommandServices`、`InstanceLifecycleService`、`SubmissionService`、`ManualReviewService`、`RuntimeLifecycleService`，HTTP handler/runtime 改为显式注入聚焦 service；DTO 共享面落到 `practice/contracts`。

- [x] `challenge/application/commands` 仍残留具体基础设施实现：Docker CLI image builder、artifact GC 的 LocalFS 遍历/删除、AWD checker artifact 的 LocalFS 持久化和环境变量读取。应移动到 infrastructure adapter，经 `challenge/ports` 注入 application service。
  - 依据：`code/backend/internal/module/challenge/application/commands/docker_image_builder.go`、`artifact_gc_service.go`、`awd_challenge_import_service.go`
  - 已处理：Docker CLI image builder、artifact GC 与 AWD checker artifact store 均移到 `code/backend/internal/module/challenge/infrastructure/`，AWD 导入经 `challenge/ports.AWDCheckerArtifactStore` 注入。

- [x] `assessment/application/commands` 报表输出边界仍在 commands：`ReportService` 负责 storage dir 创建、安全路径检查、文件存在性检查和下载路径组装。应抽出 report output/file store port，由 infrastructure LocalFS 实现。
  - 依据：`code/backend/internal/module/assessment/application/commands/report_file_output.go`、`report_service.go`
  - 已处理：新增 `assessment/ports.ReportOutputStore` 与 `assessment/infrastructure/report_output_store.go`，commands 只消费 port。

- [x] `instance/application/commands/startup_runtime_recovery_service.go` 直接依赖 `internal/infrastructure/redislock` 具体类型，并直接读取宿主机 boot id 文件。应改为 ports 层 lease / boot id reader 抽象，由 infrastructure 实现 Redis lock 和 `/proc/sys/kernel/random/boot_id` 读取。
  - 依据：`code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`、`code/backend/internal/module/instance/infrastructure/platform_runtime_state_store.go`
  - 已处理：新增 `StartupRuntimeStateStore`、`StartupRecoveryLockLease`、`HostBootIDReader` ports；Redis lock wrapper 留在 `platform_runtime_state_store.go`，boot id 文件读取落到 `instance/infrastructure/boot_id_reader.go`。

## P2：可后置观察

- [ ] `contest/application/commands/AWDService` 覆盖 round、checker、preview、attack submit、attack log 等多个 AWD 子用例。当前基本通过 ports 依赖，暂不按放错层处理；后续若继续膨胀，应按 AWD 子用例拆 application service。
  - 依据：`code/backend/internal/module/contest/application/commands/awd_service.go`、`awd_service_run_commands.go`、`awd_attack_submit_commands.go`

- [ ] `container_runtime/application/commands/provisioning_service.go` 文件较大，但职责集中在运行时资源 provisioning，且通过 `runtimeports.ContainerProvisioningRuntime` 依赖运行时。暂不按 commands 边界问题处理，只作为可读性和文件拆分观察项。
  - 依据：`code/backend/internal/module/container_runtime/application/commands/provisioning_service.go`
