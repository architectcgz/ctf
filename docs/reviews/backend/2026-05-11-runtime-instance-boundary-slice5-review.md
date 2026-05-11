# Runtime / Instance 边界 Phase 2 Slice 5 Review

## Review Target

- 仓库：`ctf`
- 分支：`main`
- 审查范围：`code/backend/internal/module/instance/contracts/services.go`、`code/backend/internal/module/runtime/runtime/{module,adapters}.go`、`docs/plan/impl-plan/2026-05-11-runtime-instance-boundary-phase2-slice5-implementation-plan.md`、`docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/design/backend-module-boundary-target.md`
- diff 来源：当前工作区未提交改动
- 分类复核：同意 `non-trivial / structural refactor`，本轮目标是给 `runtime -> instance` 之间补一个合法的 service contract landing zone
- 审查上下文：本档为当前会话自审归档；本次没有独立 reviewer，因此它不是严格意义上的独立 review gate

## Gate Verdict

`pass with minor issues`

当前改动没有新的 material correctness / regression blocker。`runtime/runtime/adapters.go` 已改为依赖 `instance/contracts`，并删掉 runtime 模块内部那组临时 instance owner 接口；这给下一刀处理 compat mirror 提供了合法落点。

## Findings

### 无 material findings

- 新增的 `instance/contracts` 只定义实例 owner 暴露给外部模块的 command / query / proxy ticket / maintenance interface，没有把业务实现搬进 contract 包。
- `runtime/runtime/adapters.go` 只是替换依赖落点，没有改变 handler adapter 的业务分支和输出结构。
- 相关最小编译面和 runtime / composition / handler 回归面都还能通过。

### Minor issue

- `runtime/application/*` compat mirror 仍保留 duplicated implementation。`instance/contracts` 只是把合法依赖落点补出来，还没有把 compat mirror 真正压成 contract-backed wrapper。
- 下一刀仍然要回到 compat mirror 本体，选择：
  - 基于 `instance/contracts` 继续缩成薄壳
  - 或继续迁空剩余依赖后直接删除 mirror

## Senior Implementation Assessment

这轮先补 `instance/contracts` 是比“再次直接尝试 runtime import instance/application”更稳的做法：

- 避开了 `code/backend/internal/app/architecture_rules_test.go` 对 concrete cross-module application import 的限制
- 也避免继续让 runtime 模块自己声明一组临时 instance owner 接口，降低后续再分叉的风险
- 改动面只落在依赖定义和 adapter wiring 上，行为面基本不动，适合作为 compat mirror 瘦身前的前置收口

## Required Re-validation

建议保留为本轮复验命令：

```bash
cd code/backend && go test ./internal/module/runtime/runtime ./internal/app/composition ./internal/module/runtime/api/http
cd code/backend && go test ./internal/module/runtime/application/... ./internal/module/instance/... ./internal/module/runtime/... ./internal/app/...
cd code/backend && go test ./internal/module/...
python3 scripts/check-docs-consistency.py
bash scripts/check-consistency.sh
bash scripts/check-workflow-complete.sh
```

## Residual Risk

- 当前 review 为同会话自审归档，不是独立 reviewer gate。
- compat mirror 本体还没动，后续如果不沿 `instance/contracts` 继续推进，runtime 底层仍会停在双份实现状态。

## Touched Known-Debt Status

- 已触达的已知结构债：`runtime` 对实例 owner 的依赖之前仍通过 runtime 模块本地临时接口维持。
- 本轮已完成的收口：`runtime/runtime` 这侧已经改成依赖 `instance/contracts`，合法 landing zone 已明确。
- 本轮未完成但已明确记录的部分：`runtime/application/*` compat mirror 仍未删除，也还没有压成 contract-backed wrapper。
