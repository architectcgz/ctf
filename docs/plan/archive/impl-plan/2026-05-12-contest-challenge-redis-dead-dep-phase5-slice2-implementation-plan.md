# Contest Challenge Redis Dead Dependency Phase 5 Slice 2 Implementation Plan

## Objective

继续 phase 5 收窄 application concrete allowlist，删除 `contest/application/commands/challenge_service.go` 里已经失效的 Redis 直接依赖：

- 去掉 `ChallengeService` 对 `*redis.Client` 的字段和构造参数
- 同步删掉 runtime wiring、测试 helper 和 allowlist 例外
- 不改变竞赛加题、改题、删题及 AWD 服务关联的现有行为

## Non-goals

- 不处理 `contest` 里真实仍在使用的 Redis 依赖，例如榜单缓存、状态锁或 AWD 运行态缓存
- 不调整 `contest` challenge command service 的业务规则、DTO 或 HTTP handler
- 不修改 `contest` challenge query service

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `contest/application/commands/challenge_service.go` 当前仍 import `github.com/redis/go-redis/v9`
- `ChallengeService` 保存了 `redis` 字段，构造函数要求传入 Redis client
- 实际 challenge command 逻辑没有读取或写入任何 Redis key，这条依赖已经变成死参数
- `architecture_allowlist_test.go` 仍保留 `contest/application/commands/challenge_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

直接删除这条无效依赖：

1. 从 `ChallengeService` 结构体和构造函数移除 `redis` 字段与参数
2. 调整 `contest/runtime` 的 challenge handler 装配
3. 简化相关测试 helper，不再伪造 Redis client
4. 删除 allowlist 例外，并在设计稿里记录 phase 5 当前进展

## Ownership Boundary

- `contest/application/commands/ChallengeService`
  - 负责：竞赛题目与 AWD 服务关联的业务编排
  - 不负责：持有未使用的 Redis client 或为历史参数保留框架依赖
- `contest/runtime`
  - 负责：只为 command service 注入真实需要的依赖
  - 不负责：继续把无效基础设施参数传进 application 层

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-contest-challenge-redis-dead-dep-phase5-slice2-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-challenge-redis-dead-dep-phase5-slice2.md`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 删除 challenge command service 的死 Redis 依赖

目标：

- service 结构体和构造函数不再出现 Redis client
- runtime / tests 同步改签名

Validation:

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'ChallengeService(Add|Update|Remove)|Context' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus:

- challenge command service 是否真的没有剩余 Redis 使用点
- runtime 是否只删除无效参数，没有漏掉真实依赖

### Slice 2: 删除 allowlist 例外并同步设计稿

目标：

- 删掉 `contest/application/commands/challenge_service.go -> github.com/redis/go-redis/v9`
- 在 phase 5 设计稿里追加当前进展

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 只删除本次真正收口的 allowlist
- 文档表述是否准确反映“删掉死依赖”，没有夸大为更大范围的 Redis 收口

## Risks

- 如果还有隐藏调用依赖旧构造签名，会在 runtime 或测试编译阶段暴露
- 如果误删真实字段，AWD challenge 关联逻辑可能回归

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'ChallengeService(Add|Update|Remove)|Context' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- 这刀不引入新抽象，只删除一条已经失效的 concrete dependency
- 结构收益明确：runtime 不再把无效 Redis 参数注入 application command service
- touched surface 上的已知 debt 就是这条例外，本切片直接删掉，不留下兼容层
