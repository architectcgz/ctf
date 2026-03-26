# Task Plan

## Goal

完成 `challenge` Phase 2 收口：删除 legacy 宽仓储接口，把应用层依赖切到按用例划分的窄端口，并让 composition 装配收口到 typed deps 与局部 builder。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 challenge 应用层真实依赖面 | completed | 已确认 command/query/flag/writeup/topology/image-usage 共享宽仓储是主要问题 |
| 2. 先补 red case 守卫测试 | completed | 已新增 ports 宽接口禁用测试与 composition typed deps / sub-builder 守卫 |
| 3. 在 `challenge/ports` 中拆出窄接口 | completed | 已新增 `ChallengeCommandRepository / ChallengeQueryRepository / ChallengeFlagRepository / ChallengeImageUsageRepository / ChallengeWriteupRepository / ChallengeTopologyRepository` |
| 4. 切换 challenge application 构造依赖 | completed | `commands / queries` 相关服务已全部切到窄端口 |
| 5. composition 收口 | completed | `challengeModuleDeps` 已切到 ports/contracts，`BuildChallengeModule` 已拆为 image/core/flag/topology/writeup 局部 builder |
| 6. focused 验证 | completed | `challenge/...` 与 `internal/app` 相关装配测试已通过 |

## Acceptance Checks

- `challenge/ports/ports.go` 不再保留 legacy 宽 `ChallengeRepository`
- `challenge/application/commands/*.go` 不再依赖单个宽主仓储接口
- `challenge/application/queries/*.go` 不再依赖单个宽主仓储接口
- `BuildChallengeModule` 不再把全部装配堆在单个函数中
- `challengeModuleDeps` 不再持有 concrete `challengeinfra` repo 字段
- `challenge/architecture_test.go` 与 `internal/app/router_test.go` 已补防回退守卫
- focused tests 通过

## Result

- 未改外部 API、handler 路由与跨模块 contract
- `challenge/infrastructure.Repository` 继续复用原 concrete 实现，但应用层与 composition 已改为面向窄端口装配
- `challenge` Phase 2 已完成，可转入下一条迁移线
