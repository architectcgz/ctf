# AWD 与竞赛目录分页实现计划

## Plan Summary
- Objective
  - 为 `/academy/awd-reviews` 补齐可用的分页目录，并把同一 owner 下的 `/platform/awd-reviews` 一起切到真实可分页、可筛选的接口契约。
  - 统一默认分页大小为 `20`，不再沿用目录页里固定 `100` 的宽松拉取方式。
  - 为 `/contests`、`/scoreboard`、`/scoreboard/:contestId`、`/platform/contest-ops/contests` 改成真实服务端分页，去掉前端固定 `page_size=100` 再本地切页的做法。
  - 先补 contest 列表后端过滤/排序契约，再让学生端与平台端列表页切到同一套真实分页能力。
- Non-goals
  - 不改 AWD 复盘详情页的 round/team 细分查询。
  - 不调整 AWD 复盘导出链路。
  - 不在本次顺手扩散到题目列表、通知列表等已经分页完成的页面。
- Source architecture or design docs
  - `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`
- Dependency order
  - 先完成 AWD 目录分页闭环。
  - 再补 contest 列表后端过滤/排序/分页契约。
  - 最后切四个分页缺口页面并补测试与 review。
- Expected specialist skills
  - `development-pipeline`
  - `frontend-engineer`
  - `backend-engineer`

## Task 1
- Goal
  - 为教师 AWD 复盘目录定义查询输入与分页响应，接通 `status`、`keyword`、`page`、`page_size`。
- Touched modules or boundaries
  - `code/backend/internal/dto`
  - `code/backend/internal/module/assessment/api/http`
  - `code/backend/internal/module/assessment/application/queries`
  - `code/backend/internal/module/assessment/infrastructure`
  - `code/backend/internal/module/assessment/runtime`
- Dependencies
  - 依赖现有 AWD 复盘目录架构，不引入新模块。
- Validation
  - `go test` 跑 AWD 复盘 query / handler / router 相关测试。
- Review focus
  - 默认分页大小是否稳定落到 `20`
  - `status/keyword` 是否真实进入 repository 查询
  - `count + page slice` 是否与排序一致
- Risk notes
  - 目录查询当前使用原生 SQL，分页和筛选拼接时要避免排序、总数、返回列表不一致。

## Task 2
- Goal
  - 把 `useTeacherAwdReviewIndex` 切到共享分页 owner，并让 teacher/platform 两个目录页都显示分页控件。
- Touched modules or boundaries
  - `code/frontend/src/api/teacher/awd-reviews.ts`
  - `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewIndex.ts`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/widgets/teacher-awd-review/*`
  - `code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue`
- Dependencies
  - 依赖 Task 1 的分页接口契约。
- Validation
  - `vitest` 跑 AWD 复盘 API / teacher view / platform view / directory widget 相关测试。
- Review focus
  - 筛选变化时是否重置到第一页
  - 异步切页与防抖筛选是否有 stale response 问题
  - teacher/platform 是否保持共享 owner，不分叉出两套分页逻辑
- Risk notes
  - 现有前端测试 mock 仍按数组返回，需要同步切到分页结构。

## Task 3
- Goal
  - 为 contest 列表查询补齐 `mode`、多状态 `statuses`、排序字段与排序方向，使学生 `/contests` 与平台 `/admin/contests` 都能做真实服务端筛选与分页。
- Touched modules or boundaries
  - `code/backend/internal/dto/contest.go`
  - `code/backend/internal/module/contest/api/http`
  - `code/backend/internal/module/contest/application/queries`
  - `code/backend/internal/module/contest/ports`
  - `code/backend/internal/module/contest/infrastructure`
  - `code/frontend/src/api/contest.ts`
  - `code/frontend/src/api/admin/contests.ts`
- Dependencies
  - 需要保持现有单状态 `status` 查询兼容，不破坏已经依赖该接口的列表页。
- Validation
  - `go test` 跑 contest query / router / full router 相关测试。
- Review focus
  - `status` 与 `statuses` 共存时的兼容策略是否明确
  - 排序字段是否白名单约束，避免拼接注入
  - 学生端与平台端是否都能稳定拿到默认 `page_size=20`
- Risk notes
  - `/scoreboard` 当前依赖按开赛时间倒序展示，后端排序能力要先补齐，否则切到服务端分页后顺序会失真。

## Task 4
- Goal
  - 把 `/contests`、`/scoreboard`、`/scoreboard/:contestId`、`/platform/contest-ops/contests` 全部切到真实分页，并统一默认页大小为 `20`。
- Touched modules or boundaries
  - `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
  - `code/frontend/src/views/contests/ContestList.vue`
  - `code/frontend/src/features/scoreboard/model/*`
  - `code/frontend/src/views/scoreboard/*`
  - `code/frontend/src/features/platform-contests/model/useContestOperationsHubPage.ts`
  - `code/frontend/src/components/platform/contest/ContestOperationsHubWorkspacePanel.vue`
  - 相关前端测试
- Dependencies
  - 依赖 Task 3 的后端过滤/排序契约。
- Validation
  - `vitest` 跑 contests / scoreboard / contest-ops 相关视图与 feature 测试。
- Review focus
  - 是否彻底移除固定 `page_size=100` 拉全量再本地切页
  - 翻页、刷新、路由切换是否保持正确的请求 owner 与错误态
  - 平台 contest ops 的汇总卡片是否仍基于真实后端结果，而不是当前页凑数
- Risk notes
  - scoreboard 目录与 contest ops 英雄指标都依赖全量计数，切服务端分页后要避免把“当前页数量”误显示为“全量数量”。

## Integration Checks
- 教师页与平台页都能在相同筛选条件下显示同样的总数与页码。
- `status/keyword` 变化后应回到第 1 页，且上一轮请求结果不能覆盖新筛选。
- 默认 `page_size` 应稳定为 `20`，接口回包异常时前端不会进入非法页码。
- `/contests` 应显示真实服务端分页，而不是只在前端渲染第一页。
- `/scoreboard` 应按后端排序后的结果分页展示可查看排行的竞赛。
- `/scoreboard/:contestId` 应按接口分页展示排行榜明细，并支持刷新当前页。
- `/platform/contest-ops/contests` 应只展示可运维 AWD 赛事，并支持翻页。

## Rollback / Recovery Notes
- 前后端改动都集中在 AWD 复盘目录 owner，可整体回退，不涉及 migration 或配置变更。
- 若后端分页查询出现问题，可以回退到当前全量列表版本，不影响详情页和导出页。

## Residual Risks
- 若 contest 列表 summary 后续需要更复杂的跨状态聚合，可能要收敛成独立 summary 查询，而不是复用多次分页请求。
