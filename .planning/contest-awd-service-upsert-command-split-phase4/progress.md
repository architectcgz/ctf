# Progress

## 2026-03-27

- 启动 `contest-awd-service-upsert-command-split-phase4`，目标是继续拆 `contest` AWD service upsert command 文件。
- 盘点确认 `application/commands/awd_service_upsert_commands.go` 同时承载三类职责：
  - UpsertServiceCheck 参数编排与输入规范化
  - 事务内 service check upsert 与 team score 重算
  - 事务后缓存重建、服务状态同步与响应映射
- 已完成文件拆分：
  - `awd_service_upsert_transaction.go` 承载事务写入与重算
  - `awd_service_upsert_response_support.go` 承载后置同步与响应映射
  - `awd_service_upsert_commands.go` 保留入口编排
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
