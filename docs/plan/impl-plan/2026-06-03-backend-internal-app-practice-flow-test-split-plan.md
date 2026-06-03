# 2026-06-03 backend internal app practice flow test split plan

## Objective

- 把 `code/backend/internal/app/practice_flow_integration_test.go` 中的巨型流程测试拆成多个按职责聚焦的顶层测试。
- 保留原文件作为 practice flow shared env、fixture 和 helper owner。
- 继续降低 `internal/app` 中 1000+ 行测试文件的 owner 混杂和 review 成本。

## Non-goals

- 本轮不改 practice、challenge、runtime、ops 的业务逻辑。
- 本轮不重写 `newPracticeFlowTestEnv`、登录 helper、JSON helper 或 timeline/assert helper。
- 本轮不继续拆 `router_test.go`。

## Inputs

- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `.harness/reuse-decisions/backend-internal-app-practice-flow-test-split.md`

## Current problem

- `practice_flow_integration_test.go` 当前有 1359 行。
- 它只有两个顶层测试，但 `TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge` 一次覆盖创建题目、访问实例、代理请求、提交、进度、时间线、教师证据、攻击会话和审计多类责任。
- 这种“单个大测试吞掉整个流程”的写法让定位回归点和增量扩展断言都很费力。

## Working design

### Target structure

- 原文件保留：
  - `flowTestEnv` 及相关响应类型
  - `newPracticeFlowTestEnv` 与各类 login / decode / assert helper
- 新增 3 个按职责归类的测试文件：
  - `practice_flow_scenario_test.go`
  - `practice_flow_lifecycle_integration_test.go`
  - `practice_flow_observability_integration_test.go`

### Boundary choice

- 不更换测试基建，只在现有 helper 上补一个 scenario runner。
- scenario runner 负责把“已发布题目完整练习链路”跑通并返回中间结果。
- 顶层测试只断言各自负责的流程面：
  - lifecycle / instance / proxy
  - submission / history / progress
  - timeline / evidence / attack sessions / audit
  - unpublished negative path

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-internal-app-practice-flow-test-split`

### Slice 2：抽 scenario runner 并拆顶层测试

- Goal：把 giant test 下沉成共享 runner，再将顶层断言拆到更短的测试函数和新文件。
- Touched files：
  - `code/backend/internal/app/practice_flow_integration_test.go`
  - `code/backend/internal/app/practice_flow_scenario_test.go`
  - `code/backend/internal/app/practice_flow_lifecycle_integration_test.go`
  - `code/backend/internal/app/practice_flow_observability_integration_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_' -count=1`

### Slice 3：一致性收口

- Goal：确认 harness 文档与测试文件组织没有引入新的漂移。
- Validation：
  - `bash scripts/check-consistency.sh`

## Expected change surface

- `internal/app` practice flow 测试文件组织
- practice flow 顶层测试结构
- harness reuse decision 与 implementation plan

## Data / API / compatibility impact

- 无生产数据、API 或业务语义变更。
- 风险主要在：
  - giant test 拆分后遗漏原有断言
  - scenario runner 返回的中间结果不全，导致测试退化
  - 为了复用而把太多断言藏进 helper，反而降低可读性

## Review fit check

- Owner 清晰：原文件收口为 env / helper owner，scenario runner 独立成共享文件，其余新文件只承载 lifecycle 和 observability 两组测试。
- Reuse 清晰：继续复用现有 practice flow fixture，不新造新的 app test env。
- 结构收敛：既然本轮已经触碰 `practice_flow_integration_test.go` 这个 oversized surface，就直接把 giant test 的多责任拆开，而不是只把小测试挪出去继续保留同一问题。

## Rollback / recovery

- 纯测试代码与文档组织调整，可直接回退到拆分前的单文件版本。
