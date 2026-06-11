<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# teaching-query-infra-split Implementation Plan

**Goal:** 拆分 `teaching_query` 基础设施层的宽查询仓储，按现有 query port 还原清晰 owner 边界，同时保持教师端 HTTP 契约与查询行为不变。

**Architecture:** `teaching_query` 已经在 application / ports 层按目录查询、overview、class insight、student review 拆出稳定用例边界；本次重构直接对齐这些 port，把 `runtime` 从单个 `NewRepository(db)` 宽 concrete 组装改成多个窄 adapter 组装，并把重查询 helper 从单文件移到对应 owner 文件。重构不改变 handler、application service、port 契约和返回结构，只调整 infrastructure 与 runtime wiring。

**Tech Stack:** Go, GORM, SQLite test fixture, architecture tests, module runtime wiring

---

## Task Metadata

- Task Slug: `2026-06-11-teaching-query-infra-split`
- Started At: `2026-06-11T13:08:50Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-teaching-query-infra-split`
- Branch: `task/2026-06-11-teaching-query-infra-split`

## Objective And Non-Goals

- Objective:
  - 消除 `teaching_query/runtime/module.go` 对宽 `repo interface{...}` 与 `queryinfra.NewRepository(deps.DB)` 的依赖。
  - 将 `teaching_query/infrastructure/repository.go` 按现有 query port / use case owner 拆成多个聚焦 adapter 文件。
  - 保持 `application/queries`、`api/http`、`ports` 外部行为与类型契约不变。
- Non-Goals:
  - 不改教师端 HTTP 路径、请求参数、响应结构。
  - 不新增跨模块 contract，也不把 `teaching_query` 重新拆回 owner 模块。
  - 不顺手优化 overview 的 N+1 查询或重写 SQL 语义；只收口当前宽仓储与 wiring 边界。

## Inputs

- Source docs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `code/backend/internal/module/teaching_query/ports/query.go`
  - `code/backend/internal/module/teaching_query/application/queries/*.go`
  - `code/backend/internal/module/teaching_query/runtime/module.go`
  - `code/backend/internal/module/teaching_query/architecture_test.go`
- Related prior work:
  - `docs/architecture/backend/07-modular-monolith-refactor.md` 中 teaching_query query surface phase 4 的拆分记录
  - `docs/todos/2026-06-11-backend-command-boundary-debt.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 这次改动会同时触达 runtime wiring、infrastructure 结构、架构测试与仓储测试，属于受保护实现面上的结构性重构。
  - 目标 surface 已被识别为 oversized / owner-mixed，不能继续在同一表面上叠加新逻辑而不收口边界。

## Files

- Create:
  - `code/backend/internal/module/teaching_query/infrastructure/class_query_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/student_directory_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/student_profile_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/student_activity_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/class_insight_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/overview_repository.go`
  - `code/backend/internal/module/teaching_query/infrastructure/shared_rows.go`
- Modify:
  - `code/backend/internal/module/teaching_query/runtime/module.go`
  - `code/backend/internal/module/teaching_query/architecture_test.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
- Review:
  - `code/backend/internal/module/teaching_query/application/queries/service.go`
  - `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- Test:
  - `go test ./internal/module/teaching_query/...`
  - `go test ./internal/module/teaching_query -run 'TestRuntimeUsesTypedDeps|TestInfrastructureDoesNotExposeWideRepository' -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - `contest/runtime/module.go`、`assessment/runtime/module.go`、`challenge/runtime/module.go` 的多仓储 runtime 装配方式
  - `teaching_query/application/queries/*` 现有按用例拆分的 repo interface
- Reuse / extend / split / create-new decision:
  - 复用现有 `ports` 和 application service 边界。
  - 拆分现有 wide infrastructure concrete；不新造额外 service 层，不改外部契约。
- Owner boundary:
  - `runtime` 只负责 concrete wiring。
  - `application/queries` 继续 owner 用例编排。
  - `ports` 继续 owner 消费侧 query capability。
  - `infrastructure` 改为每个 adapter 只 owner 自己的 query capability 与对应 SQL/ORM 细节。
- Why this is the narrowest safe surface:
  - 当前 application / ports 已经具备稳定边界，最小改动就是让 runtime 和 infrastructure 与它们对齐，而不是再扩散到 handler/contract/doc 层。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming` + `architect-agent`
- Why this pass fits:
  - 任务核心不是修单点 bug，而是收口一个已知宽仓储，先确认稳定 owner 边界和最小安全拆分面比直接改文件更重要。
- grill-with-docs findings:
  - `teaching_query` 在架构文档里被定义为教师视角只读聚合模块，拆分应按 query use case / port，而不是按底层表名。
  - `runtime/module.go` 目前通过 `repo interface{...}` + `queryinfra.NewRepository(deps.DB)` 重新把已拆开的 port 聚回一个宽 concrete，是当前结构债的直接落点。
  - `ListClassTeachingFactSnapshots` 与 `GetStudentEvidence` 是当前最重的聚合查询，应在本轮一起从单文件宽仓储里拆出 owner。
- Plan adjustments after challenge:
  - 本轮不只切 runtime wiring，还要同步拆掉宽 `Repository` / `NewRepository` surface，避免“外部看起来分了，内部仍是 God repo”的假收口。

## Validation

- [x] Step 1: 删除 `teaching_query/infrastructure/repository.go` 宽仓储表面，并按 query capability 拆出独立 adapter 文件与 shared row helper。
- [x] Step 2: 更新 `teaching_query/runtime/module.go`、相关 composition / seed / test wiring，改为按 typed dependency 装配窄 adapter，不再依赖 `NewRepository(db)`。
- [x] Step 3: 补齐 `architecture_test.go` 与 `repository_test.go` 的护栏，确认 runtime 不再暴露宽 concrete，`infrastructure` 不再持有单一 God repository owner。
- [x] Step 4: 同步更新受影响架构文档、implementation plan 与 review 证据，清理对旧 `teaching_query/infrastructure/repository.go` 路径的活动引用。
- [x] Step 5: 运行 `go test ./internal/module/teaching_query/... -count=1`、`go test ./internal/module/teaching_query -run 'TestRuntimeUsesTypedDeps|TestInfrastructureDoesNotExposeWideRepository' -count=1`、`go test ./internal/module/teaching_query/infrastructure -run TestClassInsightRepositoryListClassTeachingFactSnapshotsBackfillsAWDSuccessDimensionFacts -count=1`、`git diff --check`、`python3 scripts/check-docs-consistency.py` 与 `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`，确认本轮最小充分验证通过。

- Commands:
  - `go test ./internal/module/teaching_query -run 'TestRuntimeUsesTypedDeps|TestInfrastructureDoesNotExposeWideRepository' -count=1`
  - `go test ./internal/module/teaching_query/... -count=1`
  - `go test ./internal/module/teaching_query/infrastructure -run TestClassInsightRepositoryListClassTeachingFactSnapshotsBackfillsAWDSuccessDimensionFacts -count=1`
- Manual checks:
  - 确认 `runtime/module.go` 只做多 adapter wiring，不再持有宽 `repo interface{...}`。
  - 确认 `infrastructure` 中不再存在单个 `Repository` owner 全部教师查询。
- Review focus:
  - runtime wiring 是否仍然隐式组合成 provider-owned 宽口。
  - 学生活动 / 班级洞察 / overview 查询是否在拆分后保留原语义。
  - 是否留下新的“shared helper”回退成第二个宽 owner。
