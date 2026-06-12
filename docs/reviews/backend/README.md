# 后端 Review 目录说明

`docs/reviews/backend/` 只保留入口说明和历史 backend review 归档。

当前约定：

- 正式 backend gate review 必须绑定明确 commit 或 commit range；只看 `current worktree diff` 的记录只能作为 draft review 或预检，不单独满足 gate。
- 同一任务多轮 backend review 必须拆成多个文件并保留每轮原始结论，命名为 `YYYY-MM-DD-backend-review-<topic>-round-<n>.md` 或更具体的 `YYYY-MM-DD-gate-review-<topic>-round-<n>.md`。
- 每个 round 文件必须写清 `Commit` / `Commit Range`、`Branch`、`Task / Plan`、`Reviewer mode` 和 `Diff basis`，方便后续从 review finding 反查对应代码。
- 初审发现 blocker 后，修复提交对应下一轮 review；不要把后续 pass verdict 追加到初审文件里覆盖原始 blocked 结论。
- 如需给当前任务保留总览，可以新增轻量 summary 文件，只列出每轮文件、commit、verdict、修复提交和最终状态。
- 历史 backend review 正文统一移动到 `archive/<YYYY-MM>/`。
- `archive/` 下的正文只作为时间点证据保留，不再作为当前事实源、当前待办或默认读取入口。
- 判断当前后端边界、风险和待修项时，优先回到当前代码、`docs/architecture/`、`docs/contracts/`、`docs/design/` 与当前 `docs/plan/impl-plan/`。
- 如果后续某轮 backend review 仍需要被当前任务直接引用，可以暂留活动目录；一旦只剩追溯价值，就归档，不再在旧正文里回写现状。
