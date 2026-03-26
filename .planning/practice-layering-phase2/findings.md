# Findings

- `practice` 已完成模块物理分层，但 `ports/ports.go` 仍保留宽 `PracticeRepository`，命令服务、分数写侧、排行榜读侧共用一个仓储接口。
- `commands/service.go` 实际只需要 contest 查找、事务内实例创建、提交流程相关方法；`commands/score_service.go` 与 `queries/score_service.go` 各自只需要更小的分数/排行依赖面。
- [`practice_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/practice_module.go) 仍直接把 `practiceinfra.Repository` concrete 实现塞给多个服务，缺少与 `contest/challenge phase2` 对齐的 typed deps 收口。
- `runtime` 适配桥仍相对独立，本次优先不碰 `practiceRuntimeInstanceServiceAdapter`，先完成 `PracticeRepository` 缩口。
