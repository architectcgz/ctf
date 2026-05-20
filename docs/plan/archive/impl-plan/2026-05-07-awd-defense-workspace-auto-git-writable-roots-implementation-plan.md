# AWD Defense Workspace Auto Git Writable Roots Implementation Plan

**Goal:** 把 AWD defense workspace 的 Git 自动初始化从硬编码 `/workspace/src` 收口成基于 `defense_workspace.writable_roots` 的通用规则。后续只要是 AWD 题目，workspace companion 在首次启动时都应对每个可写业务根自动初始化本地 Git 仓库；Jeopardy 链路不受影响。

**Non-goals:**

- 不给 Jeopardy 或其他非 AWD 模式题目补 Git / 编辑器 / shell 环境
- 不把多个独立 named volume 拼成单一跨目录 Git 仓库
- 不修改 AWD 题包契约、数据库结构或 SSH 网关协议
- 不在本次切片中引入新的持久化根或额外 workspace volume

**Source architecture or design docs:**

- `challenges/awd/challenge-package-contract.md`
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `docs/reviews/backend/2026-05-07-awd-defense-workspace-auto-git-review.md`

**Dependency order:**

1. 先补失败测试，明确 writable roots 通用初始化语义
2. 再把 companion bootstrap 命令改成按 writable mounts 动态生成
3. 最后跑定向验证并归档独立 review

**Expected specialist skills:**

- `brainstorming`
- `test-driven-development`
- `backend-engineer`
- `test-engineer`
- `code-reviewer`

## Task 1

**Goal:** 用失败测试锁定“每个 writable root 都应初始化 Git，而 readonly root 不应初始化”的命令构造语义。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/service_test.go`

**Dependencies:** 无

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestCreateAWDDefenseWorkspaceCompanionInitializesGitReposForWritableMounts|TestParseAWDDefenseWorkspaceConfigTreatsRootsOutsideWritableSetAsReadonly|TestCreateSingleAWDContainerUsesPrivateTopology' -count=1 -timeout 3m
```

**Review focus:** 测试是否真正约束了 writable roots 的通用规则，而不是继续绑定 `/workspace/src` 常量。

**Risk notes:** 如果测试仍只检查常量片段，后续改动可能继续漏掉 `/workspace/app`、`/workspace/templates` 等非 `src` 根。

## Task 2

**Goal:** 把 AWD workspace companion 的 bootstrap 命令改成基于 `workspaceMounts` 动态生成，并保持 fail-fast。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`

**Dependencies:** Task 1

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestCreateAWDDefenseWorkspaceCompanionInitializesGitReposForWritableMounts|TestParseAWDDefenseWorkspaceConfigTreatsRootsOutsideWritableSetAsReadonly|TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1 -timeout 3m
```

**Review focus:** companion 启动命令是否继续保持 `set -e` 的 fail-fast 语义；Git 初始化目标是否只来自 writable mounts；readonly mount 是否被正确跳过。

**Risk notes:** 多个 writable roots 只能各自成为独立 Git repo，因为当前挂载模型是多个独立 volume，无法在 `/workspace` 根持久化一个统一 `.git`。

## Task 3

**Goal:** 跑最小充分验证并完成独立 review。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/{service_test.go,runtime_container_create_test.go,instance_start_service_test.go}`
- `docs/reviews/backend/2026-05-07-awd-defense-workspace-auto-git-writable-roots-review.md`

**Dependencies:** Task 2

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestCreateAWDDefenseWorkspaceCompanionInitializesGitReposForWritableMounts|TestParseAWDDefenseWorkspaceConfigTreatsRootsOutsideWritableSetAsReadonly|TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceReusesExistingDefenseWorkspaceContainer|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1 -timeout 3m
go test ./internal/module/practice/application/commands -run 'TestCreateSingleAWDContainer.*Workspace|TestRestartContestAWDService.*DefenseWorkspace' -count=1 -timeout 3m
```

**Review focus:** 当前实现是否把 AWD auto-git 收口到题包契约层的 writable roots，而不是继续依赖某道题的目录命名；是否保持 Jeopardy 和非 AWD 路径零影响。

**Risk notes:** 当前验证仍停留在 command 装配和 AWD workspace 创建/重建路径，没有额外起一个“非 `/workspace/src`”题包做真实容器集成验证。

## Integration Checks

- workspace companion 仍保留现有 image、env、working dir、network alias 和 mounts 形状
- auto-git 只在 AWD defense workspace companion 启动时生效，不进入 runtime 容器和 Jeopardy 实例

## Rollback / Recovery Notes

- 改动只涉及 companion bootstrap 命令生成逻辑、相关测试以及 plan/review 文档，可单独回退
- 不涉及数据库、迁移、题包导入格式或运行时 API 结构

## Residual Risks

- `git` / `vim` / `nano` 仍依赖 companion 启动时能从 Alpine 仓库安装成功
- 题包若存在多个 writable roots，当前实现会为每个根初始化独立 Git 仓库，而不是单一工作区级仓库；这与当前 volume 模型一致，但需要在结果说明里明确
