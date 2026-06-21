---
name: ctf-backend-patterns
description: >
  Use this skill for CTF backend (Go) work in this repo when the task involves
  refactoring or splitting an over-long mixed-responsibility file, writing or
  reviewing background workers / dispatchers / pollers / retry-backoff loops,
  outbox / inbox / durable-journal state transitions, idempotency under multiple
  instances or stale retries, auth / session / security-path correctness,
  runtime node health vs schedulable semantics, constructors or factories that
  touch external runtime state, completion-timestamp closures, or backend test
  layering and assertion ownership. Activate even when the user only says
  "重构这个文件 / 拆一下 / 这个 worker 有并发问题 / 改密后没登出 / 测试一直挂"
  without naming a pattern.
---

# CTF Backend Patterns

项目后端已验证的 pattern 索引（导航中心）。本文件只给一句话 + ✓Check + feedback 路径；
详细模板和案例在对应 `feedback/*.md` 里，命中后再按需读。

## Use When
- 单文件过长 / 职责混合，需要拆分或重构。
- 写或 review 后台 worker / dispatcher / poller / reconcile / retry 循环。
- outbox / inbox / saga 等 durable journal 的状态迁移、多实例并发、stale 重试。
- 认证 / 会话 / 权限这类安全路径的正确性。
- service / bootstrap / composition 中构造对象时读取 OS、环境、文件、网络、时间、随机数或真实 client。
- runtime 节点健康、调度、清理生命周期语义。
- 后端测试分层与断言归属。

## Do Not Use
- 纯前端 / UI / 主题工作 → `ctf-ui-theme-system`、`ctf-dark-surface-alignment`。
- 通用 Go 架构与分层方法（非本仓事实）→ `go-backend`、`onion-clean-architecture`、
  `ctf-go-backend-architecture-reference`。
- 项目硬契约（UTC 时间、context 传递、破坏性 DB 安全）→ 以 `AGENTS.md` 的
  Backend Contracts / Destructive Database Safety 为准，本 skill 不复制。

## Common Tasks
- 拆分大文件 → 读 Refactoring §，按 `feedback/2026-06-12-package-split-by-responsibility-template.md`。
- 改后台循环 / 重试 → 读 Concurrency & Durable State §，逐条过 ✓Check。
- 改 goroutine / panic recovery / SafeGo 风格包装 → 读 Concurrency & Durable State §（panic owner 边界）。
- 改 service 构造、factory、bootstrap 接线，或 hostname / env / file / network / time / random lookup → 读 Construction & External Effects §。
- 改 outbox / 状态机 → 读 Concurrency & Durable State §（CAS 迁移）。
- 改 auth / session → 读 Security Path §。
- 改 runtime / 实例生命周期 → 读 Runtime Semantics §。
- 写 / 修后端测试 → 读 Test Discipline §（先读 `code/backend/tests/README.md` 定层级）。
- 不在列表内 → 先读本文件，再按 `feedback/` 文件名匹配最近案例。

## Refactoring & Structure
- **同包按职责拆分**：单文件超 800–1000 行且职责混合时，在**同包内**拆成
  types/load/validate/defaults/domain-logic 等文件，不新建 package，导出 API 不变。
  ✓Check：调用方 import 与类型名是否零改动？是否先有结构护栏测试再拆？
  → `feedback/2026-06-12-package-split-by-responsibility-template.md`
- **完成态用局部 finish() 闭包收口**：多 early-return 都要写 `FinishedAt`/`Duration` 时，
  用一个 `finish()` 闭包集中写；对外时间戳归一 UTC，Duration 保留原始 monotonic 配对。
  ✓Check：新增分支只需调 `finish()` 吗？Duration 是否误用 `time.Now().UTC()` 做差（应保留原始配对）？
  → `feedback/2026-06-08-go-finish-closure-for-result-timestamps.md`

## Construction & External Effects
- **构造函数不隐藏外部依赖读取和副作用**：普通 `NewService` / `New...` 只接收已解析依赖、
  配置和 identity；不在内部读取 hostname、env、文件、证书、网络、时间、随机数或真实 client。
  这些外部事实由 bootstrap、composition、显式 factory，或名称已表达副作用的 `Load*` / `Dial*` / `Open*`
  owner 读取并处理错误后传入。
  ✓Check：这个构造函数是否调用了 `os.*`、`net.*`、文件/env 读取、`time.Now`、随机数或真实 dial/open？
  如果是，函数名和 owner 是否明确表达副作用，失败是否可观察？
  → `feedback/2026-06-21-constructors-should-not-hide-external-effects.md`

## Concurrency & Durable State
- **发布幂等 ≠ 状态迁移幂等**：DB outbox/journal 的 `sent/failed` 迁移必须是
  compare-and-set（`WHERE status = pending`），不能按 id 盲写，否则慢实例会把已 sent 复活成 pending。
  ✓Check：状态迁移是 CAS 吗？`ListPending` 只读不 claim 吗？只看了"发布幂等"是不是漏了"行状态幂等"？
  → `feedback/2026-06-07-contest-realtime-outbox-worker-must-not-revert-sent-rows-on-stale-retries.md`
- **事件只做唤醒，durable state 是事实源**：进程内事件总线只用于低延迟唤醒，
  必须保留可重扫、可恢复的 DB 状态兜底（如 `status=stopping` 扫描）；多副本用 Redis lock 防并发。
  ✓Check：事件丢失 / 进程重启后还能从 DB 状态恢复吗？长耗时副作用是否被塞进 HTTP 请求上下文？
  → `feedback/2026-06-09-event-wakeup-keeps-durable-state-owner.md`
- **控制流语义不挂在可选观测组件上**：失败退避 / 取消 / 重试节奏由循环本身决定，
  不能写成 `err != nil && logger != nil` 这种把 backoff 绑到 logger/metrics 存在性上。
  ✓Check：把 logger/metric/debug 设为 nil，pacing/retry/cancel 行为是否完全不变？
  → `feedback/2026-06-07-retry-backoff-should-not-depend-on-logger-presence.md`
- **goroutine panic 语义由 owner 决定**：不把后台 goroutine 默认套进共享
  `SafeGo` / recover helper；先明确启动 owner、取消/等待 owner、错误传播和 panic 后果。
  HTTP 边界、root background job、cron/reconcile 单轮任务、业务状态异步任务和
  `WaitGroup.Wait()` 辅助 goroutine 的正确处理不同，不能用一个 helper 吞掉差异。
  ✓Check：panic 后是 re-panic、业务失败、单轮失败重试，还是请求 500？测试锁的是行为边界还是 helper 导入？
  → `feedback/2026-06-20-goroutine-panic-owner-boundary.md`

  推荐：root 级关键后台任务由 root owner 本地记录后重新抛出，避免后台能力静默死亡。

  ```go
  wg.Add(1)
  go func() {
      defer wg.Done()
      defer func() {
          if recovered := recover(); recovered != nil {
              logger.Error("background_job_panicked",
                  zap.Any("panic", recovered),
                  zap.String("task_name", name),
                  zap.ByteString("stack", debug.Stack()),
              )
              panic(recovered)
          }
      }()
      run(runCtx)
  }()
  ```

  推荐：只用于等待 `WaitGroup.Wait()` 的 goroutine 保持最小裸等待，不挂 recover 语义。

  ```go
  done := make(chan struct{})
  go func() {
      wg.Wait()
      close(done)
  }()
  ```

  推荐：有业务失败状态的异步任务在业务 owner 内 recover，并写入业务失败结果。

  ```go
  go func() {
      defer tasks.Done()
      defer func() {
          if recovered := recover(); recovered != nil {
              service.markFailed(taskCtx, id, fmt.Errorf("report task panicked: %v", recovered))
          }
      }()
      if err := run(ctx); err != nil {
          service.markFailed(ctx, id, err)
      }
  }()
  ```

  禁止：用共享 helper 默认吞掉所有后台 panic，尤其不要用 no-op logger / inert ctx 兜底掩盖接线错误。

  ```go
  safego.Go(&wg, ctx, logger, "task_name", func(ctx context.Context) {
      run(ctx)
  })
  ```

## Security Path
- **安全语义放主鉴权链路，不依赖清理辅助结构**：用户级会话失效用 session version 在
  `GetSession` 直接判定；反向索引只做物理清理优化。`cleanup succeeded` 不等于 `security guarantee`。
  ✓Check：没有新辅助索引的历史 / 跨发布 session，能在主链路被判为失效吗？handler 是否只在撤销语义成立后才返回成功？
  → `feedback/2026-06-02-auth-session-revocation-must-not-depend-on-cleanup-index.md`

## Runtime Semantics
- **schedulable=false ≠ 节点不健康**：cordon 只表示"不再接收新调度"，
  不应被用来过滤旧容器操作、健康扫描或 failover，否则健康但被 cordon 的节点会被误判不可用且不触发 failover。
  ✓Check：node-bound 操作（清理 / checker / SSH / 文件写）是否错误地用 schedulable 当可用性判据？
  → `feedback/2026-06-13-runtime-node-schedulable-is-not-execution-health.md`
- **多容器题本地验题要等 ready 不是等 running**：`docker compose up -d` 只保证 running；
  攻击链 `public -> internal-app -> internal-data` 需显式等待依赖 ready，checker 无限重试只算兜底。
  → `feedback/2026-05-10-awd-topology-local-readiness.md`

## Test Discipline
- **测试按 owner 分层**：模块层测渲染 / 业务细节，full-router / `internal/app` 只兜集成事实
  （路由可达、权限、状态流、下载元信息）；PDF 等断言优先断稳定 token，不绑脆弱文案；同类 helper 能力对齐。
  ✓Check：集成层测试是否越界去断 renderer 文案？断言是否绑定了会变的语言 / 措辞？
  → `feedback/2026-06-03-backend-test-layering-and-pdf-assertion-ownership.md`
- **不为通过测试而迁就坏架构**：测试 / mock / fixture 离代码最近、改动成本低，但测试变绿不等于完成；
  失败时先判 owner / contract / 语义是否未收口，不放宽断言、不改 fixture 迁就错误实现。
  ✓Check：这个测试是不是只有放宽断言 / 改 mock 才能过？根因是不是契约 owner 没定？
  → `feedback/2026-05-09-do-not-bend-tests-to-fit-broken-architecture.md`

## Known Gotchas（最高价值，命中即停）
- 多实例 worker 的状态迁移不是 CAS → 状态复活、重复广播。见 Concurrency & Durable State §。
- backoff / retry 绑在 logger 存在性上 → logger 为 nil 时退化成热循环。见 Concurrency & Durable State §。
- 共享 SafeGo 默认吞 panic → 关键后台任务可能静默死亡，业务失败状态也可能丢失。见 Concurrency & Durable State §。
- 构造函数里偷偷读取外部世界并忽略错误 → 测试和运维都看不到失败 owner。见 Construction & External Effects §。
- 安全撤销建立在 best-effort 清理上 → 其他设备旧 session 仍有效。见 Security Path §。

## 添加新 Pattern
当 `feedback/` 出现通过 2/3 录入标准（可重复 / 代价高 / 代码不可见）的后端 pattern 时：
1. 归到对应主题，加 5–10 行索引：一句话 + ✓Check + feedback 路径。
2. 若是高代价陷阱，同时在 Known Gotchas 加一行锚点（激活优于存储）。
3. 泛化成脱离当前任务也能看懂的表述，不要抄会话叙事。
