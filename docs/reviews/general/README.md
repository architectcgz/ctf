# 通用 Review 目录说明

`docs/reviews/general/` 只保留仍可能被当前任务直接引用的通用 review 记录。

当前约定：

- 已被后续代码、架构事实源或专项 review 吸收的历史 general review，移动到 `archive/<YYYY-MM>/`。
- `archive/` 下的正文只作为时间点证据保留，不再作为当前事实源、当前待办或默认读取入口。
- 判断“现在是否还要修”时，优先回到当前代码、`docs/architecture/`、`docs/contracts/`、`docs/operations/` 和仍处于活动目录的最新 review。
- 如果某份 general review 仍需要长期追溯但已不适合作为活动入口，优先归档，不再继续在正文里回写现状修订。
