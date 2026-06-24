# Backend Application Service Boundary Convergence Review Round 1

## Review 对象

- Commit: 未提交，当前 worktree diff
- Branch: `task/2026-06-21-backend-application-service-boundary-convergence`
- Task / Plan: `2026-06-21-backend-application-service-boundary-convergence`；`docs/plan/archive/impl-plan/2026-06/2026-06-21-backend-application-service-boundary-convergence-implementation-plan.md`
- Reviewer mode: same-context self-check
- Diff basis: `git diff`

## 结论

- Gate verdict: pass for same-context self-check
- 正式 gate 状态：未满足独立 reviewer gate。当前记录只能作为工具受限下的临时 same-context review 证据，后续若按正式 gate 收口，应在提交后补一轮绑定 commit 或 commit range 的独立 review。

## Findings

### Minor: retired workbench surface 删除不完整

- Location: `code/backend/internal/module/instance/contracts/instance_output.go`、`code/backend/internal/testutil/runtimeadapters/http_service.go`
- Issue: 初始实现删除了 `AWDDefenseWorkbenchService` 实现和 contract interface，但仍保留 `AWDDefenseFile*` / `AWDDefenseDirectory*` / `AWDDefenseCommand*` DTO，以及 testutil facade 上的 `ReadAWDDefenseFile` / `SaveAWDDefenseFile` / `RunAWDDefenseCommand` 等方法。
- Why it matters: 本任务目标是降低历史 workbench surface 的阅读误导；这些残留会继续让读者误判文件编辑 / 命令 workbench 仍是当前 instance owner 对外能力。
- Resolution: 已删除残留 DTO 和 testutil 方法；`rg -n "AWDDefense(File|Directory|Command)|ReadAWDDefenseFile|ListAWDDefenseDirectory|SaveAWDDefenseFile|RunAWDDefenseCommand" code/backend/internal code/backend/tests -g '*.go'` 无匹配。

## 必须重跑的验证

- `go test ./internal/module/instance/... ./internal/testutil/... ./internal/app/composition/...`
- `go test ./internal/module/container_runtime/...`
- `go test ./internal/module/... ./internal/app/composition/...`
- `bash scripts/check-backend-architecture.sh`
- `bash scripts/check-workflow-complete.sh`

## 残余风险

- 本轮 review 是 same-context self-check，不是独立 reviewer gate。
- 本次只收敛低风险 application 根目录 service 和 retired workbench surface；`contest/application/commands/AWDService` 与 `container_runtime/application/commands/provisioning_service.go` 仍按既有 todo 保留为后续观察项。
