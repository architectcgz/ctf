# AWD Checker Preview 就绪等待修复计划

## 目标

- 修复 `http_standard` 类型 AWD checker preview 在临时实例刚启动时过早执行 `put_flag`，导致 `EOF / http_request_failed` 的问题。
- 让 `校园网盘` 与 `IoT 设备管理平台` 这类单容器 HTTP 题在 preview 链路上能等到最小就绪状态后再开始正式检查。

## 非目标

- 不修改题包源码、checker 配置、前端交互或 preview token 持久化协议。
- 不重写 preview 三轮聚合策略，不把无限重试塞进 checker action。
- 不处理 `tcp_standard` / `script_checker` 的额外行为变化。

## 输入事实

- `feedback/2026-05-10-awd-topology-local-readiness.md`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- 2026-05-31 实测 preview 返回：
  - `Put "http://host-gateway.internal:30001/api/flag": EOF`
  - `preview_summary=0/3 通过`

## 方案

1. 在 `AWDRoundUpdater` 的 `http_standard` preview 路径引入内部就绪等待。
2. 等待逻辑复用现有 `healthPath` 探测语义：
   - 优先探测配置里的 health path
   - 在短时间窗口内轮询，直到返回可用
   - 超时后继续走原 checker action，让最终错误仍由正式 action 产出
3. 保持 owner 单点：
   - command 层只负责 preview 编排
   - jobs 层负责 preview runtime readiness

## 预期改动面

- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/awd_probe_runtime.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_runtime_contract_test.go`

## 任务切片

### Slice 1：补失败测试

- 覆盖场景：preview 首次 health 探测失败，后续成功后才执行 `put_flag` / `get_flag`
- 验证：
  - `go test ./internal/module/contest/application/jobs -run TestAWDRoundUpdaterPreviewHTTPStandardWaitsForHealthBeforePutFlag -count=1`

### Slice 2：实现 preview readiness wait

- 在 `http_standard` preview 路径加入短轮询等待
- 保持其他 checker 类型行为不变
- 验证：
  - 同上测试转绿
  - 相关 jobs 包最小回归通过

### Slice 3：真实链路复测

- 对 `contest_id=9` 的 `service_id=27/28` 重跑 preview
- 观察 `preview_summary`、`service_status` 与 `error_code`
- 验证：
  - curl 真实 preview 返回不再死于首个 `PUT /api/flag EOF`

## 风险与回退

- 风险：等待窗口过长会放大 preview 时延。
- 控制：只做短窗口轮询，且仅用于 `http_standard` preview。
- 回退：删除 preview 前置等待 helper，恢复原有直接执行 action 的路径。
