# Runtime / Instance 边界 Phase 2 Slice 8 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/module/runtime/application/{commands/instance_service.go,commands/runtime_maintenance_service.go,queries/instance_service.go,queries/proxy_ticket_service.go}`、`code/backend/internal/module/runtime/application/{instance_service_test.go,proxy_ticket_service_test.go,commands/runtime_maintenance_service_test.go}`、`code/backend/internal/module/architecture_allowlist_test.go`、`docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice8-implementation-plan.md`、`docs/design/backend-module-boundary-target.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural cleanup`，本轮目标是在用户已确认删除 compat import path 后，彻底移除 `runtime/application/*` 中剩余的 instance / proxy ticket / maintenance thin wrapper
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with limitations`

当前改动已经把 `runtime/application/*` 里剩余的 compat wrapper 文件删除，并同步收缩了测试、allowlist 和当前架构事实。仓库内实例命令、实例查询、proxy ticket、maintenance 的唯一 owner 现在都固定在 `instance`。

## Findings

### 无 material findings

- `internal/app/composition/instance_module.go` 之前已经直接装配 `instancecmd.NewInstanceService`、`instanceqry.NewInstanceService`、`instanceqry.NewProxyTicketService`、`instancecmd.NewInstanceMaintenanceService`；本轮删除 4 个 wrapper 没有切断当前生产 wiring。
- `code/backend/internal/module/architecture_allowlist_test.go` 已移除对应白名单路径，guardrail 不再假设这些 compat 文件仍然存在。
- `code/backend/internal/app/router_test.go` 中保留的 `runtimecmd.NewInstanceService(`、`runtimeqry.NewProxyTicketService(` 等字符串只是负向断言 marker，用来禁止 runtime wiring 回流到旧 owner；它们不是运行时依赖，不需要随文件删除一起移除。
- `runtime/application` 中仍保留的 `provisioning_service`、`runtime_cleanup_service`、`container_file_service`、`image_runtime_service`、`container_stats_service` 继续代表 container capability owner；这轮删除没有误伤这些真实 runtime service。

### Minor issue

- 当前 review 仍然是同会话自审，缺少独立 reviewer gate。

## Senior Implementation Assessment

这轮删除是合理的收口：

- compat path 已经没有仓库内生产调用，再继续保留只会让 `runtime/application` 看起来还承担实例业务入口。
- 删除动作不是只改文档口径，而是连同测试和 allowlist 一起收尾，当前事实和代码保持一致。
- `router_test` 保留负向 marker 也对，它继续防止未来有人把 `instance` owner 的构造逻辑重新塞回 `runtime`。

## Validation

```bash
cd code/backend && go test ./internal/module/runtime/application/...
cd code/backend && go test -timeout 3m ./internal/module/runtime/application/... ./internal/module/runtime/runtime
cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/composition ./internal/app/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

以上命令已在当前工作区执行，通过。

## Residual Risk

- 当前 review 不是独立 reviewer gate。
- `internal/app/composition/runtime_adapter_compat.go` 与 `internal/module/runtime/runtime/adapters.go` 的重复适配逻辑仍在，本轮没有扩大到处理这组重复实现。

## Touched Known-Debt Status

- 已触达的已知结构债：`runtime/application/*` 对实例业务的 compat import path 仍残留物理文件。
- 本轮已完成的收口：4 个 compat wrapper 已删除，测试 / allowlist / 当前事实已同步到新的 owner 边界。
- 本轮剩余但不在范围内的部分：runtime HTTP adapter 的重复实现仍需后续单独切片处理。
