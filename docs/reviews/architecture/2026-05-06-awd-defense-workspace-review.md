# AWD Defense Workspace Boundary Review

## Review Target

- Repository: `ctf`
- Branch: `main`
- Diff source: commits `0e467671` through `14b5e7de`, plus current local working tree for Task 7 closeout
- Reviewed files:
  - `code/backend/internal/module/challenge/domain/awd_package_parser.go`
  - `code/backend/internal/model/awd_defense_workspace.go`
  - `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `code/backend/internal/module/runtime/application/queries/proxy_ticket_service.go`
  - `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
  - `code/backend/internal/module/contest/application/queries/awd_workspace_result.go`
  - `code/backend/internal/dto/contest_awd_workspace.go`
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/contest.ts`
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
  - `code/frontend/src/router/routes/studentRoutes.ts`
  - `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
  - `docs/architecture/features/AWD防守工作区与边界设计.md`
  - `docs/architecture/features/AWD学员实战工作台设计.md`
  - `challenges/awd/challenge-package-contract.md`
  - `docs/plan/impl-plan/2026-05-06-awd-defense-workspace-boundary-implementation-plan.md`

## Classification Check

同意结构性改动工作流和非 trivial review gate 判定。这个切片同时改了题包契约、workspace 状态持久化、runtime 挂载语义、SSH scope、学生态 read model、前端入口和历史设计事实源，不能只按单点功能变更处理。

## Gate Verdict

Pass.

未发现阻塞性 correctness、权限回退或已知债务未收口问题。

## Findings

No material findings.

## Material Findings

无。

## Senior Implementation Assessment

当前实现和设计边界基本对齐，而且收口方式比“兼容保留旧入口”更稳：

- 题包契约从文件级 `editable_paths` 切到目录级 `defense_workspace`，把运行契约、业务工作区和 checker 物理分层。
- `awd_defense_workspaces` 把 `contest + team + service` 级状态、`workspace_revision` 和 companion container 归属持久化出来，没有继续把 workspace 生命周期混进实例运行态。
- SSH 票据签发与网关校验同时绑定 `workspace_revision`，普通 restart 不轮换 revision，`reseed / recreate` 才失效旧连接，和设计语义一致。
- 学生态 `GET /contests/:id/awd/workspace` 只暴露 `defense_connection` 摘要，不再泄露 `defense_scope`、文件路径、容器标识或浏览器工作台接口。
- 前端 battle 页删除独立防守 workbench 路由和文件 API，只保留打开服务、SSH 连接、重启和战场操作，避免旧债继续挂在 touched surface 上。

更低风险的替代方案其实不是“多保留一点兼容入口”，因为那会继续让 `editable_paths`、浏览器文件工作台和 SSH 直连 runtime 这三处已知债务留在本次 touched surface。当前实现选择直接收口，风险更低。

## Required Re-validation

已执行并通过：

```bash
cd code/backend
timeout 900s go test ./internal/module/challenge/... ./internal/module/practice/... ./internal/module/runtime/... ./internal/module/contest/... ./internal/app -run 'AWD|Defense|Workspace|Router' -count=1

cd code/frontend
timeout 600s npx vitest run src/api/__tests__/contest.test.ts src/views/contests/__tests__ src/features/contest-awd-workspace

cd /home/azhi/workspace/projects/ctf
timeout 300s bash scripts/check-consistency.sh
```

结果：

- 后端集成验证 PASS，`internal/app` 用时 `54.343s`
- 前端集成验证 PASS，`10` 个 test files、`61` 个 tests 全部通过
- harness 一致性检查 PASS

## Residual Risk

- 本轮没有补真实 Docker 环境下的端到端 SSH 会话 smoke；当前证据主要来自 repository/query/runtime 测试、前端集成测试和 harness 检查。伴随 companion container、root mount 和 restart/reseed 的真实容器联调，仍建议在后续环境里补一次。
- 历史阶段性文档仍留在仓库中作为历史记录；当前事实源依赖本轮新增或更新的设计文档，后续 review 需要继续按这些事实源判断，而不是回退到旧 workbench 方案。

## Touched Known-debt Status

本次明确触达并收口了计划中标注的已知债面：

- 学生侧 `defense_scope` 提示与 `editable_paths`
- 浏览器文件树 / 在线编辑 / 命令执行 workbench
- SSH 直连 service runtime 容器

这些债务在当前 touched surface 上都已关闭，没有被降级成 residual risk 或 follow-up。
