# Reuse Decision

## Change type

service / mapper / contract

## Existing code searched

- `code/backend/internal/module/identity/application/commands/*.go`
- `code/backend/internal/module/identity/application/queries/*.go`
- `code/backend/internal/module/identity/contracts/*.go`
- `code/backend/internal/module/auth/application/commands/*.go`
- `code/backend/internal/module/auth/api/http/*.go`
- `code/backend/internal/model/user.go`
- `code/backend/internal/model/role.go`

## Similar implementations found

- `identity/contracts` 已经承担跨模块用户 contract，不需要把 `identity/entity` 直接暴露给其他模块
- `auth/application/commands` 已经存在本地 mapper wrapper，可以继续在本层吸收 mapper 输入而不是反向依赖 `identity/entity`
- `identity/application/commands` 与 `identity/application/queries` 已经分离 command/query mapper，适合继续沿用现有结构完成 owner 迁移

## Decision

refactor_existing

## Reason

这次不是新增一套用户/角色共享模型，而是把历史上落在 `internal/model` 的 owner 收回 `identity` 模块，同时保持模块外仍然通过 `identity/contracts` 交互。对应到 `auth` 一侧，也不新建并行 adapter，而是在现有 `application/commands` mapper 结构里补本地输入类型，避免把 `identity/entity` 倒灌回 `auth`。

## Files to modify

- `.harness/reuse-decisions/identity-user-role-owner-localization.md`
- `code/backend/internal/module/auth/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/auth/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/auth/application/commands/service_test.go`
- `code/backend/internal/module/identity/application/commands/admin_service_test.go`
- `code/backend/internal/module/identity/application/commands/profile_service_test.go`

## After implementation

- `auth` 不再为了 mapper 生成而直接 import `identity/entity`
- `identity` 的 user / role owner 继续留在模块内，模块外经 `identity/contracts` 访问
- 这份 decision 只作为当前任务的 reuse-first 证据，不额外回写长期 reuse 索引
