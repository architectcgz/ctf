# Progress

## 2026-03-27

- 启动 `contest-awd-domain-check-support-split-phase4`，目标是继续拆 `contest` AWD domain check support 文件。
- 盘点确认 `domain/awd_check_support.go` 同时承载两类职责：
  - check/attack source 归一化
  - check result JSON parse/marshal 与手工检查补全
- 已完成文件拆分：
  - `awd_source_support.go` 承载 check/attack source 归一化 helper
  - `awd_check_result_support.go` 承载 check result parse/marshal 与手工检查补全 helper
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
