# Runtime / Instance 边界 Phase 2 Slice 7 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/module/runtime/ports/container_runtime.go`、`code/backend/internal/module/runtime/application/*` 中与 container runtime ports 相关的 service、`code/backend/internal/module/runtime/runtime/module.go`、`docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice7-implementation-plan.md`、`docs/design/backend-module-boundary-target.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural refactor`，本轮目标是把剩余 Docker / ACL / 文件等 capability 收口成显式 container runtime ports，并明确 compat wrapper 的仓库内保留决策
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with minor issues`

当前改动把 `runtime/application` 和 `runtime/runtime.Module` 中仍分散声明的 container runtime capability 收口到了 `runtime/ports/container_runtime.go`，同时把 `runtime/application/*` compat wrapper 的状态明确为“仓库内生产调用已迁空，但删除动作等待确认”。

## Findings

### 无 material findings

- `runtime/application/commands/{provisioning_service,runtime_cleanup_service}.go` 不再各自声明一组本地 engine-ish 接口，而是直接依赖 `runtime/ports` 下的 container runtime ports；行为路径没有变化。
- `runtime/application/{container_file_service,image_runtime_service,container_stats_service}.go` 也改成复用同一组 port 定义，容器文件、镜像、指标读取能力的边界表达更一致。
- `runtime/runtime/module.go` 的 `Engine` 现在由这些 capability port 组合出来，继续保持现有 wiring 行为，但不再把宽接口含义藏在一个本地类型里。
- 文档已经同步写明：仓库内非测试生产调用不再依赖 `runtime/application/*` compat wrapper；当前保留只是为了等待是否还需要仓库外兼容入口的最终确认。

### Minor issue

- compat wrapper 文件还在；这不是本轮 correctness 问题，但下一刀需要在用户确认后直接删除，而不是继续长期保留。
- `internal/app/composition/runtime_adapter_compat.go` 与 `internal/module/runtime/runtime/adapters.go` 的重复适配逻辑仍在，本轮没有扩大到处理这组重复实现。

## Senior Implementation Assessment

这轮的收口方向是对的：

- 先把 capability 边界显式写进 `runtime/ports`，再让 application / module wiring 统一依赖这组 port，比继续在每个 service 文件里私有声明接口更接近目标架构。
- compat wrapper 没有被错误地“顺手修回去”；仓库内既然已经没有生产调用，就不应再让 repo / config / engine 级逻辑回流到这层。
- 当前留下的真正未决项只剩“要不要删 wrapper 文件”以及更后面的“重复 adapter 是否继续收口”，而不是 owner 职责不清。

## Validation

```bash
cd code/backend && go test -timeout 3m ./internal/module/runtime/application/... ./internal/module/runtime/runtime
cd code/backend && go test -timeout 5m ./internal/module/instance/... ./internal/app/composition ./internal/app/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

## Residual Risk

- 当前 review 仍是同会话自审，不是独立 reviewer gate。
- wrapper 文件尚未删除；如果仓库外其实也没有兼容需求，这层文件继续存在只会增加维护噪音。
- 重复的 runtime HTTP adapter 逻辑还没有进入本轮 slice，后续如果继续触达这块，需要单独立计划收口。

## Touched Known-Debt Status

- 已触达的已知结构债：`runtime` 物理模块里仍混着 container capability 接口声明，边界表达分散；`runtime/application/*` compat wrapper 是否还需要继续保留未决。
- 本轮已完成的收口：container runtime capability 统一落到 `runtime/ports/container_runtime.go`；compat wrapper 的仓库内保留理由已经收缩到“只剩待确认删除”。
- 本轮剩余但已明确边界的部分：是否删除 wrapper 文件，需要用户确认后执行；重复 runtime HTTP adapter 不在本轮范围内。
