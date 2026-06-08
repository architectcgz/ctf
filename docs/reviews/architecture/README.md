# 架构 Review 目录说明

`docs/reviews/architecture/` 只保留仍被当前材料直接引用的架构 review。

当前约定：

- 仍作为当前依据保留在活动目录的 review 目前有：
  - `2026-05-14-teaching-review-thesis-gap-review.md`
  - `2026-05-24-frontend-architecture-review.md`
  - `2026-06-01-ctf-frontend-architecture-review.md`
  - `2026-06-01-frontend-sliced-plan-review.md`
- 已被后续事实源吸收、只剩追溯价值，或只被历史 plan 引用的 review，移动到 `archive/<YYYY-MM>/`。
- `archive/` 下的正文只作为时间点证据保留，不再作为当前设计事实源或默认读取入口。
- 判断“现在是否还要修”时，优先回到当前代码、`docs/architecture/`、`docs/contracts/`、当前 `docs/plan/impl-plan/` 和仍留在活动目录的 review。
