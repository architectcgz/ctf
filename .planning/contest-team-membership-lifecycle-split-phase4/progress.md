# Progress

## 2026-03-27

- 启动 `contest-team-membership-lifecycle-split-phase4`，目标是继续拆 `contest` team membership repository 文件。
- 盘点确认 `infrastructure/team_membership_repository.go` 同时承载两类职责：
  - team 生命周期事务（CreateWithMember / DeleteWithMembers）
  - team 成员事务（AddMemberWithLock / RemoveMember）
- 已完成文件拆分：
  - `team_membership_lifecycle_repository.go` 承载 team 生命周期事务
  - `team_membership_repository.go` 保留成员加入/离队事务
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
