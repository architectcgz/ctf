# Progress

## 2026-03-26

- 创建 `contest-layering-phase2` 计划
- 确认首刀范围为 `contest_service / scoreboard_service / status_updater`
- 目标是先收紧宽 `ports.Repository`，而不是继续拆 composition
- 使用测试桩最小化暴露 red case：
  - `contest_service` 测试桩只保留 `Create / FindByID / Update`
  - `scoreboard_service` 新增最小仓储桩测试
  - `status_updater` 测试桩只保留 `ListByStatusesAndTimeRange / UpdateStatus`
- 新增 `ContestCommandRepository / ContestScoreboardRepository / ContestStatusRepository`
- 三处服务构造器已切到窄端口，`infrastructure.Repository` 继续复用原实现
