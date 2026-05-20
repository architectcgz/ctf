# Backend Graceful Shutdown Optimization Implementation Plan

## Objective

修复本地后端在 `scripts/dev-run.sh --hot` 场景下的优雅关闭边界，要求：

- `Ctrl+C` 后 `awd_defense_ssh_gateway` 监听的 `:2222` 能稳定释放，不再残留端口占用。
- HTTP server 进入 shutdown 时，先启动 HTTP drain，再尽快停止后台 listener / job。
- `bootstrap` 层统一收口 shutdown 入口，避免散落的信号处理逻辑。
- 正常关停期间的 `context canceled` / `deadline exceeded` 不再被 `awd_round_updater` 误报成错误日志。

## Non-goals

- 不改 HTTP 路由、AWD 业务逻辑或题目导入链路。
- 不重做 `air` 热重载流程，也不替换本地开发启动工具。
- 不引入新的 runtime service 抽象，只在现有 `bootstrap`、`app`、jobs、脚本边界内收口。

## Inputs

- `code/backend/internal/bootstrap/run.go`
- `code/backend/internal/app/http_server.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- `code/backend/scripts/dev-run.sh`
- 用户反馈日志：`awd_defense_ssh_gateway: listen tcp :2222: bind: address already in use`

## Current Facts

- 旧链路里 `dev-run.sh` 在 `--hot` 模式下会同时向 `air` 及其子进程树打 `SIGINT`，和 Go 进程自己的 graceful shutdown 抢时序。
- `HTTPServer.Shutdown` 旧顺序更偏向“先等 HTTP 停完再停后台任务”，导致 `awd_defense_ssh_gateway` 这类 listener 释放偏晚。
- `bootstrap.Run()` 的 shutdown 流程散在 `Run()` 内部，没有统一 helper，也没有显式记录“停止完成”的成功路径。
- `AWDRoundUpdater` 在正常取消上下文时会把 `sync_awd_rounds_failed {"error":"context canceled"}` 打成 error。

## Chosen Direction

1. `bootstrap`
   - 用 `signal.NotifyContext` 收口信号监听。
   - 提取 `shutdownGracefully()` helper，统一等待信号、创建 timeout context、调用 server shutdown。
2. `app.HTTPServer`
   - `Shutdown()` 改成先启动 HTTP drain，再停止后台 jobs，最后关闭 app-level closers。
   - 为测试暴露最小钩子，证明 HTTP drain 启动点先于后台 job stop。
3. `AWDRoundUpdater`
   - 把 `context.Canceled` / `context.DeadlineExceeded` 识别为正常关停，不再记录 error。
4. `dev-run.sh`
   - 停机改成两阶段：
     - 先对叶子业务进程发 `SIGINT`，给 Go 进程完整 graceful window；
     - 再对剩余 wrapper 进程收尾，避免 `air` 提前把子进程抢杀。

## Ownership Boundary

- `code/backend/internal/bootstrap/run.go`
  - 负责进程级信号接收与 graceful shutdown timeout。
- `code/backend/internal/app/http_server.go`
  - 负责应用内 HTTP、后台 job、closer 的关闭顺序。
- `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
  - 负责 AWD round updater 在正常关停时的日志降噪。
- `code/backend/scripts/dev-run.sh`
  - 负责本地开发模式下的信号转发与进程树清理。

## Change Surface

- Modify: `code/backend/internal/bootstrap/run.go`
- Modify: `code/backend/internal/bootstrap/run_test.go`
- Modify: `code/backend/internal/app/http_server.go`
- Modify: `code/backend/internal/app/http_server_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime_internal_test.go`
- Modify: `code/backend/scripts/dev-run.sh`
- Create: `docs/reviews/backend/2026-05-09-backend-graceful-shutdown-optimization-review.md`

## Task Slices

### Slice 1: 收口 bootstrap 与 HTTP shutdown 顺序

目标：

- 引入统一 `shutdownGracefully()` helper
- 确保 HTTP drain 先启动，再停后台 listener / jobs

Validation:

- `cd code/backend && go test -run 'TestHTTPServer.*|TestNewHTTPServerBuildsAndShutsDown' ./internal/app -v`
- `cd code/backend && go test ./internal/bootstrap -v`

Review focus:

- timeout context 是否按预期创建并释放 signal handler
- 是否存在新的 goroutine 死等边界

### Slice 2: 收口 AWD 正常关停误报

目标：

- 正常取消上下文不再打印 `sync_awd_rounds_failed`
- 增加最小内部测试证明该边界

Validation:

- `cd code/backend && go test -run 'TestAWDRoundUpdaterRefreshesSchedulerLockWhileRunning|TestAWDRoundUpdaterSyncContestRoundsSkipsCanceledContextErrorLog' ./internal/module/contest/application/jobs -v`

Review focus:

- 是否只压正常关停噪声，不会吞掉真实业务错误

### Slice 3: 修正热重载脚本的停机信号策略

目标：

- `Ctrl+C` 时先给叶子业务进程 graceful 时间，再回收 `air` / wrapper
- 避免 `:2222` 再次因上次进程未收干净而占用

Validation:

- `cd code/backend && bash -n scripts/dev-run.sh`
- 真实 smoke：
  - `CTF_POSTGRES_HOST=127.0.0.1 CTF_POSTGRES_PORT=15432 CTF_POSTGRES_PASSWORD=postgres123456 CTF_REDIS_ADDR=127.0.0.1:16379 CTF_REDIS_PASSWORD=redis123456 bash ./scripts/dev-run.sh --hot`
  - 等待 `awd_defense_ssh_gateway_started`
  - 发送 `SIGINT`
  - 确认日志出现 `http_server_stopped`、`postgres_closed`、`redis_closed`
  - 确认日志不再出现 `sync_awd_rounds_failed {"error":"context canceled"}`
  - 确认 `ss -ltnp '( sport = :2222 )'` 为空

Review focus:

- `--hot` 下的 `air -> shell -> ctf-api` 进程树是否按预期回收
- 是否还会出现 wrapper 抢先杀掉业务进程的时序问题

## Risks

- 同时对 `air` 和叶子 Go 进程发 `SIGINT` 会和 graceful path 抢时序，导致看不到 `http_server_stopped`。
- `startHTTPShutdown()` 新增 goroutine 后，若没有同步“drain 已启动”点，可能让后台 listener 停止仍然偏晚。
- AWD round updater 的日志降噪若写得过宽，可能把真实业务错误误当成正常关停吞掉。

## Verification Plan

1. `cd code/backend && go test -run 'TestAWDRoundUpdaterRefreshesSchedulerLockWhileRunning|TestAWDRoundUpdaterSyncContestRoundsSkipsCanceledContextErrorLog' ./internal/module/contest/application/jobs -v`
2. `cd code/backend && go test -run 'TestHTTPServer.*|TestNewHTTPServerBuildsAndShutsDown' ./internal/app -v`
3. `cd code/backend && go test ./internal/bootstrap -v`
4. `cd code/backend && bash -n scripts/dev-run.sh`
5. 真实 `--hot` smoke，确认 `:2222` 释放、shutdown 成功日志存在、正常关闭误报消失
6. `timeout 120 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 边界明确：进程级信号在 `bootstrap`，应用关闭顺序在 `HTTPServer`，日志降噪在 AWD job，本地开发信号策略在脚本。
- 本次不是只修“端口偶尔占用”，而是同时收口关停顺序、脚本信号竞争、正常关停误报三条同一条链上的边界。
- touched surface 上的 `AWD round updater` 关闭误报被纳入同一切片，没有继续留作 follow-up。
