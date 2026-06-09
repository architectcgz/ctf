# Runtime State Repository Review

日期：2026-06-09

范围：删除宽 `runtime/infrastructure.Repository`，把 runtime managed instance lookup、active container inventory、container-to-node state lookup、ACL migration state 持久化拆到 `runtime/infrastructure.RuntimeStateRepository`，并让 proxy traffic recorder 拥有独立 concrete recorder。

## 结论

无 blocker。Review verdict：pass。

这次拆分后，runtime infrastructure 的 concrete owner 已经按 allocation / AWD / state / proxy traffic recorder 分开，composition 也已经同步切到 `RuntimeStateRepository`，没有再把“空壳宽仓储”作为默认落点。

## Findings

- Low：初版如果只拆 state repository、不同时处理 `proxy_traffic_recorder.go`，删除 `repository.go` 后会留下一个仅用于承载 `dbWithContext` 的隐式耦合。
  - 处理：本轮已把 proxy traffic recorder 改成独立 `runtimeinfra.ProxyTrafficEventRecorder`，不再依赖被删除的宽仓储。
- Low：state owner guard 不能只搜方法名 `FindByID`，否则会误伤 `RuntimeNodeRepository.FindByID`。
  - 处理：已把 guard 收窄到 `func (r *RuntimeStateRepository) ...` 级别，并补 `TestRuntimeStatePersistenceWiredFromCompositionRoot` 约束 production wiring。

## 复验

本轮本地复跑通过：

```bash
go test ./internal/app -run 'TestRuntimeRepositoryDoesNotOwnStatePersistence|TestRuntimeStatePersistenceWiredFromCompositionRoot|TestRuntimeRepositoryDoesNotOwnAWDPersistence|TestRuntimeRepositoryDoesNotOwnAllocationPersistence|TestRuntimeCompositionInjectsRuntimePersistenceIntoRuntimeModule|TestBuildContainerRuntimeModuleDelegatesToSubBuilders|TestBuildInstanceModuleDelegatesToSubBuilders' -count=1
go test ./internal/app/composition -run 'TestRuntimeNodeExecutionRouter.*' -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module/instance/... -count=1
go test ./internal/module -count=1
python3 scripts/check-docs-consistency.py
bash scripts/check-code-changes.sh
git diff --check
```

补充说明：

- `go test ./internal/app -count=1` 仍失败于既有断言 `TestUserSelfRoutesAreExtractedIntoDedicatedRegistrarFile`，报错为 `router_user_self_routes.go should contain "contestGroup := apiV1.Group(\"/contests\")"`；本轮未触达相关 router 文件，故未在此切片中一并处理。

## 剩余风险

- `RuntimeStateRepository` 现在是明确 owner，但内部仍同时承载 inventory、index lookup、ACL migration-facing state update 三类读取/更新。如果后续要继续细拆，建议按使用视角拆成更窄 view，而不是重新引入广义仓储。
- 更大的结构问题已经收敛到 capability interface / host adapter / `ContainerRuntimeModule` 的物理 owner 选择，这不再是 runtime persistence owner 的问题。
