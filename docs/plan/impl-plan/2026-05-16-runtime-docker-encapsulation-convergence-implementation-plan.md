# Runtime Docker Encapsulation Convergence Implementation Plan

## Objective

把 runtime 底层 Docker 封装继续收口到更稳定的模块边界，解决这轮 review 暴露出的 4 个问题：

- `engine.go` 责任堆积过重，文件级 owner 混杂
- `CreateContainer()` 会原地修改调用方传入的 `*model.ContainerConfig`
- Docker not-found / unavailable / port-conflict 语义还没有完全收口成 typed runtime error
- `runtime.Module` 已经构建 core service，但 `instance` composition 还在重复 new 同一批 service

## Non-goals

- 不替换 Docker SDK，也不引入第二种 runtime provider
- 不改 `runtime/ports` 公开能力面，不做新的 API / DTO / 数据库契约变更
- 不重写 provisioning / cleanup 主流程，只在保持行为等价的前提下收口 owner 和错误语义
- 不处理 AWD runtime hardening 之外的其他工作区改动

## Inputs

- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `code/backend/internal/module/runtime/infrastructure/runtime_metrics.go`
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/module/runtime/runtime/adapters.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/instance_module.go`
- `code/backend/internal/module/runtime/ports/errors.go`
- `docs/reviews/backend/2026-05-16-awd-runtime-hardening-review.md`

## Brainstorming Summary

候选方向：

1. 保持现状，只补几处注释和测试
   - 拒绝：不能解决输入对象被 mutation、typed error 漂移和 module owner 分裂
2. 新建第二层 runtime façade，把现有 `Engine` 整体包起来
   - 拒绝：只会再叠一层壳，实际 owner 还是没变清楚
3. 在保持 `Engine` struct 与 `ports` 接口不变的前提下，按能力拆文件、补 typed error、回收 module wiring
   - 采用：这是最小可审查改动，同时能把 touched surface 的结构债一起收口

## Chosen Direction

- `runtime/ports/errors.go`
  - 新增 runtime unavailable / container not found / network not found 等 typed error
  - Docker daemon message 解析只允许留在 `runtime/infrastructure`
- `runtime/infrastructure`
  - 把 `engine.go` 拆成 provisioning / file access / inventory / error helpers 多文件
  - `CreateContainer()` 改为生成局部 resolved config，不再回写调用方 `cfg`
  - `RemoveContainer()` / `RemoveNetwork()` / `InspectManagedContainer()` 统一做 not-found 归一化
- `runtime/application/commands`
  - cleanup / provisioning 改成消费 typed runtime error，不再自行做字符串匹配
- `runtime/runtime/module.go` + `app/composition/instance_module.go`
  - runtime module 暴露已构建的 `ProvisioningService` / `RuntimeCleanupService`
  - instance module 直接复用，不再重复构造

## Ownership Boundary

- `runtime/ports`
  - 负责：上层可分支处理的 runtime typed error 语义
  - 不负责：识别 Docker SDK 各种 message 细节
- `runtime/infrastructure`
  - 负责：Docker SDK concrete 调用、daemon error 识别、container config default merge
  - 不负责：业务层字符串分支、实例生命周期规则
- `runtime/application/commands`
  - 负责：根据 typed runtime error 决定 cleanup / provisioning 分支
  - 不负责：理解 Docker daemon 原始 message
- `runtime/runtime` + `app/composition`
  - 负责：模块 owner 和装配复用
  - 不负责：重复 new 同一 runtime application service

## Change Surface

- Add: `.harness/reuse-decisions/runtime-docker-encapsulation-convergence.md`
- Add: `docs/plan/impl-plan/2026-05-16-runtime-docker-encapsulation-convergence-implementation-plan.md`
- Add: `code/backend/internal/module/runtime/infrastructure/engine_errors.go`
- Add: `code/backend/internal/module/runtime/infrastructure/engine_provisioning.go`
- Add: `code/backend/internal/module/runtime/infrastructure/engine_files.go`
- Add: `code/backend/internal/module/runtime/infrastructure/engine_inventory.go`
- Add: `code/backend/internal/module/runtime/runtime/module_test.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/engine.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/engine_test.go`
- Modify: `code/backend/internal/module/runtime/infrastructure/engine_error_test.go`
- Modify: `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
- Modify: `code/backend/internal/module/runtime/application/commands/runtime_cleanup_service.go`
- Modify: `code/backend/internal/module/runtime/runtime/module.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters.go`
- Modify: `code/backend/internal/module/runtime/runtime/adapters_test.go`
- Modify: `code/backend/internal/module/runtime/ports/errors.go`
- Modify: `code/backend/internal/module/runtime/service_test.go`
- Modify: `code/backend/internal/app/composition/instance_module.go`

## Task Slices

- [ ] Slice 1: typed runtime error 收口
  - Goal
    - unavailable / container not found / network not found / published port conflict 全部由 `runtime/ports` 暴露，上层不再解析 Docker message
  - Validation
    - `cd code/backend && go test ./internal/module/runtime/infrastructure -run 'TestNormalize.*|TestWrap.*' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/runtime -run 'TestService(CleanupRuntime|RemoveContainer|CreateContainer).*' -count=1 -timeout 5m`
  - Review focus
    - 是否仍保留 unwrap 链
    - 是否把字符串匹配完全压回 infrastructure

- [ ] Slice 2: `Engine` 输入与职责收口
  - Goal
    - `CreateContainer()` 不再 mutation 调用方 config
    - `engine.go` 按能力拆分，维持外部接口稳定
  - Validation
    - `cd code/backend && go test ./internal/module/runtime/infrastructure -run 'TestResolveContainer|TestDefaultSecurityConfig|TestSelectServicePort' -count=1 -timeout 5m`
  - Review focus
    - 默认资源 / 安全配置是否仍然等价
    - 新 helper 是否只在 infrastructure 内部复用，没有把 owner 再往上抬

- [ ] Slice 3: runtime module wiring 收口
  - Goal
    - `runtime.Module` 成为 core runtime service 的真实 owner
    - `BuildInstanceModule()` 不再重复 new provisioning / cleanup service
  - Validation
    - `cd code/backend && go test ./internal/module/runtime/runtime ./internal/app/composition -count=1 -timeout 5m`
  - Review focus
    - 是否没有改坏现有 runtime access host 接线
    - 是否没有让 composition 重新依赖 runtime infrastructure 细节

## Risks

- typed error 归一化如果漏掉某条旧分支，cleanup 可能把“应忽略的 not found”重新当成失败
- 拆文件时如果 import 或 helper owner 放错，容易破坏 runtime 分层架构测试
- module wiring 如果没有兼顾当前 `access_host` 改动，容易把最近的 runtime access host 行为回退掉

## Verification Plan

1. `cd code/backend && go test ./internal/module/runtime/infrastructure -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/runtime/runtime -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/runtime -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/app/composition -count=1 -timeout 5m`

## Rollback / Recovery Notes

- 如果 typed error 归一化引入误判，先回退 `runtime/ports/errors.go` 与 infrastructure normalization helpers，不影响数据库契约
- 如果 module wiring 收口引发 composition 回归，可回退到原来的重复构造方式，运行时行为会恢复，但结构债也会回来

## Architecture-Fit Evaluation

- Docker concrete 仍然只停留在 `runtime/infrastructure`
- runtime typed error 继续由 `runtime/ports` 对上暴露，没有把 Docker message 带进 application
- module owner 收口在 `runtime/runtime`，没有新建横切共享包或额外 façade
