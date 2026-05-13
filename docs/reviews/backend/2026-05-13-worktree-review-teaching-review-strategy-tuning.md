# 2026-05-13 教学复盘策略调优 Worktree Review

## Review Target

- repository: `ctf`
- branch/worktree: `feat/teaching-review-strategy-tuning-wt`
- diff source: local worktree vs `main`
- files reviewed:
  - `code/backend/internal/teaching/advice/advice.go`
  - `code/backend/internal/teaching/advice/advice_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/infrastructure/repository_test.go`
  - `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
  - `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`
  - `code/backend/cmd/seed-teaching-review-data/main.go`
  - `code/backend/cmd/seed-teaching-review-data/main_test.go`

## Classification Check

- agree with pipeline classification: non-trivial backend + seed + docs slice

## Gate Verdict

- `pass`

## Findings

- 无 findings。

## Material Findings

- 无。

## Senior Implementation Assessment

- 推荐维度归因、班级流程型建议去挂题、seed 稀疏分类回退这几处实现方向是对的，owner 也基本收口到了 `internal/teaching/advice`、`assessment` 和 `teaching_readmodel` 各自应负责的位置。
- review archive 活跃事实现在已经和已有事实源对齐到 `timeline / evidence / writeups / manualReviews` 四类输入，`low_activity` 这条规则不再漏算最近的人工评阅训练。
- 另外，`code/backend/cmd/seed-teaching-review-data/main.go` 已增长到 1766 行，本次把 coverage scenario 生成、fallback 维度选择和 coverage summary 都继续堆进同一文件，短期可用，但后续再扩 seed 规则时应优先拆出独立 builder / summary 文件，避免 seed 命令继续膨胀成难 review 的单体入口。

## Required Re-validation

- 已完成：
  - `go test ./internal/module/assessment/application/commands ./internal/teaching/advice -count=1 -timeout 180s`
  - `go vet ./internal/module/assessment/application/commands ./internal/teaching/advice`
  - `python3 scripts/check-docs-consistency.py`
  - `bash scripts/check-consistency.sh`

## Validation Evidence

- `git diff --check`
- `go test ./cmd/seed-teaching-review-data ./internal/module/assessment/application/commands ./internal/teaching/advice ./internal/module/assessment/application/queries ./internal/module/challenge/infrastructure ./internal/module/teaching_readmodel/application/queries -count=1 -timeout 180s`
- `go vet ./cmd/seed-teaching-review-data ./internal/module/assessment/application/commands ./internal/teaching/advice ./internal/module/assessment/application/queries ./internal/module/challenge/infrastructure ./internal/module/teaching_readmodel/application/queries`
- `go test ./internal/module/assessment/application/commands ./internal/teaching/advice -count=1 -timeout 180s`
- `go vet ./internal/module/assessment/application/commands ./internal/teaching/advice`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

## Residual Risks

- `bash scripts/check-workflow-complete.sh` 仍未全绿，原因不是本次后端修改，而是本地缺少 `vitest`，前端架构守卫无法完成。
- 本次 review 只覆盖当前 worktree diff，没有重新跑 seed 导入后的整条业务联调。

## Touched Known-Debt Status

- `code/backend/internal/teaching/advice/advice.go` 和 `code/backend/cmd/seed-teaching-review-data/main.go` 都已经是偏大的 owner surface；本次没有继续做结构拆分，只在 touched surface 内补了必要回归和健壮性修正。
- 其中 `main.go` 的体量增长已达到需要后续拆分的程度，但当前未发现因此直接造成新的 correctness bug。
