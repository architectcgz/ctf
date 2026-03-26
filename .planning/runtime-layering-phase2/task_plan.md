# Task Plan

## Goal

推进 `runtime` Phase 2 分层首刀：接通已存在的 `application/commands` 与 `application/queries`，让 composition、HTTP handler 与关键测试依赖从旧 `runtime/application` 根包切到分层后的读写服务。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 runtime 现有 Phase 2 半成品 | completed | 已确认 `application/commands` 与 `application/queries` 已存在，但主装配和调用点仍主要依赖 root `application` |
| 2. 补分层守卫 red case | completed | 已为 composition 与 `runtime/api/http` 增加防回退约束，避免继续依赖旧 root `application` 读写 facade |
| 3. 切换 runtime 主调用面 | completed | `composition`、`runtime/api/http`、`testutil`、focused tests 已切到 `commands / queries` |
| 4. 清理旧 root facade | completed | 已删除不再使用的 root `instance/query/proxy ticket` facade，保留共享 contract / topology / cleanup / provisioning 能力 |
| 5. focused 验证 | completed | `internal/app`、`internal/module/runtime/...` 与受影响模块定向测试已通过 |

## Acceptance Checks

- `runtime_module.go` 不再通过 root `runtime/application` 构造 `InstanceService / QueryService / ProxyTicketService`
- `runtime/api/http/handler.go` 不再依赖 root `runtime/application` 的 `ProxyTicketClaims`
- `runtime/testutil` 与受影响 focused tests 改为显式组合 `commands / queries`
- `runtime/application` 根包不再保留 `instance_service.go`、`query_service.go`、`proxy_ticket_service.go`
- `runtime/architecture_test.go` 与 `internal/app/router_test.go` 已补防回退守卫
- focused tests 通过

## Result

- 不改外部 API、路由和跨模块公开 contract
- `runtime` 继续作为运行时 owner，不拆新模块；本刀只做模块内部 Phase 2 分层收口
- `runtime` 的实例读写与 proxy ticket 主调用面已切到 `application/commands` 与 `application/queries`
