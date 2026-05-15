# Reuse Decision

## Change type
service / handler / api / hook

## Existing code searched

- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/app/router_routes.go`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdServiceOperations.ts`
- `code/frontend/src/components/platform/contest/AWDInstanceOrchestrationPanel.vue`

## Similar implementations found

- `StartAdminContestAWDTeamService` 已经负责单个队伍服务实例启动
- `GetContestAWDInstanceOrchestration` 已经提供队伍 × 服务矩阵读取面
- `useAwdServiceOperations.ts` 已经承接单格 / 本队 / 全量启动动作，但当前本队 / 全量是前端循环单格接口
- `router_routes.go` 与 `handler.go` 已经有 `GET/POST /admin/contests/:id/awd/instances` 资源族

## Decision
extend_existing

## Reason

本次不新建并行模块。后端直接在现有 `practice` 命令服务和 `awd/instances` 资源族上扩一个专用 `prewarm` 批量入口，前端继续复用现有 orchestration 面板与操作 hook，只把报名阶段的“本队 / 全量”改为调用新接口。这样可以复用现有单格启动 owner、实例复用语义和编排视图，避免复制第二套 AWD 实例控制链路。

## Files to modify

- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_scope.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/admin/contests.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdServiceOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts`

## After implementation

- 如果这条“批量预热复用单格启动 owner、前端只切换调用方式”的模式会继续复用，再补充到 `harness/reuse/history.md` 或 `harness/reuse/index.yaml`
