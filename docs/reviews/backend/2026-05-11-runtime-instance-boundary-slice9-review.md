# Runtime / Instance 边界 Phase 2 Slice 9 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/module/runtime/runtime/{adapters.go,adapters_test.go,module.go}`、`code/backend/internal/app/composition/runtime_module_test.go`、`docs/{design/backend-module-boundary-target.md,architecture/backend/07-modular-monolith-refactor.md,architecture/backend/03-container-architecture.md,architecture/features/AWD防守工作区与边界设计.md,plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice9-implementation-plan.md}`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural cleanup`，本轮目标是删除 runtime 物理模块里已经没有生产 wiring 的 HTTP adapter 平行副本，并把当前事实收口到 composition owner
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with limitations`

当前改动把 `runtime/runtime/adapters.go` 中已经失活的 `runtimeHTTPServiceAdapter` 和对应重复测试删除，生产使用的 runtime HTTP adapter 只保留在 `internal/app/composition/runtime_adapter_compat.go`。同时，当前架构文档不再继续引用已删除的 runtime application owner 路径。

## Findings

### 无 material findings

- `internal/app/composition/instance_module.go` 仍然是唯一构造 runtime HTTP adapter 的生产 wiring；删除 runtime 包里的平行副本没有切断 handler 依赖。
- `code/backend/internal/module/runtime/runtime/adapters.go` 删除的只是 dead HTTP adapter，`runtimePracticeServiceAdapter`、`runtimeChallengeServiceAdapter`、`runtimeOpsStatsProviderAdapter` 仍保留在原位，practice / challenge / ops 的复用面没有被误删。
- `code/backend/internal/app/composition/runtime_module_test.go` 已补上 runtime 包删掉后仍有价值的 SSH access / editable defense file 保存行为测试，因此不是简单删 test 让代码变“干净”。
- `docs/architecture/backend/03-container-architecture.md` 与 `docs/architecture/features/AWD防守工作区与边界设计.md` 已同步到新的 owner 路径，不再把 `runtime/application/queries/proxy_ticket_service.go`、`runtime/application/commands/runtime_maintenance_service.go` 或 runtime 包 dead HTTP adapter 写成当前事实。

### Minor issue

- 当前 review 仍然是同会话自审，缺少独立 reviewer gate。

## Senior Implementation Assessment

这轮的方向比“继续抽第三层共享 helper”更对：

- 这里的问题不是两份实现都还活着，而是生产 owner 已经迁走，runtime 包里还留着一份 dead parallel adapter。
- 既然 composition 已经是实际 owner，最小正确改动就是删除 dead code，把缺失测试补到活跃 owner 那侧，而不是再提一层共享抽象继续延长过渡结构。
- 文档也一起收了，这样当前事实、生产 wiring 和测试覆盖终于落到同一个 owner 面上。

## Validation

```bash
cd code/backend && go test -timeout 3m ./internal/module/runtime/runtime ./internal/app/composition
cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

以上命令已在当前工作区执行，通过。

## Residual Risk

- 当前 review 不是独立 reviewer gate。
- `runtime/ports/container_runtime.go` 的最终物理落点仍未决定；这轮只处理 dead HTTP adapter，不处理 container runtime 物理模块迁出问题。

## Touched Known-Debt Status

- 已触达的已知结构债：`internal/app/composition/runtime_adapter_compat.go` 与 `internal/module/runtime/runtime/adapters.go` 平行保留 runtime HTTP adapter。
- 本轮已完成的收口：runtime 包 dead HTTP adapter 与重复测试已删除，当前事实已收口到 composition owner。
- 本轮剩余但不在范围内的部分：container runtime capability ports 未来是否迁出 `runtime` 物理模块，需要后续单独切片评估。
