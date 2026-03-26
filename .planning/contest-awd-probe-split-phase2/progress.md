# Progress

## 2026-03-26

- 启动 `contest-awd-probe-split-phase2`，目标是继续拆 `contest` AWD probe 主流程文件。
- 盘点确认 `application/jobs/awd_probe.go` 同时承载多类职责：
  - probe 执行与 HTTP 请求逻辑
  - health check URL 构造
  - error normalize 与消息清洗
  - timeout / health path normalize helper
- 已完成文件拆分：
  - `awd_probe.go` 保留 probe 主流程与 probe 结果结构
  - `awd_probe_support.go` 承载 URL / error / timeout helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
