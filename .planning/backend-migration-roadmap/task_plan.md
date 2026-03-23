# Task Plan

## Goal

把 CTF 后端剩余架构迁移拆成可独立推进的 5 个后续切片，按依赖顺序逐个完成。

## Tracks

| Track | Status | Notes |
|---|---|---|
| 1. `runtimeinfra` 并回 `runtime` | completed | 过渡模块已删除，职责已并回 `runtime/infrastructure` |
| 2. 两个 readmodel 根壳清理 | completed | `practice_readmodel` / `teaching_readmodel` 已删除根壳并统一 contract 边界 |
| 3. `auth + adminuser -> identity` | in_progress | Phase 1 已完成：用户主数据已收拢到 `identity`，后续还需继续消化残余 auth 流程实现 |
| 4. `system -> ops` | pending | 让运营能力从杂项模块变成明确 owner |
| 5. 大业务模块内部物理分层 Phase 1 | pending | 先处理 `challenge / contest / assessment / practice` 的 concrete 暴露 |

## Recommended Order

1. 先做 `runtimeinfra-merge-phase1`
2. 再做 `readmodel-root-cleanup`
3. 再做 `identity-convergence-phase1`
4. 再做 `ops-convergence-phase1`
5. 最后做 `domain-layering-phase1`

## Constraints

- 不保留兼容层
- 外部 API 路径默认保持稳定
- 测试按最小充分验证执行，默认限核
- 不碰当前主仓库的 frontend 脏改动
