# Progress

## 2026-03-26

- 启动 `contest-submission-flow-split-phase2`，目标是继续拆 `contest` submission 写侧文件。
- 盘点确认 `application/commands/submission_submit.go` 同时承载三类职责：
  - `SubmitFlagInContest` 提交流程入口
  - `resolveTeamID` team 解析与报名校验
  - `handleCorrectSubmission` 正确提交后的事务计分与排行榜同步
- 已完成文件拆分：
  - `submission_submit.go` 保留提交入口流程
  - `submission_validation.go` 承载 team 解析与报名校验
  - `submission_scoring.go` 承载正确提交后的事务计分与排行榜同步
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
