# AWD 教师复盘归档与报告导出架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`contracts`
- 关联模块：
  - `internal/module/assessment/application/queries`
  - `internal/module/assessment/application/commands`
  - `internal/module/assessment/infrastructure`
  - `frontend/src/features/teacher-awd-review`
  - `frontend/src/widgets/teacher-awd-review`
- 关联文档：
  - `docs/architecture/features/赛事导出与复盘归档架构.md`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-29-awd-checker-completion`
- 最后更新：`2026-05-09`

## 1. 背景与问题

教师端当前已经有独立的 AWD 复盘目录、详情页和正式导出链路，不再依附管理员运维页或浏览器本地拼装 JSON。这里需要说明的，是“单场 AWD 赛事 archive 数据集”本身如何组织；通用 `reports` 生命周期和下载契约已经由共享专题单独 owning。

- 教师复盘以“单场 AWD 赛事”为主身份
- 页面详情与导出归档共用同一份 contest archive 读模型
- 归档 ZIP 与评估 PDF 已接入统一报告生命周期，而不是临时下载逻辑

## 2. 架构结论

- 教师 AWD 复盘的正式只读模型是 `TeacherAWDReviewArchiveResp`。
- `TeacherAWDReviewService` 负责目录和详情查询，`ReportService` 负责异步导出，二者共享同一份 `AWDReviewExportBuilder`。
- 通用 `reports` 生命周期、状态轮询和下载契约由 `赛事导出与复盘归档架构.md` owning；本文只收口 AWD 赛事级 archive 数据集和导出门禁。
- ZIP 归档可对任意 AWD 赛事生成当前快照；PDF 报告只允许在赛事 `ended` 后导出。
- 页面详情和导出归档都来自同一份 archive builder，不允许各自拼一套事实。
- AWD 赛事级复盘 archive 继续独立于个人 / 班级报告；不会把队伍分数直接回写成学生个人 `score / solved / rank`。
- 前端正式入口是 `/academy/awd-reviews` 和 `/academy/awd-reviews/:contestId`。

## 3. 模块边界与职责

### 3.1 模块清单

- `TeacherAWDReviewRepository`
  - 负责：读取 contest、round、team、service、attack、traffic 的教师复盘原始记录
  - 不负责：导出文件渲染

- `TeacherAWDReviewService`
  - 负责：把 repository 结果组装成 `TeacherAWDReviewArchiveResp`
  - 不负责：重新定义通用 report 任务生命周期

- `AWDReviewExportBuilder`
  - 负责：复用 `GetContestArchive` 构建导出输入
  - 不负责：文件格式渲染

- `ReportService`
  - 负责：复用通用 report 任务链创建 AWD review archive / report 导出任务
  - 不负责：重新组织 AWD 复盘事实

- `RenderAWDReviewArchiveZip` / `RenderAWDReviewReportPDF`
  - 负责：把 archive 落为 ZIP / PDF
  - 不负责：查询数据库

### 3.2 事实源与所有权

- 复盘 archive 事实源：`TeacherAWDReviewRepository + TeacherAWDReviewService`
- 导出任务事实源：`reports` 生命周期与 `ReportService`
- 前端详情态 owner：`useTeacherAwdReviewDetail`

## 4. 关键模型与不变量

### 4.1 核心实体

- `TeacherAWDReviewArchiveResp`
  - `generated_at`
  - `scope`
  - `contest`
  - `overview`
  - `rounds`
  - `selected_round`

- `TeacherAWDReviewScopeResp`
  - `snapshot_type`
  - `requested_by`
  - `requested_id`

- 导出文件
  - 归档：`ZIP`
  - 报告：`PDF`

### 4.2 不变量

- `snapshot_type` 只分为：
  - `live`：赛事未结束
  - `final`：赛事已结束
- `team_id` 查询必须与 `round` 一起使用；后端会拒绝“只给 team 不给 round”的请求。
- `export_ready` 由 contest 状态是否 `ended` 决定，前端据此控制是否允许导出 PDF 报告。
- AWD 赛事导出的任务状态、下载轮询和过期控制复用共享 `reports` 契约，不在本文重复定义第二套任务模型。
- AWD 赛事 archive 只服务赛事级复盘；个人 / 班级报告仍沿用练习型统计口径，AWD 个人事件通过证据链和归档页表达，而不是直接抬高 `score / solved / rank`。
- 归档 ZIP 至少包含：
  - `manifest.json`
  - `overview.json`
  - `rounds.json`
  - `teams.json`
  - 可选的 `selected-round.json`

## 5. 关键链路

### 5.1 教师目录与详情链路

1. 教师打开 `/academy/awd-reviews`。
2. 前端调用 `listTeacherAWDReviews`，后端列出所有 AWD 赛事及 `current_round`、`round_count`、`team_count`、`latest_evidence_at`、`export_ready`。
3. 进入详情页后，`useTeacherAwdReviewDetail` 调用 `getTeacherAWDReview`。
4. 后端按赛事维度返回 overview、round 列表，必要时附带 `selected_round` 的 service / attack / traffic 明细。

### 5.2 导出链路

1. 教师在详情页点击“归档导出”或“生成评估报告”。
2. `TeacherAWDReviewHandler` 调用 `ReportService.CreateTeacherAWDReviewArchive` 或 `CreateTeacherAWDReviewReport`。
3. `ReportService` 创建 `report` 记录，后台任务通过 `AWDReviewExportBuilder.BuildArchive` 读取当前 archive。
4. ZIP 或 PDF 渲染完成后，生命周期仓库把状态更新为 `ready`；前端轮询和下载契约复用共享 report 任务链。

## 6. 接口与契约

### 6.1 后端接口

- `GET /teacher/awd/reviews`
- `GET /teacher/awd/reviews/:id`
- `POST /teacher/awd/reviews/:id/archive`
- `POST /teacher/awd/reviews/:id/report`

### 6.2 前端导出契约

- `exportArchive` 不要求赛事结束
- `exportReport` 仅在 `canExportReport = true` 时开放
- 详情页 round 切换通过 query `round` 驱动，不另建第二套路由
- 任务状态轮询与下载接口不在本文重复列出，复用 `赛事导出与复盘归档架构.md` 的共享契约

## 7. 兼容与迁移

- 当前没有单独的“教师 AWD 报告系统”；AWD 导出复用通用 `reports` 任务链。
- 详情页默认按整场展示；按 round 下钻是附加切片能力，不改变主身份。
- 当前前端详情页没有直接把 `team_id` 放进详情查询，而是在已加载的 `selected_round` 内做本地 team drawer 细分。

## 8. 代码落点

- `code/backend/internal/module/assessment/infrastructure/awd_review_repository.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/app/router_routes.go`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- `code/frontend/src/features/teacher-awd-review/model/useTeacherAwdReviewDetail.ts`
- `code/frontend/src/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue`

## 9. 验证标准

- 教师能从 `/academy/awd-reviews` 查看 AWD 赛事目录并进入详情页。
- 详情页与 ZIP 导出读取的是同一份 archive 事实，而不是两套查询拼装。
- 未结束赛事可导出归档 ZIP，已结束赛事额外可导出 PDF 报告。
- 导出任务通过统一 `report status` 轮询完成，不需要额外导出通道。
