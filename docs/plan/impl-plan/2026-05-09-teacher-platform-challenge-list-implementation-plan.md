# 教师可见平台题库列表 Implementation Plan

## Objective

让教师访问 `/platform/challenges` 时能够看到平台题库中的 Jeopardy 题目列表，而不是因为后端默认附加 `created_by = 当前教师` 过滤而返回空列表。

## Non-goals

- 不放开教师对他人题目的详情读取
- 不放开教师对他人题目的更新、删除、发布检查、题解管理或拓扑编辑
- 不调整管理员在平台题库页的现有行为
- 不修改前端页面结构与操作入口

## Inputs

- `code/backend/internal/module/challenge/api/http/handler.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/full_router_integration_test.go`

## Change Surface

- 后端 authoring challenge 列表接口的教师查询约束
- 对应的 full-router 集成测试预期

## Decision Notes

- 当前教师访问 `/api/v1/authoring/challenges` 时，`Handler.ListChallenges` 会强制把 `query.CreatedBy` 设为当前教师 ID。
- 详情与写操作仍由 `challengeOwnerGuard` 约束，因此只放开列表不会自动放开 `GET /authoring/challenges/:id`、`PUT`、`DELETE`、`publish-requests` 等 owner 受限路由。
- 本轮推荐最小改动：移除列表接口中的教师强制 owner 过滤，保留 ownerGuard 不动。

## Task Slices

### Slice 1: 放开教师列表查询

- 修改 `Handler.ListChallenges`
- 教师访问 `/api/v1/authoring/challenges` 时不再自动注入 `CreatedBy`

Files / modules:

- `code/backend/internal/module/challenge/api/http/handler.go`

Validation:

- 相关 Go 测试可编译通过

Review focus:

- 是否只影响列表接口，不影响 ownerGuard 路由
- 是否保持 admin 查询行为不变

### Slice 2: 更新权限回归测试

- 调整 full-router 集成测试
- 断言教师能在列表中看到 admin 创建的题
- 同时保留教师对他人题目的详情/修改等操作仍返回 `403`

Files / modules:

- `code/backend/internal/app/full_router_integration_test.go`

Validation:

- 运行对应 full-router 集成测试

Review focus:

- 行为边界是否和目标一致
- 是否避免把教师权限扩大到详情或写操作

## Verification Plan

1. `go test ./internal/module/challenge/...`
2. `go test ./internal/app -run TestFullRouter_TeacherCanBrowseAllChallengesButOnlyManageOwnChallenges`

## Rollback

- 恢复 `Handler.ListChallenges` 中对 teacher 的 `CreatedBy` 强制注入
- 恢复 full-router 测试对教师列表只见本人题目的旧断言
