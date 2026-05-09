# Contest 列表排序 owner 收口实施计划

## Plan Summary
- Objective
  - 收口 `contest` 列表排序参数的 `normalize / default / validate` owner，避免 application 与 repository 两层同时兜底。
  - 让 `ContestListFilter` 不再把排序语义作为裸字符串在内部层间传递，repository 只做已收敛排序语义到 SQL 列/方向的映射。
  - 把内部排序 contract 收口成 opaque value object，避免 touched surface 继续暴露“可手工伪造，再靠下游 panic 或 defensive branch 发现”的无效状态。
  - 补充坏品味记录，明确“跨层重复归一化 + repo 接未收敛裸字符串”是需要被 harness 提前拦截的反模式。
- Non-goals
  - 不调整对外 HTTP query 参数形态；`sort_key` / `sort_order` 的 API 契约保持不变。
  - 不扩散到其他模块的 filter contract，只收口本次触达的 contest 列表查询。
  - 不改 contest 列表以外的排序能力。
- Source architecture or design docs
  - `AGENTS.md`
  - `docs/plan/impl-plan/2026-05-09-awd-review-index-pagination-implementation-plan.md`
  - `docs/reviews/general/2026-05-09-contest-scoreboard-pagination-review.md`

## Ownership Decision
- `handler`
  - 负责 HTTP trust boundary 上的 `status/statuses/mode` 解析、trim、去重和白名单约束；不承担排序默认值决策。
- `application`
  - 唯一 owner：负责把 `ListContestsInput.SortKey / SortOrder` 归一化成内部受限排序语义，并设置默认值。
- `repository`
  - 只负责消费已收敛 filter，并把内部排序语义映射成固定 SQL 片段；不再对 `statuses/mode/sort` 做 trim、去重、默认化或“发现调用方手工伪造无效状态”的主要防线。

## Task 1
- Goal
  - 为 `ContestListFilter` 引入 opaque 排序 value object，替代 `SortKey string` 和 `SortOrder string`，同时避免保留可随意拼装的导出 enum。
- Touched modules or boundaries
  - `code/backend/internal/module/contest/ports`
  - `code/backend/internal/module/contest/application/queries`
- Dependencies
  - 依赖现有 handler 入参绑定；对外 API 无需改动。
- Validation
  - `go test ./internal/module/contest/application/queries -run TestContestService`
- Review focus
  - 默认值是否只在 application 出现一次
  - `statuses/mode` 是否已经明确留在 handler trust boundary，而不是继续在 application/repository 重复整理
  - 测试是否直接断言受控排序值，而不是继续依赖字符串或导出枚举兜底
- Risk notes
  - 若仍保留字符串 filter 字段，repository 后续很容易重新长出第二个 normalize 点。
  - 若只把字符串换成导出 enum，但 invalid state 仍可手工构造，下游就会重新长出 panic 或 defensive fallback。

## Task 2
- Goal
  - 删除 repository 层对 contest 列表排序的默认化逻辑，只保留 opaque sort 到 SQL 列/方向的确定映射。
- Touched modules or boundaries
  - `code/backend/internal/module/contest/infrastructure`
- Dependencies
  - 依赖 Task 1 先把 filter contract 收敛为受限类型。
- Validation
  - `go test ./internal/module/contest/application/queries ./internal/module/contest/api/http ./internal/app -run 'TestContestService|TestUpdateContestSkipsReadinessAuditPayloadWhenCommandFailsBeforeGate|TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary'`
- Review focus
  - SQL 排序字段是否仍然只来自固定白名单
  - repository 是否完全去掉默认值 owner
  - repository 是否不再需要通过 panic 或 fallback 来兜住可伪造的无效状态
- Risk notes
  - 若映射函数仍携带 fallback 语义，只是把双重兜底换成了隐式兜底，问题并没有真正解决。

## Task 3
- Goal
  - 在 harness 归档里记录本次坏品味，明确其识别信号和应对动作。
- Touched modules or boundaries
  - `works/harness-good-practices.md`
- Dependencies
  - 依赖 Task 1/2 先完成 owner 收口，避免只记录问题不修正。
- Validation
  - `bash scripts/check-consistency.sh`
- Review focus
  - 记录是否准确描述反模式，而不是把具体实现细节误写成全局规则
- Risk notes
  - 若只在对话里说明而不落盘，后续 agent 仍会重复犯同类错误。

## Integration Checks
- `ContestService` 生成的 `ContestListFilter` 中排序字段为受限类型，默认值只在 application 出现一次。
- `Repository.List` 不再接收裸字符串排序键，也不再对排序做二次默认化。
- `statuses/mode` 的 normalize 只保留一个明确 owner，repository 不再隐式 trim 或修正。
- touched surface 上不存在“外部可拼出无效 sort，再由 repository 晚发现”的内部 contract 漏口。
- 现有 admin contest list query 仍能按 `sort_key=start_time&sort_order=desc` 正常返回。

## Rollback / Recovery Notes
- 本次仅涉及 Go 代码和 harness 文档，无 migration、配置或数据回填。
- 若回退，需要整体回退 contract、repository 映射和测试，避免回到“application 已改枚举，repository 仍按字符串兜底”的中间态。

## Plan Review
- 非 trivial，原因是：
  - 改动跨 `ports -> application -> repository -> tests -> harness docs`
  - 直接触达已经被用户指出的结构性 owner 混杂面
- 方案可执行性结论：
  - opaque 排序 value object 是最小可行收口面，不改变 HTTP 对外契约，只修内部 contract
  - 这次切片同时解决行为正确性与结构收敛，不把“重复 normalize/default”或“可伪造 invalid state”留作 follow-up
