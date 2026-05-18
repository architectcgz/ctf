# contest announcement model 模块内化实现方案

## Objective

把 `internal/model/contest_announcement.go` 中的 `ContestAnnouncement` 持久化实体收回 `internal/module/contest/entity`，让 contest 公告读写链路和 app 测试直接依赖模块内实体。

## Non-goals

- 不处理 `Contest`
- 不处理 `ContestRegistration`
- 不改变公告接口、排序语义或表结构

## Inputs

- `internal/model/contest_announcement.go`
- `internal/module/contest/...`
- `internal/app/full_router_integration_test.go`
- `.harness/reuse-decisions/contest-announcement-model-localization.md`

## Ownership Evaluation

- owner 明确：`ContestAnnouncement` 只由 `contest` 模块消费
- landing zone 明确：`internal/module/contest/entity/announcement.go`
- 非目标明确：不扩到 `Contest` / `Submission`
- 结构收敛目标明确：删除旧全局模型文件

## Task slices

1. 新增 `contest/entity/announcement.go`
2. 更新 contest ports / repository / service / mapper / tests
3. 更新 app 集成测试 schema 与 seed 引用
4. 删除 `internal/model/contest_announcement.go`

## Validation

- `go generate ./internal/module/contest/application/commands`
- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_AdminContestAnnouncementLifecycle' -count=1`
- `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- `bash scripts/check-consistency.sh`

## Review focus

- announcement 实体是否完全收敛到 `contest/entity`
- repository / service / mapper 是否没有残留全局依赖
- app / testsupport schema 与 seed 是否同步切换

## Rollback

本刀无 schema 变更，如有回归可直接恢复到 `internal/model/contest_announcement.go`。
