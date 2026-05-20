# Practice Instance Readiness Probe Phase 5 Slice 17 Implementation Plan

## Objective

继续 phase 5 收窄 `practice` application concrete allowlist，把实例 access URL 的 HTTP/TCP 探活细节下沉到模块 port / infrastructure adapter：

- 去掉 `practice/application/commands/instance_provisioning.go -> net/http`
- 保持 provisioning 的重试次数、间隔、失败标记和 AWD stable alias 跳过宿主机探活行为不变

## Non-goals

- 不处理 `contest/application/jobs/* -> net/http` 的 AWD checker / probe concrete
- 不改 `runtimeService` 的容器创建 contract、实例状态机或 `CreateTopology` 语义
- 不改 practice 的外部 HTTP API、DTO 或实例访问 URL 生成规则

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/ports/ports.go`

## Current Baseline

- `practice/application/commands/instance_provisioning.go` 当前直接持有 `http.Client`、`http.NewRequestWithContext`、`url.Parse`、`net.Dialer`
- provisioning application service 同时负责重试编排和协议级探测
- `practice/runtime/module.go` 当前没有专门的 readiness probe wiring
- allowlist 当前保留：
  - `practice/application/commands/instance_provisioning.go -> net/http`

## Chosen Direction

把实例 readiness 探测表达成 practice 自己的窄 probe port：

1. 在 `practice/ports` 新增 `PracticeInstanceReadinessProbe`
2. `instance_provisioning.go` 保留 access URL 非空校验、重试次数 / 间隔控制和失败处理，只通过 probe port 执行单次探测
3. `practice/infrastructure` 提供 HTTP/TCP probe adapter，统一承接 `url.Parse`、HTTP GET、TCP dial 和响应体回收细节
4. `practice/runtime/module.go` 统一构建 probe，并注入 command service；相关测试在真正走 provisioning 路径时显式设置 probe

## Ownership Boundary

- `practice/application/commands/instance_provisioning.go`
  - 负责：实例 provisioning 编排、重试窗口、失败标记和 AWD stable alias 跳过探活
  - 不负责：知道 `http.Client`、`http.NewRequestWithContext`、`url.Parse`、`net.Dialer` 或响应体回收细节
- `practice/infrastructure/instance_readiness_probe.go`
  - 负责：对 access URL 执行单次 HTTP/TCP 探测
  - 不负责：决定重试次数、重试间隔、实例失败后的补偿或日志口径
- `practice/runtime/module.go`
  - 负责：装配 readiness probe 并传给 command service
  - 不负责：把 `net/http` concrete 继续留在 practice application surface

## Change Surface

- Add: `.harness/reuse-decisions/practice-instance-readiness-probe-phase5-slice17.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-instance-readiness-probe-phase5-slice17-implementation-plan.md`
- Add: `code/backend/internal/module/practice/ports/instance_readiness_probe_context_contract_test.go`
- Add: `code/backend/internal/module/practice/infrastructure/instance_readiness_probe.go`
- Add: `code/backend/internal/module/practice/infrastructure/instance_readiness_probe_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- Modify: `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- Modify: `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [ ] Slice 1: 提取 readiness probe port 与 infrastructure adapter
  - 目标：practice provisioning application surface 不再直接持有 HTTP/TCP concrete，探活行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/practice/infrastructure -run 'InstanceReadinessProbe' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'ProvisionInstance|RunProvisioningLoop' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/ports -count=1 -timeout 5m`
  - Review focus：application 是否仍只保留重试编排；HTTP/TCP readiness 语义和 AWD alias skip 是否保持不变

- [ ] Slice 2: 删除 allowlist 并同步文档
  - 目标：删掉 `practice/application/commands/instance_provisioning.go -> net/http`，phase5 当前事实同步更新
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次实际收口的 allowlist；文档正确反映 practice 剩余 concrete 依赖范围

## Risks

- 如果 probe adapter 改变了 HTTP/TCP 成功判定，实例可能误判 ready 或 failed
- 如果 runtime 没有注入 probe，生产 provisioning 会退化成配置错误
- 如果测试 helper 没有同步设置 probe，现有 provisioning 测试会因“probe 未配置”失败

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/infrastructure -run 'InstanceReadinessProbe' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/application/commands -run 'ProvisionInstance|RunProvisioningLoop' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/ports -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`
8. `git diff --check`

## Architecture-Fit Evaluation

- owner 明确：provisioning application service 继续持有 readiness 重试编排，协议探测落回 infrastructure
- reuse point 明确：复用前两刀已经验证的 `application -> ports -> infrastructure -> runtime wiring` 模式，不引入新的宽 runtime abstraction
- 这刀同时解决行为与结构：保留实例探活语义，同时删除 practice application surface 的 `net/http` concrete 例外
