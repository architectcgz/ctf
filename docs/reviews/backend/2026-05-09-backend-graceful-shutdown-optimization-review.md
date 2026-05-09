# Backend Graceful Shutdown Optimization Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf-jeopardy-container-verification`
- Branch: `fix/jeopardy-container-verification-boundary`
- Diff source: working tree changes on 2026-05-09
- Files reviewed:
  - `code/backend/internal/bootstrap/run.go`
  - `code/backend/internal/bootstrap/run_test.go`
  - `code/backend/internal/app/http_server.go`
  - `code/backend/internal/app/http_server_test.go`
  - `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
  - `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime_internal_test.go`
  - `code/backend/scripts/dev-run.sh`

## Classification Check

- 认同本次变更为 non-trivial backend bugfix。
- 变更跨越启动脚本、bootstrap、应用关闭顺序和后台 job 日志语义，直接影响本地开发与 AWD listener 的真实关闭行为。

## Gate Verdict

- Pass

## Findings

- 无 material findings。

## Material Findings

- None.

## Senior Implementation Assessment

- 这次改动把同一条关闭链路上的 owner 拆得比较清楚：
  - `dev-run.sh` 只负责本地进程树的信号策略，不再和 Go 进程的 graceful 逻辑抢时序。
  - `bootstrap` 只负责进程级信号等待和 timeout shutdown 入口。
  - `HTTPServer` 负责应用内的关闭顺序，明确把 HTTP drain 启动点前置。
  - `AWDRoundUpdater` 只处理正常关停误报，不扩大到真实业务错误吞没。
- `--hot` 真实 smoke 证明了问题核心确实在“同时打断 wrapper 与 leaf process”。把 `SIGINT` 第一阶段只给叶子业务进程后，`http_server_stopped / postgres_closed / redis_closed` 可以稳定出现，`:2222` 也能释放。
- `startHTTPShutdown()` 先等待“drain 已启动”这个同步点是必要的；否则即使脚本信号策略正确，listener 释放仍可能偏后。

## Required Re-validation

- `cd code/backend && go test -run 'TestAWDRoundUpdaterRefreshesSchedulerLockWhileRunning|TestAWDRoundUpdaterSyncContestRoundsSkipsCanceledContextErrorLog' ./internal/module/contest/application/jobs -v`
- `cd code/backend && go test -run 'TestHTTPServer.*|TestNewHTTPServerBuildsAndShutsDown' ./internal/app -v`
- `cd code/backend && go test ./internal/bootstrap -v`
- `cd code/backend && bash -n scripts/dev-run.sh`
- 真实 `--hot` smoke：启动后发送 `SIGINT`，确认 `http_server_stopped`、`postgres_closed`、`redis_closed` 出现，且 `ss -ltnp '( sport = :2222 )'` 为空。

## Residual Risk

- `timeout 120 bash scripts/check-workflow-complete.sh` 未全绿，失败原因是环境缺少前端测试依赖 `vitest`，不是本次后端关停改动的功能回归：
  - `sh: 1: vitest: not found`
- 当前 review 为同会话切换到 code review 心智完成，未使用独立 subagent。原因是当前 turn 未获得额外代理委派授权。

## Touched Known-debt Status

- 本次 touched surface 上“正常关闭被误记成错误日志”的噪声已经收口，没有继续遗留在 `AWDRoundUpdater`。
- 关闭链路中的脚本层和 Go 层职责边界比之前更清晰，没有继续把进程树信号和应用内 shutdown 顺序混在一起。
