# 后端 Review 目录说明

`docs/reviews/backend/` 只保留入口说明和历史 backend review 归档。

当前约定：

- 历史 backend review 正文统一移动到 `archive/<YYYY-MM>/`。
- `archive/` 下的正文只作为时间点证据保留，不再作为当前事实源、当前待办或默认读取入口。
- 判断当前后端边界、风险和待修项时，优先回到当前代码、`docs/architecture/`、`docs/contracts/`、`docs/design/` 与当前 `docs/plan/impl-plan/`。
- 如果后续某轮 backend review 仍需要被当前任务直接引用，可以暂留活动目录；一旦只剩追溯价值，就归档，不再在旧正文里回写现状。
