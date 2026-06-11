# Teaching Fact Semantics Unification Review

- Review Target：`/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-teaching-fact-semantics-unification`
- Task Slug：`2026-06-11-teaching-fact-semantics-unification`
- Plan：`docs/plan/impl-plan/2026-06-11-teaching-fact-semantics-unification-implementation-plan.md`
- Review Scope：
  - `code/backend/internal/module/assessment/application/commands/profile_service.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
  - `code/backend/internal/module/assessment/infrastructure/repository.go`
  - `code/backend/internal/module/contest/application/commands/submission_*.go`
  - `code/backend/internal/module/contest/contracts/events.go`
  - `code/backend/internal/module/teaching_query/infrastructure/repository.go`
  - `docs/architecture/features/教学复盘建议生成架构.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/plan/impl-plan/2026-06-11-teaching-fact-semantics-unification-implementation-plan.md`

## Classification Check

- 结论：同意当前任务属于 `非琐碎任务`。
- 原因：这次改动同时触达 `assessment`、`teaching_query`、`contest` 事件链和架构事实源，属于会改变 owner、事实口径和回归面的一次结构性后端改动。

## Gate Verdict

- Verdict：`blocked`
- 原因：`code-workflow` 要求的独立 reviewer gate 尚未满足。
  - 第一次 reviewer subagent 调用返回外部 `403 Forbidden`，原因是上游额度不足。
  - 第二次 reviewer subagent 调用长时间无最终结果，随后主动 shutdown，未产出独立 review 结论。
  - 当前文档只记录 same-context self-check 结果，不能替代独立 gate。

## Findings

1. `Blocker`：独立 review gate 未完成。
   - 位置：`code-workflow` 独立 review 阶段，非具体代码文件。
   - 风险：当前只有实现上下文内的 self-check，按仓库 workflow 不能把它当成最终 completion review。
   - 结论：在获得可用 reviewer subagent 或其他真正独立上下文之前，这轮任务不能宣称完成所有 gate。

## Material Findings

- 独立 reviewer subagent 不可用，导致 `code-workflow` 最终 gate 未满足。

## Senior Implementation Assessment

- 从当前代码看，recommendation / class review 的 owner 已经回到 teaching snapshot + `internal/teaching/advice`，这是比原来“repository 事实一套、recommendation 自己再判一套”更低风险的实现。
- snapshot owner 继续留在 `assessment` / `teaching_query` repository，本次只收口 live 语义和最小事件链，没有顺手引入新的持久化表或并行 service，这个范围控制是合理的。
- self-check 过程中发现 `contest.flag_accepted` 的 publish 失败原本会被静默吞掉；该点已在实现上下文内修正为 warn 日志，和 practice / AWD 的弱事件发布模式对齐。

## Validation Evidence Reviewed

- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/module/contest/application/commands/... -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-backend-architecture.sh --full`
- `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## Required Re-Validation

- 在独立 reviewer 可用后，重新执行一次真正的 gate review。
- 本次 self-check 后新增的可观测性修正已经补跑：
  - `go test ./internal/module/contest/application/commands/... -count=1`
  - `bash scripts/check-backend-architecture.sh --full`
  - `git diff --check`

## Residual Risk

- same-context self-check 未发现新的 correctness blocker。
- 但在没有独立 reviewer 的情况下，仍然存在 owner 漂移、语义遗漏或测试盲区未被第二视角发现的风险。

## Touched Known-Debt Status

- touched known debt：`docs/todos/2026-05-17-project-tech-debt-from-migrations.md` 里“recommendation / class review 与 practice 语义仍未完全统一”。
- self-check 结论：本次实现已经在 touched surface 内把这条债收口，活动 backlog 已移除；当前剩余开放项只剩 `challenges.image_id = 0`。
