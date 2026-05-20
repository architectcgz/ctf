# Reuse Decision

## Change type
runtime contracts file normalization

## Existing code searched

- `code/backend/internal/module/runtime/contracts/runtime_model_compat.go`
- `code/backend/internal/module/runtime/contracts/*.go`
- `code/backend/internal/module/runtime/**/*.go`
- `code/backend/internal/module/instance/**/*.go`
- `code/backend/internal/module/practice/**/*.go`
- `code/backend/internal/module/challenge/**/*.go`
- `code/backend/internal/module/contest/**/*.go`

## Similar implementations found

- `code/backend/internal/module/runtime/contracts/persistence.go`
- `code/backend/internal/module/challenge/contracts/topology_spec.go`

## Decision
refactor_existing

## Reason

`runtime_model_compat.go` 已不再承载“兼容层”语义，而是 runtime 模块真实对外 contract 的事实源。继续把容器规格、拓扑策略、运行时明细、访问地址解析混放在同一个 `compat` 文件里，会误导 owner 判断，也会让后续 boundary 清理继续围绕错误命名打补丁。

这次不改变导出符号和运行时行为，只把现有 contract 按职责拆回 `runtime/contracts`：

- 容器规格：`container_spec.go`
- 拓扑策略：`topology_contract.go`
- 运行时明细编解码：`runtime_details.go`
- 访问地址解析：`access_url.go`

## Files to modify

- `code/backend/internal/module/runtime/contracts/runtime_model_compat.go`
- `code/backend/internal/module/runtime/contracts/container_spec.go`
- `code/backend/internal/module/runtime/contracts/topology_contract.go`
- `code/backend/internal/module/runtime/contracts/runtime_details.go`
- `code/backend/internal/module/runtime/contracts/access_url.go`

## After implementation

- runtime 模块仍作为这些 contract 的唯一 owner。
- 后续如果要进一步拆分 challenge/runtime 之间的拓扑协议 owner，再基于清晰文件边界继续做，而不是围绕 `compat` 命名继续累积歧义。
