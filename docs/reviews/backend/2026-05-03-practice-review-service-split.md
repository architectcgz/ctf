# Review Target

- Repository: `/home/azhi/workspace/projects/ctf`
- Diff source: local working tree
- Scope:
  - `code/backend/internal/module/practice/application/commands/service.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
  - `code/backend/internal/module/practice/application/commands/contest_awd_runtime_subject.go`
  - `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
  - `code/backend/internal/module/practice/application/commands/awd_runtime_rules.go`
  - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/instance_start_service.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/manual_review_service.go`
  - `docs/plan/impl-plan/2026-05-03-practice-command-service-split-plan.md`

# Classification Check

- 结论：同意将本次改动视为非 trivial 的结构性后端改动。
- 原因：改动跨越实例启动、provisioning、AWD 运行时规则、提交与人工审核，但保持在同 package 内，不改变公开接口和行为边界。

# Gate Verdict

- `pass`

# Findings

- 本轮未发现需要阻塞合并的正确性、回归或测试缺口问题。

# Material Findings

- None.

# Senior Implementation Assessment

- 当前做法是这类拆分里风险最低的一种：先做同 package 的等价拆文件，再考虑是否继续拆 service owner。
- 这避免了在同一轮同时引入 package 边界调整、依赖注入变化和测试替身重写，符合当前代码的最小可审查改动原则。

# Required Re-validation

- 已执行：
  - `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/practice/application/commands`
- 如后续继续拆测试文件或继续向 `practice` / `runtime` 交界移动代码，建议追加：
  - `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/practice/... ./internal/module/runtime/...`

# Residual Risk

- 当前验证只覆盖 `practice/application/commands` 包级测试，没有重新跑更大范围的 `practice/...` 或 `runtime/...`。
- `service_test.go` 仍然是超大测试文件，后续 review 成本依旧偏高；不过这属于下一阶段拆测试文件的问题，不是本轮实现缺陷。
