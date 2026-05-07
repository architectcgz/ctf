# AWD Defense Workspace Shell Environment Review

- Review target:
  - Repository: `ctf`
  - Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/fix-awd-defense-workspace-shell`
  - Branch: `fix/awd-defense-workspace-shell`
  - Diff source: working tree changes against current branch base
  - Plan: `docs/plan/impl-plan/2026-05-07-awd-defense-workspace-shell-environment-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
    - `code/backend/internal/module/practice/application/commands/awd_runtime_rules.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
    - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
    - `code/backend/internal/module/practice/application/commands/service_test.go`
    - `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
    - `code/backend/internal/module/practice/ports/ports.go`
    - `code/backend/internal/module/runtime/runtime/adapters.go`
    - `code/backend/internal/module/runtime/runtime/adapters_test.go`
    - `code/backend/internal/module/runtime/runtime/module.go`
    - `code/backend/internal/module/runtime/service_test.go`

- Classification check:
  - 同意按结构性 / non-trivial gate review 处理。

- Gate verdict:
  - `pass`

- Findings:
  - 无 findings。

- Material findings:
  - 无 material findings。

- Senior implementation assessment:
  - 本次实现保持了最小闭环，没有把问题扩成新的镜像构建/分发体系。
  - `InspectManagedContainer` 已在 practice port、runtime adapter 和 module wiring 上接通，workspace 复用逻辑不再停留在编译缺口。
  - workspace companion 的 `LANG/LC_ALL/TERM`、启动命令、`WorkingDir` 与共享挂载已经落到实际建容器链路，并由 adapter/runtime/practice 三层测试共同覆盖。
  - review loop 中补掉了两个 blocker：
    - 多节点 AWD topology 不再复用同一个 `ContainerName`
    - restart 时 workspace companion 缺失会重建，并有自动化测试证明

- Required re-validation:
  - 已执行并通过：

```bash
cd code/backend && go test ./internal/module/runtime ./internal/module/runtime/runtime -run 'ServiceCreateTopology|Practice|Topology' -count=1 -timeout 3m
cd code/backend && go test ./internal/module/practice/application/commands -run 'TestRestartContestAWDService.*DefenseWorkspace|TestCreateSingleAWDContainer.*Workspace|AWD|Workspace' -count=1 -timeout 3m
cd code/backend && go test ./internal/module/practice/application/commands -list 'RestartContestAWDService.*DefenseWorkspace'
```

- Residual risk:
  - workspace companion 首次启动仍依赖 Alpine 仓库可达；若 `apk add` 失败，shell 可用性会受影响。这是本切片明确接受的残余风险。
  - 本轮验证以单元/命令层为主，没有追加真实 Docker 端到端 smoke。

- Touched known-debt status:
  - 本次 touched surface 上的已知 blocker 已在同一 review loop 内收口，没有把已触达债务留成 follow-up。
