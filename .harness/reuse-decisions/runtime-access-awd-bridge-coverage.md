# Reuse Decision

## Change type

query / mapper / repository / test

## Existing code searched

- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `.harness/reuse-decisions/docker-runtime-access-host.md`
- `.harness/reuse-decisions/awd-control-plane-gap.md`
- `.harness/reuse-decisions/awd-runtime-hardening.md`

## Similar implementations found

- `docker-runtime-access-host`
  - 已经把 runtime 内部访问地址和学生侧展示地址拆成两个 owner，相关 handler / query 测试应继续挂在这条链路下
- `awd-control-plane-gap`
  - 已经把 AWD team/service scope control 与 orchestration 查询归到现有 repository / runtime query owner
- `awd-runtime-hardening`
  - 已经把 AWD effective schedule、恢复链路和运行时状态收口到既有 contest / practice / runtime owner

## Decision

extend_existing

## Reason

这几个文件本质上不是独立新能力：

- `response_mappers.go` 只是把 contest 对外响应切到现有 effective schedule owner
- `awd_service_instance_repository.go` 只是给已有 orchestration 查询补运行态字段
- `handler_test.go`、`handler_proxy_traffic_test.go`、`instance_service_test.go` 只是覆盖 runtime access host 重写和 AWD control gate 的既有出口

因此不新建并行 owner，而是继续复用 runtime access host、AWD control plane 和 AWD runtime hardening 这三条已存在的实现边界。

## Files to modify

- `code/backend/internal/module/contest/application/commands/response_mappers.go`
- `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`
- `code/backend/internal/module/runtime/api/http/handler_test.go`
- `code/backend/internal/module/runtime/api/http/handler_proxy_traffic_test.go`
- `code/backend/internal/module/runtime/application/instance_service_test.go`

## After implementation

- 后续如果 orchestration 查询继续扩字段，优先补到现有 contest/runtime query owner，不再新增平行 read model
- runtime access host 或 AWD control gate 的回归测试继续沿用现有出口测试，不分裂到新的测试入口
