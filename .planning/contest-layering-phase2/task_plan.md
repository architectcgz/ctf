# Task Plan

## Goal

启动 `contest` Phase 2 后续收口，先把 `contest_service / scoreboard_service / status_updater` 从宽 `ports.Repository` 迁到按用例拆分的窄端口。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点三处服务真实依赖面 | completed | 已确认各自只需要 contest 写侧、计分板读侧、状态调度侧的子集能力 |
| 2. 以测试桩最小化暴露 red case | in_progress | 先让测试桩只实现必要方法，暴露当前宽接口问题 |
| 3. 在 `contest/ports` 中拆出窄接口 | pending | 保持 `infrastructure.Repository` 继续实现这些接口，不复制实现 |
| 4. 切换三处服务与相关测试 | pending | `commands/contest_service`、`queries/scoreboard_service`、`jobs/status_updater` |
| 5. focused 验证 | pending | `contest/...` + `internal/app` 关键装配测试 |

## Acceptance Checks

- `contest/application/commands/contest_service.go` 不再依赖宽 `ports.Repository`
- `contest/application/queries/scoreboard_service.go` 不再依赖宽 `ports.Repository`
- `contest/application/jobs/status_updater.go` 不再依赖宽 `ports.Repository`
- 对应测试桩不需要再实现无关方法
- focused tests 通过

## Constraints

- 只做第一刀端口收紧，不顺手改造其他 `contest` service
- 不改外部 API 与路由
- `contest/infrastructure.Repository` 继续作为 concrete 实现，先不拆仓储文件
