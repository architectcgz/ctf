# 2026-06-03 backend test architecture phase 2 plan

## Objective

- 把 `practice_flow` 的真实测试断言从 `internal/app` 迁到 `code/backend/tests/system/http`。
- 让 `internal/app` 只保留兼容测试入口，不再承担场景断言 owner。
- 验证 `tests/system/http` 作为 system test 落点的第一条迁移路径可行。

## Non-goals

- 本轮不删除 `internal/app/practice_flow_*` 测试文件。
- 本轮不迁移 `full_router_*` 或 contest / awd 的系统测试。
- 本轮不继续扩张 `internal/testutil/systemapp` 到新的业务场景。
- 本轮不改生产代码、路由或 runtime 行为。

## Inputs

- `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
- `code/backend/internal/app/practice_flow_observability_integration_test.go`
- `code/backend/internal/testutil/systemapp/*.go`
- `code/backend/tests/README.md`
- `.harness/reuse-decisions/backend-test-architecture-phase2.md`

## Current problem

- phase 1 已经把 helper owner 从 `internal/app` 抽到 `internal/testutil/systemapp`，但真实测试断言还留在 `internal/app/practice_flow_*`。
- 如果下一步继续拆大文件，这些断言仍会把 `internal/app` 留成系统测试事实 owner，目录职责还是混的。
- 直接删除旧测试文件又会进入高风险删除流程，不适合在这一刀里做。

## Working design

### Target structure

- 在 `code/backend/tests/system/http/practiceflow/` 建立新的场景断言 package。
- 新 package 直接依赖 `internal/testutil/systemapp` 提供的 env、scenario runner、request helper 和断言 helper。
- `internal/app/practice_flow_lifecycle_integration_test.go`
  与 `internal/app/practice_flow_observability_integration_test.go`
  改成只调用新 package 的 `Verify...` 函数。

### Boundary choice

- `internal/app` 保留测试入口，是为了兼容当前包路径和现有 `go test ./internal/app` 工作流。
- 场景断言迁移到 `tests/system/http`，先把 owner 迁走，后续再决定是否物理删除旧入口。
- 新 package 只服务 `practice_flow`，不在本轮泛化成更宽的测试框架。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 2 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase2`

### Slice 2：迁移 practice flow 场景断言 owner

- Goal：把 lifecycle / observability 两个测试文件里的断言逻辑迁到 `tests/system/http/practiceflow`，并让 `internal/app` 只保留兼容壳。
- Touched files：
  - `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
  - `code/backend/internal/app/practice_flow_observability_integration_test.go`
  - `code/backend/tests/system/http/practiceflow/*.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_' -count=1`

### Slice 3：更新 tests 入口说明

- Goal：把 `tests/system/http/practiceflow` 的现有落点写进 `code/backend/tests/README.md`。
- Touched files：
  - `code/backend/tests/README.md`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- `practice_flow` 场景测试 owner
- `tests/system/http` 首个真实目录
- `internal/app` 兼容测试壳

## Data / API / compatibility impact

- 无生产 API、数据结构或业务语义变化。
- 风险主要在：
  - 迁断言时误把 `internal/app` 私有类型耦合带进新目录
  - 兼容壳改薄后漏掉原有断言覆盖
  - 新测试目录的包路径组织不合理，导致后续迁移继续绕回 `internal/app`

## Review fit check

- Owner 清晰：系统测试断言 owner 从 `internal/app` 迁到 `tests/system/http/practiceflow`。
- Reuse 清晰：继续用已有 `internal/testutil/systemapp`，不重复封装 scenario helper。
- 结构收敛：这刀继续沿 phase 1 的边界推进，把 `internal/app` 从“helper owner”进一步收缩为“兼容入口”。

## Rollback / recovery

- 纯测试目录和断言 owner 调整，可直接回退到 `internal/app` 中的原始断言版本。
