# Progress

## 2026-03-27

- 启动 `contest-awd-validation-support-split-phase4`，目标是继续拆 `contest` AWD validation support 文件。
- 盘点确认 `application/commands/awd_validation_support.go` 同时承载三类职责：
  - contest / round 校验
  - challenge / team 资源校验与加载
  - user team 归属解析
- 已完成文件拆分：
  - `awd_resource_validation_support.go` 承载 challenge / team 资源校验与加载
  - `awd_team_validation_support.go` 承载 user team 归属解析
  - `awd_validation_support.go` 保留 contest / round 校验
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
