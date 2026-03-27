# Progress

## 2026-03-27

- 启动 `contest-submission-submit-split-phase4`，目标是继续拆 `contest` submission submit 文件。
- 盘点确认 `application/commands/submission_submit.go` 同时承载三类职责：
  - contest/challenge/team 与 rate-limit/flag 的前置校验
  - incorrect submission 落库与响应构造
  - correct submission 编排转发
- 已完成文件拆分：
  - `submission_submit_validation.go` 承载前置校验
  - `submission_incorrect_submit.go` 承载 incorrect submission 落库与响应
  - `submission_submit.go` 仅保留入口编排
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
