# 2026-05-20 runtime network occupied subnet review

> 当前状态：已完成
> 范围：runtime 网络子网分配优化、topology 级 occupied set 共享、owner 历史预留迁移
> 验证证据：`go test ./internal/module/runtime/... -count=1`；`go test ./internal/app/composition/... -count=1`

## Review 结论

- 结论：`pass with minor issues`
- blocker：无
- 当前实现满足两项核心目标：
  - `CreateTopology` 在单次 topology 创建中只读取一次 Docker 当前已占用子网。
  - 同一 topology 内多个 network 共享同一份 occupied set，显式 subnet、冲突重试 subnet、owner 旧预留迁移后的 subnet 都会进入这份集合。
  - 只有 topology 中存在需要动态分配 subnet 的 network 时，才会读取 Docker 当前已占用子网；`shared-only` 和 `explicit-subnet-only` 场景会跳过这一步。

## 已确认项

- `code/backend/internal/module/runtime/application/commands/provisioning_service.go`
  - `ListNetworkSubnets` 已提升到 network 循环外部，避免每个 network 重复读取 Docker 子网。
  - 子网一旦被当前 topology 采用，就会立刻并入共享 occupied set，后续 network 直接复用这份排除集。
  - 当前只会在存在 `!Shared && Subnet == ""` 的 network 时才预读 Docker 子网；单容器独立题目仍会命中该分支，因为它默认创建非 shared、无显式 subnet 的默认 network。
- `code/backend/internal/module/runtime/infrastructure/repository.go`
  - 当同一 `owner/networkKey` 已有历史预留，且该 subnet 同时落在 `excludedSubnets` 中时，仓储层不再返回旧 subnet。
  - 当前实现会尝试把 owner 记录迁移到新的可用 subnet，和 `subnet` 主键、`(instance_id, network_key)` 唯一键约束保持一致。
- `code/backend/internal/module/runtime/service_test.go`
  - 已覆盖 topology 级共享 occupied set。
  - 已覆盖显式 subnet 对后续动态 subnet 的影响。
  - 已覆盖 owner 旧预留在 runtime occupied 情况下的无重试迁移路径。
  - 已覆盖 `shared-only` 和 `explicit-subnet-only` 场景下跳过 `ListNetworkSubnets` 的懒加载行为。
- `code/backend/internal/module/runtime/infrastructure/repository_destroyed_at_test.go`
  - 已覆盖 owner 历史预留迁移到新 subnet 的仓储层行为。

## Residual Risk

- 目前没有单独具象化“候选 subnet 插入失败后，owner 记录迁移又因并发冲突失败，再继续下一个候选 subnet”的更窄竞争路径测试。
- `shared=true` 且由 Docker 临时分配 subnet 的组合场景，当前仍更偏性能残余风险，而不是 correctness 缺陷；如后续出现相关拓扑，可再补一条组合回归测试。
