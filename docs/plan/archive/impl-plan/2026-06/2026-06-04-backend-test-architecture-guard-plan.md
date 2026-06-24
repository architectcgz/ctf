# 2026-06-04 backend test architecture guard plan

## Objective

- 给后端测试分层补最小有效的机械 guardrail，防止系统测试 owner 再回流到 `internal/app` 或 `tests/system/http` 自己接管 DB/env setup。
- 把新 guardrail 接入现有 `scripts/check-backend-architecture.sh --full`，让它成为后端架构检查的一部分。

## Non-goals

- 本轮不迁移新的系统测试文件。
- 本轮不处理 `TestFullRouter_AuthorizedSmokeMatrix` 既有 500。
- 本轮不处理 `runtime/practice` 依赖真实环境的失败。
- 本轮不重做 `tests/testkit` 或 `tests/runtime` 目录结构。

## Inputs

- `code/backend/tests/README.md`
- `works/backend-test-architecture-rewrite-blueprint.md`
- `code/backend/internal/module/architecture_test.go`
- `code/backend/internal/app/backend_context_architecture_test.go`
- `scripts/check-backend-architecture.sh`
- `.harness/reuse-decisions/backend-test-architecture-guard.md`

## Current problem

- 测试重构蓝图和 README 已经说明了推荐分层，但仓库里还没有对应的源码级 guardrail。
- 现在只能靠 review 和记忆维持约定，后续很容易出现两类回流：
  - `internal/app` 里的系统测试壳重新长回几百行场景 owner。
  - `tests/system/http` 为了图省事开始自己创建 DB、AutoMigrate、或持有本地 env setup。

## Working design

### Target structure

- 新增 `code/backend/tests/architecture/test_architecture_test.go`
  - `TestInternalAppSystemTestShellsStayThinAndReferenceScenarioOwners`
  - `TestSystemHTTPScenarioPackagesDoNotOwnEnvOrPersistenceSetup`
  - `TestBackendTestsReadmeListsCurrentSystemHTTPScenarioDirectories`
- 更新 `scripts/check-backend-architecture.sh`
  - `--full` 额外跑 `go test ./tests/architecture`

### Boundary choice

- 守卫先覆盖最确定、误报最低的两条边界：
  - `internal/app` 系统测试壳必须薄，并显式指向 `tests/system/http` 的场景 owner。
  - `tests/system/http` 只写 scenario/assertion，不拥有 DB/env/persistence setup。
- README 同步检查作为轻量第三条守卫，确保目录说明不漂移。
- 对仍未拆完的超大壳文件保留显式 allowlist，但要求上限明确、allowlist 可失效。

## Task slices

### Slice 1：startup gate

- Goal：补齐 reuse decision 与 implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-guard`

### Slice 2：实现测试架构守卫

- Goal：在 `code/backend/tests/architecture/` 下补源码扫描测试，并接入 backend architecture script。
- Validation：
  - `cd code/backend && go test ./tests/architecture -count=1`

### Slice 3：文档入口与回归

- Goal：同步 README / architecture entry，并确认 full backend architecture gate 正常。
- Validation：
  - `bash scripts/check-backend-architecture.sh --full`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- 后端测试架构 guardrail
- backend architecture script 入口
- 测试目录说明与架构入口文档

## Data / API / compatibility impact

- 无生产代码、数据库或 API 行为变化。
- 风险主要在 guardrail 过严导致噪音，因此本轮只做源码级、低误报的约束，并保留少量 reviewed allowlist。

## Review fit check

- Owner 清晰：测试架构守卫落在 `code/backend/tests/architecture`，由 backend architecture script 调度。
- Scope 清晰：只增加 guardrail，不夹带新的测试迁移。
- 回滚简单：全是测试和脚本入口改动，直接回退即可。

## Rollback / recovery

- 若 guardrail 误伤现有结构，可单独回退 `code/backend/tests/architecture` 与 `scripts/check-backend-architecture.sh` 的本轮改动，不影响生产逻辑。
