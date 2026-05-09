# Contest 列表排序 owner 收口独立 Review

## Review Target
- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-review-pagination`
- Branch: `feat/awd-review-pagination`
- Diff source: working tree vs `HEAD` (`a86f205e`)
- Review mode: re-review against the latest scoped working-tree diff
- Files reviewed:
  - `AGENTS.md`
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/dto/contest.go`
  - `code/backend/internal/module/contest/ports/contest.go`
  - `code/backend/internal/module/contest/api/http/contest_query_handler.go`
  - `code/backend/internal/module/contest/application/queries/contest_list_query.go`
  - `code/backend/internal/module/contest/infrastructure/contest_repository.go`
  - `code/backend/internal/module/contest/application/queries/contest_service_test.go`
  - `code/frontend/src/api/contest.ts`
  - `code/frontend/src/api/admin/contests.ts`
  - `code/frontend/src/api/__tests__/contest.test.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `docs/plan/impl-plan/2026-05-09-contest-列表排序-owner-收口-implementation-plan.md`
  - `works/harness-good-practices.md`
- Necessary context read:
  - `docs/reviews/backend/README.md`

## Classification Check
- 结论：同意当前切片按 non-trivial / 独立 review gate 处理。
- 依据：
  - 改动跨 `ports -> application -> repository -> tests -> harness docs`
  - 目标不是局部修 bug，而是收口内部 filter/sort contract owner 与表示方式
  - `AGENTS.md` 已把这类 `API / filter / sort / pagination` contract 改动明确提升为结构性 review 面

## Gate Verdict
- `pass`

## Findings
- 无未解决的 material findings。
- 本轮最终 re-review 确认，本次切片中出现过的 blocker 已全部收口：
  - `ContestListSort` / `ContestListFilter` 已收成带未导出 marker 的 sealed interface，并通过 `NewContestListFilter(...)` 强制注入非空 sort；当前 scoped diff 中不存在“漏填导出 `Sort` 字段后由 repository 静默给默认值”的通路。见 `code/backend/internal/module/contest/ports/contest.go:34-119`、`code/backend/internal/module/contest/application/queries/contest_list_query.go:52-74`。
  - `statuses/mode` 的 normalize owner 已回收到 handler trust boundary；application 不再重复整理 `statuses`，repository 也不再 `TrimSpace()` 或隐式修正 `mode`。见 `code/backend/internal/module/contest/api/http/contest_query_handler.go:24-120`、`code/backend/internal/module/contest/infrastructure/contest_repository.go:125-132`。
  - 测试已从“比较导出 struct 值”切到“通过受控 accessor 断言收口后的内部语义”，并同时覆盖 list/summary 两条查询路径。见 `code/backend/internal/module/contest/application/queries/contest_service_test.go:82-134`。
  - 跨端 summary key 漂移已修复：前端 raw contract 统一读取 `registering_count`，学生端与管理员端 API normalize 都已对齐；后端 full-router 集成测试也新增了对原始 JSON key 的断言，不再只靠 Go DTO 自解码回正确字段。
  - 管理员 contest API contract test 已显式断言 `statuses / mode / sort_key / sort_order / signal`，避免再依赖 matcher 对缺省 `undefined` 字段的宽松语义。

## Material Findings
- 无 material findings。

## Senior Implementation Assessment
- 当前实现已经达到这次切片的目标边界：
  - `handler` 明确承担 `status/statuses/mode` 的 trust-boundary 解析与校验。
  - `application` 只承担排序 `normalize/default` owner，并通过 `NewContestListFilter(...)` 产出受控 filter。
  - `repository` 只消费 sealed filter 并把 sort 映射到固定 SQL 片段，不再承担二次 normalize/default 语义。
  - `frontend api layer` 明确承担后端 raw status / summary contract 到前端内部语义的映射，`registration -> registering` 与 `registering_count` 的 contract 现在有单测和后端原始 JSON 断言共同约束。
- 以当前 review scope 来看，这是比前一版更简单、owner 更清晰、误用面更小的实现，没有看到需要为了“更优雅”而继续扩大改动面的必要。

## Required Re-validation
- 本轮 blocker 修复后的受影响验证已具备最小充分证据：
  - `go test ./internal/app -run TestFullRouter_AdminContestListSupportsModeStatusesSortAndSummary`
  - `pnpm test:run src/api/__tests__/contest.test.ts src/api/__tests__/admin.test.ts src/views/contests/__tests__/ContestList.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/ContestOperationsHub.test.ts`
  - `pnpm test:run src/api/__tests__/admin.test.ts`
  - `pnpm typecheck`
  - `git diff --check -- code/backend/internal/app/full_router_state_matrix_integration_test.go code/frontend/src/api/contest.ts code/frontend/src/api/admin/contests.ts code/frontend/src/api/__tests__/contest.test.ts code/frontend/src/api/__tests__/admin.test.ts`

## Residual Risk
- 本次 review 只覆盖上述文件的当前工作树 diff，没有评审工作树中的其他未提交改动。
- 评审本身没有独立执行测试；测试通过结论依赖主执行链路提供的实际运行结果。
- `NewContestListFilter(...)` 对 `nil sort` 采用 fail-fast `panic` 保护编程期误用；当前 scoped diff 的生产路径不会命中该分支，但本轮仍未新增专门覆盖该 invariant 的测试。
- `registration` 与 `registering` 这组命名差异仍然横跨后端领域语义与前端展示语义，不过当前已经由前端 API normalize 和后端原始 JSON 测试共同兜住，不构成当前 blocker。

## Touched Known-debt Status
- 状态：`touched, closed for this slice`
- 说明：
  - 这次改动直接触达了 `AGENTS.md` 与 implementation plan 中已经显式记录的 contract owner / opaque representation 债务面。
  - 在当前 review scope 内，前一版独立 review 标出的“可绕过 application 的 sort 默认值 owner”与“同一 filter contract 跨层重复 normalize”都已经收口完成，未再看到同类漏口残留在 touched surface 上。
