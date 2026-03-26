# Progress

## 2026-03-26

- 启动 `contest-awd-http-handler-split-phase2`，目标是继续拆 `contest` AWD HTTP 主流程文件。
- 盘点确认 `api/http/awd_handler.go` 同时承载四类职责：
  - round 创建与 round 查询入口
  - service check 写入与服务状态查询入口
  - attack log / submit attack 入口
  - round summary 查询入口
- 已完成文件拆分：
  - `awd_handler.go` 收缩为 AWD handler 类型、接口与构造函数
  - `awd_round_handler.go` 承载 round 相关 HTTP 入口
  - `awd_service_handler.go` 承载 service check 相关 HTTP 入口
  - `awd_attack_handler.go` 承载 attack 相关 HTTP 入口
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestCompositionModulesExposeContracts' -count=1`
