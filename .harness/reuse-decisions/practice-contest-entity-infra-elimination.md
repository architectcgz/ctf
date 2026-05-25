# Reuse Decision

## Change type
repository / mapper / persistence row / architecture guardrail

## Existing code searched
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/contest/entity/*.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `.harness/reuse-decisions/practice-contest-record-localization.md`
- `docs/plan/archive/impl-plan/2026-05-18-practice-contest-record-localization-implementation-plan.md`

## Similar implementations found
- `practice/infrastructure/repository.go`
  - 已经把跨模块输出形状收口成 `practiceports.*Record`，剩下的问题主要是 infra 仍直接拿 `contestentity` 当 ORM row。
- `practice/infrastructure/timeline_query_repository.go`
  - 已经在模块内定义 query row，再转成本地 record，适合沿用同样思路。
- `contest/entity/*.go`
  - 仍然是 owner 模块的持久化事实源，本刀只复制消费侧所需最小 row / snapshot 形状，不改 owner。

## Decision
refactor_existing

## Reason
上一刀已经把 `practice` 上层 contract 从 `contest/entity` 收到了本地 `Record`，但 `practice/infrastructure`
还直接依赖 `contestentity` 做 ORM 查询、锁和 snapshot 解析，导致模块边界仍靠 allowlist 兜着。

这刀继续沿用“消费侧本地 row + 本地 decode helper”的做法：

- 在 `practice/infrastructure` 内定义只服务本模块的 contest row
- `repository.go`、`score_repository.go` 改查本地 row，不再 import `contest/entity`
- `contest_awd_runtime_subject_mapper.go` 改围绕 `practiceports.ContestAWDServiceRecord` 和本地 snapshot decode
- 删掉 `practice -> contest/entity` 的剩余架构 allowlist

这样 `contest` 仍保留 owner 身份，但 `practice` 非测试代码不再直接依赖 owner 私有持久化实现。

## Files to modify
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_runtime_subject_mapper.go`
- `code/backend/internal/module/practice/infrastructure/score_repository.go`
- `code/backend/internal/module/practice/infrastructure/contest_persistence_rows.go`
- `code/backend/internal/module/practice/infrastructure/contest_awd_service_snapshot.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/plan/archive/impl-plan/2026-05-18-practice-contest-entity-infra-elimination-implementation-plan.md`

## After implementation
- `practice` 非测试 `infrastructure` 不再 import `ctf-platform/internal/module/contest/entity`
- `contest_awd_runtime_subject_mapper.go` 的 snapshot 解析留在 `practice` 本地 helper
- `architecture_allowlist_test.go` 不再保留 `practice -> contest/entity` 例外
