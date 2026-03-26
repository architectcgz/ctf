# Progress

## 2026-03-26

- 启动 `contest-team-repository-split-phase2`，目标是继续拆 `contest` team repository 主流程文件。
- 盘点确认 `infrastructure/team_repository.go` 同时承载多类职责：
  - 创建 team 与成员绑定事务
  - 成员加入/移除/删除 team 事务
  - team / member / user / registration 查询
  - unique violation 判断 helper
- 已完成文件拆分：
  - `team_repository.go` 收缩为 TeamRepository 类型与构造函数
  - `team_membership_repository.go` 承载成员事务与 registration 绑定 helper
  - `team_query_repository.go` 承载 team/member/user 查询
  - `team_repository_support.go` 承载 unique violation helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/infrastructure -run 'TestTeamRepositoryCreateWithMemberSyncsContestRegistration' -count=1`
