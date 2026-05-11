# Runtime / Instance 边界 Phase 2 Slice 3 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/app/composition/instance_module.go`、`code/backend/internal/app/composition/runtime_adapter_compat.go`、`code/backend/internal/module/runtime/runtime/module.go`、`code/backend/internal/module/runtime/application/{commands,queries}` 中的 instance/proxy-ticket/maintenance 兼容面、相关 guardrail test 与架构文档
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural refactor`，本轮目标是把实例 owner 的生产 wiring 从 `runtime` 迁到 `instance`
- 审查上下文：本档为当前会话归档的 code review 证据；本次没有启用独立 reviewer agent，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with minor issues`

本轮没有发现新的 material correctness / regression blocker。实例 handler、proxy ticket、maintenance scheduler 的生产装配已经上移到 `composition.InstanceModule`，`runtime/runtime.Module` 也不再直连这些 instance owner use case。

## Findings

### 无 material findings

- 代码路径已经满足本轮结构目标：
  - `internal/module/instance/*` 成为实例命令、查询、proxy ticket、maintenance 的实际 owner
  - `internal/app/composition/instance_module.go` 直接装配 `instance` use case，并注册 `runtime_cleaner` 与 AWD defense SSH gateway
  - `internal/module/runtime/runtime/module.go` 收窄为 container-facing builder，不再生产装配实例 handler / cleaner
  - `internal/app/architecture_rules_test.go` 不再报 `runtime -> instance application` 的跨模块违规

### Minor issue

- `internal/module/runtime/application/*` 仍保留 compatibility mirror，并带着一份与 `instance` 同步的实例业务实现。当前它已经不再是生产 wiring owner，所以不阻塞本轮；但它仍然增加了后续维护成本，下一步应继续把剩余调用方和测试切到 `instance/*` 或 `InstanceModule`，再删除这层 mirror。

## Senior Implementation Assessment

这轮采用的实现方式是当前约束下风险最低的一条线：

- 不再尝试让 `runtime` 模块横向引用 `instance/application`，避免再次撞上 `internal/app/architecture_rules_test.go`
- 把实例 owner 的最终装配上移到 `app/composition`，让 `runtime` 只保留 container-facing builder
- 保留 `runtime/application/*` 兼容面，避免这轮把测试和调用方迁移面扩大到不可控

如果继续追求更彻底的收口，下一刀应该删掉 `runtime/application/*` 的 compatibility mirror，而不是再把新的实例逻辑补回 `runtime/runtime.Module`。

## Required Re-validation

已执行并建议保留为本轮复验命令：

```bash
cd code/backend && go test ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...
cd code/backend && go test ./internal/module/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

## Residual Risk

- 当前 review 为同会话自审归档，不是独立 reviewer gate。
- `runtime/application/*` 的 compatibility mirror 仍然存在，后续如果 `instance/*` 再演进，这两处需要继续同步，直到 mirror 被删除。

## Touched Known-Debt Status

- 已触达并部分收口的已知结构债：`runtime` 同时承担实例 owner 与容器适配。
- 本轮已完成的收口：生产 wiring 中的实例命令、查询、proxy ticket、maintenance 已从 `runtime/runtime.Module` 挪到 `instance` + `composition.InstanceModule`。
- 本轮仍保留但不再是生产 owner 的过渡面：`runtime/application/*` compatibility mirror。
