# Runtime / Instance 边界 Phase 2 Slice 4 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/app/practice_flow_integration_test.go`、`code/backend/internal/module/runtime/service_test.go`、`code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`、`docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice4-implementation-plan.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/design/backend-module-boundary-target.md`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural refactor`，但本轮在 guardrail 复核后已回退成更小的“调用点迁移 + 文档收口”切片
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with minor issues`

最终保留下来的改动没有新的 material correctness / regression blocker。`practice_flow_integration_test.go` 与 `runtime/service_test.go` 已经改用 `instance/*` owner，slice4 也把“compat wrapper 方案会撞上架构 guardrail”明确回收到计划和设计文档里。

## Findings

### 无 material findings

- 当前提交没有继续保留会触发 `code/backend/internal/app/architecture_rules_test.go` 的 `runtime -> instance application` 直接 import。
- 外部直接构造 compat service 的两个主要调用点已经切到 `instancecmd.NewInstanceService`、`instanceqry.NewInstanceService`、`instanceqry.NewProxyTicketService`、`instancecmd.NewInstanceMaintenanceService`。
- `runtime/application/commands/runtime_maintenance_service_test.go` 保持原有 helper coverage，没有因为本轮回环而额外削弱测试面。

### Minor issue

- `runtime/application/*` compat mirror 仍保留 duplicated implementation。当前它已经不再是生产 wiring owner，但它也还不能直接改成 wrapper，因为 `code/backend/internal/app/architecture_rules_test.go` 会阻止 `runtime -> instance application` 的 concrete cross-module import。
- 下一刀必须在两条路里选一条继续收口：
  - 继续迁走剩余 compat 调用方，待外部依赖清空后删除旧 mirror
  - 或者先引入允许跨模块复用的中性 landing zone，再把 compat 层缩成真正的薄壳

## Senior Implementation Assessment

这轮最终采用的实现方式是当前约束下风险最低的一条线：

- 不再硬推 `runtime/application/*` 直接 import `instance/application/*`，避免提交一个会被架构测试立即拒绝的版本
- 先把 broad integration / runtime service 测试里的外部调用点迁到 `instance/*`，阻止 compat import path 继续外扩
- 同时把 guardrail 阻塞显式写进计划、架构事实和目标设计稿，避免后续重复把“thin wrapper”当成已验证方案

如果继续沿这条线推进，下一次不应该再直接尝试 `runtime -> instance/application` 转发，而应该先解决兼容层的合法落点问题。

## Required Re-validation

建议保留为本轮复验命令：

```bash
cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...
cd code/backend && go test ./internal/module/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

## Residual Risk

- 当前 review 为同会话自审归档，不是独立 reviewer gate。
- `runtime/application/*` compat mirror 仍保留 duplicated implementation；如果后续只继续迁调用点、不处理最终落点，结构债还会继续停留在底层 runtime 目录。

## Touched Known-Debt Status

- 已触达的已知结构债：`runtime` 物理目录里还残留 `instance / proxy-ticket / maintenance` compat mirror。
- 本轮已完成的收口：外部 broad integration / service test 已继续切到 `instance/*` owner，不再新增对 compat constructor 的依赖。
- 本轮未完成但已明确记录的部分：compat mirror 本体仍未删除，也还不能在当前 guardrail 下改成跨模块 wrapper。
