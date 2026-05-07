# AWD Defense Workspace Auto Git Implementation Plan

**Goal:** 让 AWD defense workspace 在首次创建 companion shell 容器时，自动把持久化的 `/workspace/src` 初始化为一个本地 Git 仓库并生成初始提交，这样学生进入 `workspace/src` 后可以直接执行 `git status`，不需要手工 `git init`。

**Non-goals:**

- 不修改题包 `defense_workspace` 契约，不新增新的配置字段
- 不把 `.git` 放到 `/workspace` 根目录，也不尝试把只读 root 初始化为仓库
- 不在本次切片中引入预构建 workspace shell 镜像或新的 seed 同步机制

**Source architecture or design docs:**

- `challenges/awd/challenge-package-contract.md`
- `challenges/awd/README.md`
- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`

**Dependency order:**

1. 先用测试把 bootstrap 命令应包含的 Git 初始化语义钉住
2. 再最小修改 workspace companion 启动命令
3. 最后跑定向验证并归档独立 review

**Expected specialist skills:**

- `brainstorming`
- `test-driven-development`
- `backend-engineer`
- `test-engineer`
- `code-reviewer`

## Task 1

**Goal:** 为 AWD workspace companion 的 bootstrap 行为补一个会失败的测试，明确 `/workspace/src` 首次启动时必须自动初始化 Git 仓库。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/service_test.go`

**Dependencies:** 无

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestAWDDefenseWorkspaceBootstrapCommandInitializesGitRepo|TestCreateSingleAWDContainerUsesPrivateTopology' -count=1 -timeout 3m
```

**Review focus:** 测试是否直接约束了 `/workspace/src` 的 Git 初始化语义，而不是只跟随常量改动一起通过。

**Risk notes:** 如果测试只断言常量相等，会失去红灯价值；必须检查具体命令内容。

## Task 2

**Goal:** 在 workspace companion bootstrap 命令中补齐幂等的 Git 初始化和初始提交逻辑。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/awd_defense_workspace_support.go`

**Dependencies:** Task 1

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestAWDDefenseWorkspaceBootstrapCommandInitializesGitRepo|TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1 -timeout 3m
```

**Review focus:** `/workspace/src` 是否作为唯一 Git repo 目标；命令是否幂等；是否为首次提交补齐本地 `user.name` 和 `user.email`，避免容器内 commit 因 identity 缺失失败。

**Risk notes:** 仍然依赖容器启动时可安装 `git`；若 `/workspace/src` 目录不存在或为空，应保持命令不致命并继续进入常驻态。

## Task 3

**Goal:** 跑最小充分验证并完成独立 review 归档。

**Touched modules or boundaries:**

- `code/backend/internal/module/practice/application/commands/{service_test.go,runtime_container_create_test.go,instance_start_service_test.go}`
- `docs/reviews/backend/2026-05-07-awd-defense-workspace-auto-git-review.md`

**Dependencies:** Task 2

**Validation:**

```bash
cd code/backend
go test ./internal/module/practice/application/commands -run 'TestAWDDefenseWorkspaceBootstrapCommandInitializesGitRepo|TestCreateSingleAWDContainerUsesPrivateTopology|TestRestartContestAWDServiceReusesExistingDefenseWorkspaceContainer|TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer' -count=1 -timeout 3m
go test ./internal/module/practice/application/commands -run 'TestCreateSingleAWDContainer.*Workspace|TestRestartContestAWDService.*DefenseWorkspace' -count=1 -timeout 3m
```

**Review focus:** 当前改动是否只影响 workspace companion；是否留下新的回归点；测试覆盖是否足以证明学生进入 `/workspace/src` 时可以直接使用 Git。

**Risk notes:** 当前验证仍停留在 bootstrap 命令装配层，没有直接起容器验证真实 `git status`；若需要运行态验收，应另走 Docker 集成检查。

## Integration Checks

- workspace companion create request 仍保持原有 image、working dir、mounts、network alias 与 entrypoint 配置
- runtime 容器 mounts 不受影响，Git 自动初始化只发生在 companion shell 启动链路

## Rollback / Recovery Notes

- 该切片只改 bootstrap 命令和测试，回退时可以单独撤回相关提交
- 不涉及数据库、迁移、持久化 schema 或题包格式变更

## Residual Risks

- 首次 companion 启动仍依赖 Alpine 包仓库能成功安装 `git`
- 目前没有额外把 `vim` 缺失问题扩展成编辑器行为测试，仍沿用现有“命令包含安装步骤”的验证粒度
