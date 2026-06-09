# Runtime Allocation Repository Review

日期：2026-06-09

范围：`runtime/infrastructure.Repository` 中 allocation persistence 拆到 `runtime/infrastructure.AllocationRepository`，以及对应 composition、practice、runtime、contest 测试适配。

## 结论

无 blocker。

当前 diff 在 owner 划分上是自洽的：`AllocationRepository` 接走 port/subnet reservation、allocation release 和 restart host port sync；`runtime/infrastructure.Repository` 不再声明这些 allocation 方法。生产装配也已经切到新 owner，包括 runtime module provisioning / cleanup persistence、runtime node router allocation 依赖、practice runtime port owner、instance / contest mixed lifecycle release 路径。

`instance_runtime_lifecycle_tx` 现在只把 instance 状态写和 allocation release 放进同一个 DB transaction，没有把 AWD workspace / operation 这类 runtime state 再混入 lifecycle release，符合当前拆分目标。

文档同步准确：`backend-module-boundary-target.md` 和技术债 backlog 都已经说明 allocation persistence 已在 `runtimeinfra.AllocationRepository`，剩余 `runtimeinfra.Repository` 责任主要是 AWD workspace / AWD service operation persistence、runtime state index 和 migration-facing state lookup。

## Findings

- Low：测试适配里仍有几处手工复制 production 的“instance 状态写 + allocation release”事务编排。它们已经改成 `instanceRepo + allocationRepo`，没有把职责退回旧 runtime repo，因此不阻塞当前 slice；但后续如果 lifecycle 规则继续变化，需要同步维护这些 test helper。

## 复验

Reviewer 独立复跑并通过：

```bash
go test ./internal/app -run 'TestPracticeModuleWiresRuntimePortOwnerFromCompositionRoot|TestRuntimeRepositoryDoesNotOwnAllocationPersistence' -count=1
go test ./internal/app/composition -count=1
```

## 剩余风险

- `runtimeinfra.Repository` 里剩余的 AWD workspace / AWD service operation persistence、runtime state index 和 migration-facing state lookup 仍是下一批真实结构债。
- 如果后续继续改 lifecycle mixed path，建议补更聚焦的 composition transaction 原子性测试，而不是只依赖包级测试间接覆盖。
