# Progress

## 2026-03-27

- 启动 `contest-awd-domain-split-phase4`，目标是继续拆 `contest` AWD domain helper 文件。
- 盘点确认 `domain/awd.go` 同时承载三类职责：
  - AWD DTO 映射与 summary 排序
  - check source / check result 规范化
  - flag 生成与 unique error support
- 已完成文件拆分：
  - `awd_response.go` 承载 AWD DTO 映射与 summary 排序
  - `awd_check_support.go` 承载 check source / check result support
  - `awd_flag_support.go` 承载 flag 生成与 unique error support
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
