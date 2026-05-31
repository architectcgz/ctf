# 2026-05-31 composables owner cleanup batch7 probe easter eggs plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch7-probe-easter-eggs.md`

## 目标

把 `useProbeEasterEggs` 从历史 `code/frontend/src/composables/` 收口到 `shared/model/common/`，并同步修正消费方、测试与架构文档。

## 非目标

- 不改 probe key、threshold、存储 key 或返回值结构
- 不改 `sessionStorage` 优先、内存降级的运行语义
- 不处理 `useWebSocket`

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useProbeEasterEggs.ts`
- `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`

## 目标归属

- `useProbeEasterEggs` -> `shared/model/common/useProbeEasterEggs.ts`

理由：

- 该能力跨 feature 复用、无业务 owner，但持有 `sessionStorage` 与内存回退状态，属于共享 common model

## 任务切片

### Slice 1

迁移 owner：

- 新建 `shared/model/common/useProbeEasterEggs.ts`
- 删除 `composables/useProbeEasterEggs.ts`
- 修正运行时代码 import

验证：

- `cd code/frontend && timeout 180s npm run typecheck`

### Slice 2

修正测试与文档：

- 修正 `useProbeEasterEggs` 测试路径
- 修正 `useLoginPage.test.ts` mock
- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/composables/__tests__/useProbeEasterEggs.test.ts src/features/auth/model/useLoginPage.test.ts`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- `useProbeEasterEggs` 被 feature model 和 shared ui 同时消费，路径断言和 mock 容易漏
- 文档里当前明确写了 `useProbeEasterEggs.ts` 还留在 `composables/`，需要同步收口

## Review focus

- 是否只发生 owner 迁移，没有混入行为变化
- `useProbeEasterEggs` 是否落在 `shared/model/common` 而不是 `shared/lib`
- 是否清干净了旧 `src/composables/useProbeEasterEggs` 路径
