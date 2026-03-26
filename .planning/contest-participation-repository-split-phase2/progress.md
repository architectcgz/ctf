# Progress

## 2026-03-26

- 启动 `contest-participation-repository-split-phase2`，目标是继续拆 `contest` participation repository 主流程文件。
- 盘点确认 `infrastructure/participation_repository.go` 同时承载三类职责：
  - registration 查询与写入
  - announcement 查询与写入
  - solved progress 查询
- 已完成文件拆分：
  - `participation_repository.go` 收缩为 ParticipationRepository 类型、构造函数与 `dbWithContext`
  - `participation_registration_repository.go` 承载 registration 查询与写入
  - `participation_announcement_repository.go` 承载 announcement 查询与写入
  - `participation_progress_repository.go` 承载 solved progress 查询
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
