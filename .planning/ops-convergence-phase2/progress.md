# Progress

## 2026-03-23

- 创建 `ops-convergence-phase2` 计划
- 完成 `notification -> ops`：
  - 通知 HTTP / websocket handler 已迁入 `ops/api/http`
  - 通知 service 已迁入 `ops/application`
  - 通知 repository 已迁入 `ops/infrastructure`
  - `composition.SystemModule.NotificationHandler` 已切到 `ops` contract
- 删除后端 `internal/module/system` 中剩余通知实现
- 通知路由与 websocket 路径保持不变
- 定向验证通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports' -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/auth -count=1`
