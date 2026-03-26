# Task Plan

## Goal

完成 `contest` Phase 2 后续收口：消除宽仓储接口、继续收紧应用层查询/查找依赖，并把 composition 装配收口到按子能力拆分的 builder。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点三处服务真实依赖面 | completed | 已确认各自只需要 contest 写侧、计分板读侧、状态调度侧的子集能力 |
| 2. 以测试桩最小化暴露 red case | completed | 测试桩已最小化，并新增 `contest_service / scoreboard_service / team_service / scoreboard_admin_service` 窄依赖验证 |
| 3. 在 `contest/ports` 中拆出窄接口 | completed | 已新增 `ContestCommandRepository / ContestLookupRepository / ContestListRepository / ContestScoreboardRepository / ContestScoreboardAdminRepository / ContestStatusRepository` |
| 4. 切换服务与相关测试 | completed | `contest / scoreboard / status_updater / challenge / participation / team / awd / submission` 已切到窄端口 |
| 5. composition 收口 | completed | `BuildContestModule` 已拆为 `core / awd / challenge / participation / team / submission` 局部 builder，且 deps 改为 ports 窄接口 |
| 6. focused 验证 | completed | `contest/...` 与 `internal/app` 关键装配测试已多轮通过 |

## Acceptance Checks

- `contest/application/commands/contest_service.go` 不再依赖宽 `ports.Repository`
- `contest/application/queries/scoreboard_service.go` 不再依赖宽 `ports.Repository`
- `contest/application/jobs/status_updater.go` 不再依赖宽 `ports.Repository`
- `contestports.Repository` 已删除，不再保留 legacy 宽仓储接口
- `contest` 相关应用服务已统一收敛到按用例划分的窄端口
- `BuildContestModule` 不再是单个大装配函数，`contestModuleDeps` 也不再持有 concrete contest repo 字段
- 对应测试桩不需要再实现无关方法
- focused tests 通过

## Result

- 未改外部 API 与路由
- `contest/infrastructure` 继续提供 concrete 实现，但 composition 与应用层已改为面向窄端口装配
- 该阶段目标已完成，可进入下一轮 `contest` 后续优化或结束当前迁移线
