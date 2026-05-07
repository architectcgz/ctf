# AWD Defense Workspace Auto Git Writable Roots Review

- Review target:
  - Repository: `ctf`
  - Diff source: current uncommitted working tree, scoped to the AWD defense workspace auto-git writable-roots slice
  - Plan: `docs/plan/impl-plan/2026-05-07-awd-defense-workspace-auto-git-writable-roots-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
    - `code/backend/internal/module/practice/application/commands/service_test.go`
    - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
  - Surrounding context checked:
    - `code/backend/internal/module/challenge/domain/awd_package_parser.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`

- Classification check:
  - 同意按 `non-trivial` backend review gate 处理。
  - 原因：这次改动改变了 AWD defense workspace companion 的 bootstrap 目标选择规则，直接涉及 fail-fast、readonly root 跳过语义，以及 mount target 到 Git 初始化路径的映射，不是单纯常量替换。

- Gate verdict:
  - `pass`

- Findings:
  - 无 findings。

- Material findings:
  - 无 material findings。

- Senior implementation assessment:
  - 这次 re-review 只回看上轮 blocker：`writable_roots` owner 是否真正进入实现语义。
  - 当前版本已经在 `parseAWDDefenseWorkspaceConfig` 里显式读取 `writable_roots`，并把不在该集合内的 `workspace_root` 默认标成 readonly；后续 bootstrap 仍然只对非 readonly mount 生成 Git 初始化片段，所以 auto-git 目标现在由 `writable_roots` 决定，而不是再靠 `readonly_roots` 缺省值反推。
  - 新增的解析级测试也覆盖了上轮指出的反例：当某个 `workspace_root` 不在 `writable_roots` 里时，它会被视为 readonly，不再落入 auto-git 目标集合。

- Required re-validation:
  - 本次 scoped re-review 已执行并通过的定向验证：

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/practice/application/commands -run 'TestParseAWDDefenseWorkspaceConfigTreatsRootsOutsideWritableSetAsReadonly|TestCreateAWDDefenseWorkspaceCompanionInitializesGitReposForWritableMounts|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1 -timeout 3m
```

- Residual risk:
  - 这次是 scoped re-review，只检查上轮 `writable_roots owner` blocker 是否收口，没有重新展开整个 diff 的 correctness / regression 面。
  - 当前验证仍然是 command 装配和 workspace 创建/重建链路级别，没有真实起容器去验收非 `/workspace/src` 根的首次进入体验。

- Touched known-debt status:
  - 这次 review 没有发现当前 active fact source 里要求在该 touched surface 同步收口的已知结构债。
