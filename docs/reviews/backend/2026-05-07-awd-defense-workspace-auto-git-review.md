# AWD Defense Workspace Auto Git Review

- Review target:
  - Repository: `ctf`
  - Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/fix-awd-defense-workspace-auto-git`
  - Branch: `fix/awd-defense-workspace-auto-git`
  - Diff source: current uncommitted working tree against `3df581a5902fa67c3e9d2a2098be5c77379ef4ed`
  - Plan: `docs/plan/impl-plan/2026-05-07-awd-defense-workspace-auto-git-implementation-plan.md`
  - Files reviewed:
    - `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
    - `code/backend/internal/module/practice/application/commands/service_test.go`
    - `docs/plan/impl-plan/2026-05-07-awd-defense-workspace-auto-git-implementation-plan.md`
  - Surrounding context checked:
    - `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
    - `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`

- Classification check:
  - 同意按 `non-trivial` 后端 gate review 处理。
  - 原因：这次改动虽然 diff 很小，但它改变了 AWD workspace companion 的启动行为，并新增了 plan 与独立 review 要求；风险点在容器 bootstrap 成败语义，而不只是字符串常量更新。

- Gate verdict:
  - `pass`

- Findings:
  - 无 findings。

- Material findings:
  - 无 material findings。

- Senior implementation assessment:
  - 把 auto-git 放在 workspace companion bootstrap 里是这类需求的最小实现面，这个方向本身没有问题。
  - 本轮 re-review 关注的 fail-fast 契约已经收口：`awdDefenseWorkspaceBootstrapCommand` 开头加入 `set -e`，`apk add` 和后续 Git 初始化步骤恢复为失败即退出，不再把未完成 bootstrap 的 companion 容器伪装成可复用的 `Running` workspace。
  - 对应测试虽然仍是命令装配级断言，但现在已经显式约束了 `set -e` 和 `/workspace/src` Git 初始化片段，足以覆盖这次 review 指出的回归面。

- Required re-validation:
  - 本轮已执行并通过：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/fix-awd-defense-workspace-auto-git/code/backend
go test ./internal/module/practice/application/commands -run 'TestAWDDefenseWorkspaceBootstrapCommandInitializesGitRepo|TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceReusesExistingDefenseWorkspaceContainer|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1
```

- Residual risk:
  - 当前验证仍停留在 command 装配和 AWD workspace 创建/重建路径，没有真实起 Docker 容器去验收“首次进入 `/workspace/src` 后可直接 `git status`”。
  - `apk add` 依赖 Alpine 仓库可达，这个外部依赖本身仍然存在；不过当前脚本已经恢复 fail-fast，不会再把这类失败伪装成成功。

- Touched known-debt status:
  - 这次 diff 触达了 `service_test.go`，而 `docs/reviews/backend/2026-05-03-practice-review-service-split.md` 里记录过“后续拆测试文件”的历史债务。
  - 基于当前已读事实源，这个债务目前是历史维护项，不是当前 backend README 里的 active blocker；本次只新增了一个聚焦断言测试，没有明显扩大测试 ownership 混杂度，因此我没有把这条历史债务单独升级成新的 gate blocker。
