# Runtime / Instance 边界 Phase 2 Slice 6 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/module/runtime/application/{commands,queries}` 下的 instance / proxy ticket / maintenance compat service 及其相关测试、`docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice6-implementation-plan.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/design/backend-module-boundary-target.md`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural refactor`，本轮目标是把 `runtime/application/*` 中仍承载实例业务的 compat mirror 压成基于 `instance/contracts` 的薄 wrapper
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with minor issues`

当前改动把 `runtime/application/*` 中 instance / proxy ticket / maintenance 相关 compat service 从 duplicated implementation 压成了 contract-backed wrapper，生产 wiring 继续固定在 `instance/application/*`，owner 方向没有回流。

## Findings

### 无 material findings

- compat service 现在只委托 `internal/module/instance/contracts`，没有重新引入 `runtime -> instance/application` 的 concrete cross-module import。
- 原本覆盖实例业务行为的测试已经切到 `instancecmd` / `instanceqry`，compat 层只保留最小 wrapper 测试，不再把行为覆盖绑在旧 owner 上。
- `ProxyTicketService.MaxAge()` 仍保留兼容方法，但 TTL 秒数由 compat wrapper 自身持有，没有把 convenience API 反向塞回 owner contract。

### Minor issue

- compat import path 仍然保留；这已经不再是结构 blocker，但后续仍要决定是否继续支持这组 legacy import path。
- `runtime/application` 目录里的行为测试虽然已改为验证 `instance` owner，但物理位置还没迁出旧目录；这不影响当前 correctness，只是后续是否继续整理测试归属的取舍问题。

## Senior Implementation Assessment

这轮的收口方式是对的：

- 不再试图让 `runtime/application/*` 直接 import `instance/application/*`，因此不会违反 `code/backend/internal/app/architecture_rules_test.go`
- 也没有继续保留 repo / config / engine 级别的第二份实例业务实现
- compat 包保留下来的只是 import path 和少量兼容方法，owner 已经稳定地固定在 `instance/application/*`

## Required Re-validation

建议保留为本轮复验命令：

```bash
cd code/backend && go test ./internal/module/runtime/application/...
cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

## Residual Risk

- 当前 review 仍是同会话自审，不是独立 reviewer gate。
- 是否继续保留 `runtime/application/*` 这组 compat import path 还没有最终决策；如果后续确认仓库内外都不再需要，它们应当被删除而不是继续增长。

## Touched Known-Debt Status

- 已触达的已知结构债：`runtime/application/*` 中仍保留实例 owner 的 duplicated implementation。
- 本轮已完成的收口：compat mirror 已压成 `instance/contracts` wrapper，不再承载第二份实例业务实现。
- 本轮剩余但不再是同一债务本体的部分：compat import path 仍保留，后续只需决定保留还是删除。
