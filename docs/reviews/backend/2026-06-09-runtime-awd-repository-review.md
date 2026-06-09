# Runtime AWD Repository Review

日期：2026-06-09

范围：`runtime/infrastructure.Repository` 中 AWD defense workspace / AWD service operation persistence 拆到 `runtime/infrastructure.AWDRepository`，以及 instance / contest composition 与测试适配。

## 结论

无 blocker。Review verdict：pass with minor issues。

本轮拆分方向和落点正确：AWD workspace / operation persistence 已从宽 `runtimeinfra.Repository` 移到 `runtimeinfra.AWDRepository`；生产 wiring、contest / instance 组合层、runtime / practice / contest 测试 adapter、事务例外 baseline 迁移和文档表述整体一致。

## Findings

- Low：`TestRuntimeRepositoryDoesNotOwnAWDPersistence` 初版没有覆盖 `BumpAWDDefenseWorkspaceRevision`、`FindRunningAWDDefenseWorkspaceByInstanceID`、`FinishAWDServiceOperation`。
  - 处理：已把这 3 个方法加入 `expected` / `blocked` owner guard。
- Low：composition adapter 在 `awdRepo == nil` 时会静默 no-op，未来新增 composition root 时可能漏接 AWD persistence。
  - 处理：已增加 `TestRuntimeAWDPersistenceWiredFromCompositionRoot`，明确要求 `BuildInstanceModule` 和 `buildContestEndedRuntimeCleaner` 从 composition root 注入 `runtimeinfra.NewAWDRepository(root.DB())`。

## 复验

Reviewer 独立复跑并通过：

```bash
go test ./internal/app -run 'TestRuntimeRepositoryDoesNotOwnAWDPersistence|TestRuntimeRepositoryDoesNotOwnAllocationPersistence' -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/runtime/infrastructure -count=1
```

本地修正后复跑并通过：

```bash
go test ./internal/app -run 'TestRuntimeRepositoryDoesNotOwnAWDPersistence|TestRuntimeAWDPersistenceWiredFromCompositionRoot|TestRuntimeRepositoryDoesNotOwnAllocationPersistence' -count=1
go test ./internal/module/runtime/infrastructure -count=1
go test ./internal/app/composition -count=1
go test ./internal/module/practice/... -count=1
go test ./internal/module/runtime/... -count=1
go test ./internal/module/contest/infrastructure -count=1
go test ./internal/module -count=1
python3 scripts/check-docs-consistency.py
bash scripts/check-code-changes.sh
git diff --check
```

## 剩余风险

- `runtimeinfra.Repository` 仍保留 active container inventory、container-to-node lookup、ACL migration state、runtime managed instance lookup 和 proxy traffic recorder support。它们没有在本轮误迁移，后续应作为 state/index slice 继续拆分。
