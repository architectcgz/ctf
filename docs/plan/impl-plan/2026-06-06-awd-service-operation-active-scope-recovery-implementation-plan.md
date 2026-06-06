# AWD Service Operation Active Scope Recovery Implementation Plan

**Goal:** 在创建新的 AWD service operation 前，先收口同 `contest/team/service` scope 下遗留的活跃 operation，避免继续撞上 scope 级唯一约束。

**Architecture:** 唯一活跃约束的 owner 仍然是 `awd_service_operations` 表上的 scope 级约束；仓储写入口负责在插入新 operation 前把同 scope 的旧活跃记录关闭，不把这段补救逻辑散落到多个 command/service。

**Tech Stack:** Go repository, GORM transaction, repository tests

---

## Task Metadata

- Task Slug: `2026-06-06-awd-service-operation-active-scope-recovery`
- Started At: `2026-06-06T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/ctf`
- Branch: `task/2026-06-06-awd-service-operation-active-scope-recovery`

## Objective And Non-Goals

- Objective:
  - 在 `CreateAWDServiceOperation` 写入口里补 scope 级遗留活跃 operation 收口。
  - 用 repository 测试覆盖“旧实例残留 + 新实例接管”这条回归路径。
- Non-Goals:
  - 不改 `awd_service_operations` 表的唯一约束定义。
  - 不把清理逻辑扩散到多个 application command。
  - 不引入新的 runtime migration 或异步补偿任务。

## Inputs

- Source docs:
  - `.harness/reuse-decisions/awd-service-operation-active-scope-recovery.md`
- Related architecture/contracts:
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/internal/module/practice/infrastructure/repository_test.go`
  - `code/backend/migrations/000002_create_awd_service_operations.up.sql`
- Related prior work:
  - `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 命中受保护的 backend repository surface。
  - 需要把数据库唯一约束 owner 与 repository 行为重新收口一致。

## Files

- Create:
  - `docs/plan/impl-plan/2026-06-06-awd-service-operation-active-scope-recovery-implementation-plan.md`
- Modify:
  - `code/backend/internal/module/practice/infrastructure/repository.go`
  - `code/backend/internal/module/practice/infrastructure/repository_test.go`
- Review:
  - `.harness/reuse-decisions/awd-service-operation-active-scope-recovery.md`
- Test:
  - `go test ./internal/module/practice/infrastructure -run TestRepositoryCreateAWDServiceOperationClosesStaleActiveScopeEntries -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/practice/infrastructure/repository.go`
- Reuse / extend / split / create-new decision:
  - 扩展现有 repository 写入口，而不是在多个 command 层复制收口逻辑。
- Owner boundary:
  - 数据库唯一约束继续定义正确 scope。
  - repository 在插入前关闭遗留活跃记录，确保应用写路径与数据库 owner 一致。
- Why this is the narrowest safe surface:
  - 只修改单个写入口和对应测试面，就能覆盖当前异常来源，不需要扩散到更上层 orchestration。

## Validation

- Commands:
  - `go test ./internal/module/practice/infrastructure -run TestRepositoryCreateAWDServiceOperationClosesStaleActiveScopeEntries -count=1`
- Manual checks:
  - 确认旧 operation 被标记为 `failed + superseded_by_new_operation`
  - 确认新 operation 保持活跃
- Review focus:
  - scope owner 是否收口在 repository 写入口
  - 是否仍然允许同 scope 下残留多个活跃 operation
