# AWD Phase 9 教师复盘归档与报告导出 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为教师工作台新增 AWD 赛事复盘目录与详情页，统一整场/单轮复盘事实源，并基于同一份 archive 交付证据归档包和教师阅读型报告，同时移除独立 `/academy/reports` 页面。

**Architecture:** 后端继续复用现有 `reports` 任务、轮询、下载和过期控制，但把 AWD 复盘的查询与导出业务拆成 assessment 模块内的新 `teacher awd review` read model、archive builder 和 renderer。前端新增教师端 AWD 复盘目录/详情页，路由与导航并入 `academy` 工作区；旧的班级报告总页面拆掉，改为在教师上下文页面里挂一个共享的班级报告导出对话框。

**Tech Stack:** Go, Gin, GORM, Vue 3, TypeScript, Vitest, Vue Test Utils, existing report task infrastructure, existing teacher workspace surface system

---

## Execution Notes

- 后端实现严格按 `@superpowers:test-driven-development` 先补 RED 测试，再做最小实现。
- 教师端新页面和迁移后的对话框必须显式使用 `@ctf-ui-theme-system`、`@frontend-skill`，必要时补 `@ctf-dark-surface-alignment`，保持和现有 `/academy/*` 一致。
- 任何测试、类型检查和构建命令都遵守 `@runtime-ops-safety`，全部在当前 worktree `/home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design` 内执行，不保留后台进程。
- 对外声称“完成”前，必须执行 `@superpowers:verification-before-completion`。
- `docs/superpowers/*` 仍被 `.gitignore` 覆盖；提交本计划和后续文档时，统一使用 `git add -f` 精确收录单文件。
- `/academy/reports` 已在 spec 中确认移除；后续实现时，不要只删路由，必须同步处理侧边栏、旧按钮跳转和相关测试。

## Planned File Map

### Backend: teacher AWD review read model / export

- Modify: `code/backend/internal/model/report.go`
  - 增加 `ReportTypeAWDReviewArchive`、`ReportTypeAWDReviewReport` 和 `ReportFormatZIP`。
- Modify: `code/backend/internal/dto/report.go`
  - 增加教师 AWD 复盘导出请求体。
- Create: `code/backend/internal/dto/teacher_awd_review.go`
  - 定义教师赛事目录、详情、整场/单轮 archive DTO。
- Create: `code/backend/internal/module/assessment/domain/awd_review.go`
  - 定义 archive 内部模型和教学观察结构。
- Create: `code/backend/internal/module/assessment/ports/awd_review.go`
  - 抽离教师 AWD 复盘查询仓储接口，避免继续把方法堆进 `ReportRepository`。
- Create: `code/backend/internal/module/assessment/infrastructure/awd_review_repository.go`
  - 负责从 contests / awd_rounds / awd_team_services / awd_attack_logs / awd_traffic_events / teams 读取复盘数据。
- Create: `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`
  - 负责目录查询、整场 archive 组装和单轮切片。
- Create: `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
  - 负责把查询结果整理成正式导出用 archive。
- Create: `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
  - 负责渲染 zip 证据包和 pdf 教师报告。
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 复用现有 report task 生命周期，接入 AWD 复盘导出。
- Create: `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
  - 提供教师 AWD 赛事目录、详情和导出入口。
- Modify: `code/backend/internal/module/assessment/api/http/report_handler.go`
  - 保持 status / download，不把教师 AWD 查询硬塞进去。
- Modify: `code/backend/internal/app/composition/assessment_module.go`
  - 装配 `TeacherAWDReviewHandler`、query service 和 export builder。
- Modify: `code/backend/internal/app/router_routes.go`
  - 注册 `/api/v1/teacher/awd/reviews*` 路由。
- Modify: `code/backend/internal/app/router_test.go`
  - 断言新路由和新模块字段暴露。
- Modify: `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - 补教师 AWD 复盘查询和导出集成路径。

### Backend tests

- Create: `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service_test.go`
  - 覆盖赛事目录、整场 archive、单轮切片和赛中/赛后状态。
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - 覆盖 AWD 复盘导出任务、文件扩展名和失败路径。
- Modify: `code/backend/internal/app/router_test.go`
  - 覆盖新路由与 assessment 模块暴露字段。
- Modify: `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - 覆盖教师创建 archive/report 任务、轮询、下载和权限拒绝。

### Frontend: teacher AWD review pages

- Modify: `code/frontend/src/api/contracts.ts`
  - 增加教师 AWD 赛事目录、详情、导出请求和 archive 数据类型。
- Modify: `code/frontend/src/api/teacher.ts`
  - 新增教师 AWD 赛事目录、详情和导出 API。
- Create: `code/frontend/src/composables/useTeacherAwdReviewIndex.ts`
  - 管理目录页加载、筛选和比赛跳转。
- Create: `code/frontend/src/composables/useTeacherAwdReviewDetail.ts`
  - 管理整场/单轮切换、导出、赛中/赛后状态和队伍下钻。
- Create: `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - 目录页容器。
- Create: `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  - 详情页容器。
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewIndexPage.vue`
  - 目录页展示层。
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewDetailPage.vue`
  - 详情页展示层。
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue`
  - 队伍下钻视图。
- Modify: `code/frontend/src/router/index.ts`
  - 新增 `/academy/awd-reviews` 和 `/academy/awd-reviews/:contestId`，移除 `/academy/reports`。
- Modify: `code/frontend/src/components/layout/Sidebar.vue`
  - 新增 `AWD复盘` 导航，去掉 `报告导出` 导航。

### Frontend: class report migration after `/academy/reports` removal

- Create: `code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue`
  - 承接原来 `/academy/reports` 的班级报告导出能力，但改为上下文对话框。
- Create: `code/frontend/src/composables/useTeacherClassReportExport.ts`
  - 从旧 `useTeacherReportExportPage` 抽取出可复用的导出和预览逻辑。
- Delete: `code/frontend/src/views/teacher/ReportExport.vue`
  - 移除独立页面。
- Delete: `code/frontend/src/composables/useTeacherReportExportPage.ts`
  - 被共享对话框替代。
- Modify: `code/frontend/src/views/teacher/ClassManagement.vue`
  - 将 `openReportExport` 改成弹出本地对话框。
- Modify: `code/frontend/src/views/teacher/TeacherClassStudents.vue`
  - 同上。
- Modify: `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
  - 同上。
- Modify: `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
  - 同上。

### Frontend tests

- Modify: `code/frontend/src/api/__tests__/teacher.test.ts`
  - 覆盖教师 AWD review API 和导出 API。
- Create: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
  - 覆盖目录页筛选和跳转。
- Create: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - 覆盖整场/单轮切换、赛中/赛后导出状态。
- Create: `code/frontend/src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts`
  - 覆盖旧导出页迁移后的对话框行为。
- Modify: `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
  - 覆盖新路由，去掉 `ReportExport` 断言。
- Modify: `code/frontend/src/router/__tests__/guards.test.ts`
  - 用新路由替代 `/academy/reports` 的 teacher guard 用例。
- Modify: `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
  - 覆盖侧边栏新 `AWD复盘` 入口并确认不再出现 `报告导出`。
- Modify: `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
  - 用新的 AWD 复盘页替代旧 `ReportExport.vue` surface 样本。
- Modify: `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
  - 同步新的教师页面来源。

## Task 1: 为后端教师 AWD 复盘查询与导出补 RED 测试

**Files:**
- Create: `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- Modify: `code/backend/internal/app/router_test.go`

- [ ] **Step 1: 先写教师 AWD 赛事目录和整场 archive 的失败测试**

在 `teacher_awd_review_service_test.go` 里先锁 4 个最小用例：

- `TestTeacherAWDReviewServiceListContestsReturnsOnlyAWDContests`
- `TestTeacherAWDReviewServiceGetContestArchiveBuildsOverviewAndRounds`
- `TestTeacherAWDReviewServiceGetContestArchiveSupportsSelectedRound`
- `TestTeacherAWDReviewServiceMarksEndedContestAsFinalSnapshot`

至少写出这种断言：

```go
if len(resp.Contests) != 1 || resp.Contests[0].Mode != model.ContestModeAWD {
	t.Fatalf("expected awd-only contest list, got %+v", resp.Contests)
}
if resp.Scope.SnapshotType != "live" {
	t.Fatalf("expected live snapshot, got %s", resp.Scope.SnapshotType)
}
if resp.SelectedRound == nil || resp.SelectedRound.Round.RoundNumber != 2 {
	t.Fatalf("expected selected round 2, got %+v", resp.SelectedRound)
}
```

- [ ] **Step 2: 写 AWD 复盘导出任务和文件名的失败测试**

在 `report_service_test.go` 新增：

- `TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask`
- `TestReportServiceCreateAWDReviewReportExportRejectsRunningContest`
- `TestReportDownloadFileNameUsesZIPForAWDReviewArchive`
- `TestReportDownloadFileNameUsesPDFForAWDReviewReport`

先按目标 API 写最小断言：

```go
resp, err := service.CreateTeacherAWDReviewArchive(context.Background(), teacher.ID, contest.ID, &dto.CreateTeacherAWDReviewExportReq{
	RoundNumber: intPtr(2),
})
if err != nil {
	t.Fatalf("CreateTeacherAWDReviewArchive() error = %v", err)
}
if resp.Status != model.ReportStatusProcessing {
	t.Fatalf("expected processing status, got %+v", resp)
}
```

- [ ] **Step 3: 给 router 测试补教师 AWD 复盘路由断言**

在 `router_test.go` 增加：

```go
assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews", "internal/module/assessment/api/http")
assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews/:id", "internal/module/assessment/api/http")
assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/archive", "internal/module/assessment/api/http")
assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/report", "internal/module/assessment/api/http")
```

- [ ] **Step 4: 运行定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
go test ./internal/module/assessment/application/queries -run TeacherAWDReview -count=1
go test ./internal/module/assessment/application/commands -run 'AWDReview|ReportDownloadFileName' -count=1
go test ./internal/app -run TestNewRouterRegistersStudentChallengeRoutes -count=1
```

预期：

- query 包因为新 service / dto / repo 不存在而 FAIL
- report service 因为新导出入口和新报告类型不存在而 FAIL
- router 测试因新教师 AWD 复盘路由不存在而 FAIL

- [ ] **Step 5: 提交 RED 测试**

```bash
git add code/backend/internal/module/assessment/application/queries/teacher_awd_review_service_test.go \
  code/backend/internal/module/assessment/application/commands/report_service_test.go \
  code/backend/internal/app/router_test.go
git commit -m "test(awd): 补充教师复盘归档后端红测"
```

## Task 2: 实现后端教师 AWD 复盘查询链路

**Files:**
- Create: `code/backend/internal/dto/teacher_awd_review.go`
- Create: `code/backend/internal/module/assessment/domain/awd_review.go`
- Create: `code/backend/internal/module/assessment/ports/awd_review.go`
- Create: `code/backend/internal/module/assessment/infrastructure/awd_review_repository.go`
- Create: `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`
- Create: `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- Modify: `code/backend/internal/app/composition/assessment_module.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`

- [ ] **Step 1: 先补 DTO 和领域结构**

在 `teacher_awd_review.go`、`awd_review.go` 里先把查询态结构写死，避免后面 service 一边查一边拼匿名 map。

目标至少包括：

```go
type TeacherAWDReviewContestResp struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	CurrentRound     *int       `json:"current_round,omitempty"`
	RoundCount       int        `json:"round_count"`
	TeamCount        int        `json:"team_count"`
	LatestEvidenceAt *time.Time `json:"latest_evidence_at,omitempty"`
	ExportReady      bool       `json:"export_ready"`
}

type TeacherAWDReviewArchiveResp struct {
	GeneratedAt time.Time                      `json:"generated_at"`
	Scope       TeacherAWDReviewScopeResp      `json:"scope"`
	Contest     TeacherAWDReviewContestMetaResp `json:"contest"`
	Overview    TeacherAWDReviewOverviewResp   `json:"overview"`
	Rounds      []TeacherAWDReviewRoundResp    `json:"rounds"`
	SelectedRound *TeacherAWDSelectedRoundResp `json:"selected_round,omitempty"`
}
```

- [ ] **Step 2: 写仓储接口和最小 SQL 查询**

在 `ports/awd_review.go` 先收敛仓储边界，不要在 service 里直接写 SQL。

建议先收成这组方法：

```go
type TeacherAWDReviewRepository interface {
	ListTeacherAWDReviewContests(ctx context.Context) ([]assessmentdomain.TeacherAWDReviewContestCard, error)
	FindTeacherAWDReviewContest(ctx context.Context, contestID int64) (*assessmentdomain.TeacherAWDReviewContestMeta, error)
	ListTeacherAWDReviewRounds(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewRoundSummary, error)
	ListTeacherAWDReviewTeams(ctx context.Context, contestID int64) ([]assessmentdomain.TeacherAWDReviewTeamSummary, error)
	ListTeacherAWDReviewRoundServices(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewServiceRecord, error)
	ListTeacherAWDReviewRoundAttacks(ctx context.Context, roundID int64) ([]assessmentdomain.TeacherAWDReviewAttackRecord, error)
	ListTeacherAWDReviewRoundTraffic(ctx context.Context, contestID, roundID int64) ([]assessmentdomain.TeacherAWDReviewTrafficRecord, error)
}
```

- [ ] **Step 3: 实现 query service 的整场和单轮切片**

在 `teacher_awd_review_service.go` 先让 service 支持：

- 列赛事目录
- 按 contest 获取整场 archive
- 可选 `round_number` 取单轮切片
- 根据 contest `status=ended` 切 `snapshot_type=final`

先做最小实现，不在第一版就引入复杂排序算法；教学观察可以先做规则聚合，例如：

```go
if round.ServiceAlertCount > 0 {
	observations = append(observations, assessmentdomain.TeacherAWDObservation{
		Key:      "service_alerts",
		Severity: "warning",
		Summary:  fmt.Sprintf("第 %d 轮出现 %d 个异常服务记录", round.RoundNumber, round.ServiceAlertCount),
	})
}
```

- [ ] **Step 4: 接上 HTTP handler 和 teacher 路由**

在 `teacher_awd_review_handler.go` 增加：

- `ListTeacherAWDReviews`
- `GetTeacherAWDReview`

查询参数先用：

```go
type GetTeacherAWDReviewReq struct {
	RoundNumber *int `form:"round"`
	TeamID      *int64 `form:"team_id"`
}
```

路由注册到 `teacherOrAbove` 下：

```go
teacherOrAbove.GET("/awd/reviews", deps.assessment.TeacherAWDReviewHandler.ListReviews)
teacherOrAbove.GET("/awd/reviews/:id", middleware.ParseInt64Param("id"), deps.assessment.TeacherAWDReviewHandler.GetReview)
```

- [ ] **Step 5: 运行后端查询与路由测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
go test ./internal/module/assessment/application/queries -run TeacherAWDReview -count=1
go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestCompositionModulesExposeContracts' -count=1
```

预期：PASS。

- [ ] **Step 6: 提交查询链路**

```bash
git add code/backend/internal/dto/teacher_awd_review.go \
  code/backend/internal/module/assessment/domain/awd_review.go \
  code/backend/internal/module/assessment/ports/awd_review.go \
  code/backend/internal/module/assessment/infrastructure/awd_review_repository.go \
  code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go \
  code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go \
  code/backend/internal/app/composition/assessment_module.go \
  code/backend/internal/app/router_routes.go \
  code/backend/internal/app/router_test.go
git commit -m "feat(awd): 增加教师复盘查询链路"
```

## Task 3: 实现 AWD 复盘导出任务并复用现有下载链路

**Files:**
- Modify: `code/backend/internal/model/report.go`
- Modify: `code/backend/internal/dto/report.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
- Create: `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- Create: `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- Modify: `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- Modify: `code/backend/internal/app/full_router_state_matrix_integration_test.go`

- [ ] **Step 1: 扩 report 类型、格式和请求 DTO**

在 `model/report.go` 增加：

```go
const (
	ReportTypeAWDReviewArchive = "awd_review_archive"
	ReportTypeAWDReviewReport  = "awd_review_report"
	ReportFormatZIP            = "zip"
)
```

在 `dto/report.go` 增加最小请求体：

```go
type CreateTeacherAWDReviewExportReq struct {
	RoundNumber *int `json:"round_number,omitempty"`
}
```

- [ ] **Step 2: 先写 builder，让 report service 不直接拼 archive**

在 `awd_review_export_builder.go` 里收一个明确入口：

```go
type AWDReviewExportBuilder interface {
	BuildArchive(ctx context.Context, contestID int64, roundNumber *int, snapshotType string) (*assessmentdomain.AWDReviewArchive, error)
}
```

第一版 `snapshotType` 约定：

- 比赛未结束：`live`
- 比赛已结束：`final`

- [ ] **Step 3: 再写 renderer，固定 archive=zip、report=pdf**

在 `awd_review_export_renderer.go` 先把输出格式固定下来，不再让前端传 format。

最小接口：

```go
func RenderAWDReviewArchiveZip(targetPath string, archive *assessmentdomain.AWDReviewArchive) error
func RenderAWDReviewReportPDF(targetPath string, archive *assessmentdomain.AWDReviewArchive) error
```

zip 包首版至少写入：

- `manifest.json`
- `overview.json`
- `rounds.json`
- `teams.json`
- `selected-round.json`（有选中轮时）

- [ ] **Step 4: 把 report service 接到新 builder / renderer**

在 `report_service.go` 里增加：

- `CreateTeacherAWDReviewArchive`
- `CreateTeacherAWDReviewReport`

并保持现有模式：

```go
s.runAsyncReport(report.ID, func(runCtx context.Context) error {
	archive, err := s.awdReviewBuilder.BuildArchive(runCtx, contestID, req.RoundNumber, snapshotType)
	if err != nil {
		return err
	}
	filePath, expiresAt, err := s.generateTeacherAWDReviewArchive(runCtx, report.ID, archive)
	if err != nil {
		return err
	}
	return s.repo.MarkReady(runCtx, report.ID, filePath, expiresAt)
})
```

教师报告导出要先卡住比赛状态：

```go
if contest.Status != model.ContestStatusEnded {
	return nil, errcode.New(errcode.ErrInvalidParams.Code, "教师复盘报告仅支持赛后导出", errcode.ErrInvalidParams.HTTPStatus)
}
```

- [ ] **Step 5: 在 teacher handler 上挂导出入口**

在 `teacher_awd_review_handler.go` 增加：

- `CreateReviewArchiveExport`
- `CreateReviewReportExport`

路由：

```go
teacherOrAbove.POST("/awd/reviews/:id/export/archive", deps.assessment.TeacherAWDReviewHandler.CreateArchiveExport)
teacherOrAbove.POST("/awd/reviews/:id/export/report", deps.assessment.TeacherAWDReviewHandler.CreateReportExport)
```

- [ ] **Step 6: 补 full-router 集成测试**

在 `full_router_state_matrix_integration_test.go` 增加：

- 教师查询 `/teacher/awd/reviews`
- 教师查询 `/teacher/awd/reviews/:id?round=2`
- 教师创建 archive 导出并轮询 ready
- 教师在 running 赛事上导出 report 被拒绝
- 教师在 ended 赛事上导出 report 成功

- [ ] **Step 7: 运行后端导出相关测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
go test ./internal/module/assessment/application/commands -run 'AWDReview|ReportDownloadFileName' -count=1
go test ./internal/app -run 'TestNewRouterRegistersStudentChallengeRoutes|TestBuildAssessmentModuleDelegatesToSubBuilders' -count=1
go test ./internal/app -run 'TestFullRouter' -count=1
```

预期：PASS。

- [ ] **Step 8: 提交导出链路**

```bash
git add code/backend/internal/model/report.go \
  code/backend/internal/dto/report.go \
  code/backend/internal/module/assessment/application/commands/report_service.go \
  code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go \
  code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go \
  code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go \
  code/backend/internal/module/assessment/application/commands/report_service_test.go \
  code/backend/internal/app/full_router_state_matrix_integration_test.go
git commit -m "feat(awd): 接入教师复盘导出任务"
```

## Task 4: 为前端 AWD 复盘页和导航迁移补 RED 测试

**Files:**
- Modify: `code/frontend/src/api/__tests__/teacher.test.ts`
- Create: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Create: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- Modify: `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- Modify: `code/frontend/src/router/__tests__/guards.test.ts`
- Modify: `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`

- [ ] **Step 1: 先写 teacher API 的失败测试**

在 `teacher.test.ts` 增加：

- `listTeacherAWDReviews`
- `getTeacherAWDReview`
- `exportTeacherAWDReviewArchive`
- `exportTeacherAWDReviewReport`

最小断言示例：

```ts
await listTeacherAWDReviews({ status: 'running', keyword: '春季' })

expect(requestMock).toHaveBeenCalledWith({
  method: 'GET',
  url: '/teacher/awd/reviews',
  params: {
    status: 'running',
    keyword: '春季',
  },
})
```

- [ ] **Step 2: 写目录页和详情页的失败测试**

在两个新测试文件里先锁定这些行为：

- 目录页会加载赛事目录并渲染“进入复盘”
- 详情页默认显示整场总览
- query 带 `round=2` 时会切到单轮摘要
- running 赛事上 `导出教师报告` 按钮禁用
- ended 赛事上两个导出按钮都可点击

至少写出这种断言：

```ts
expect(wrapper.text()).toContain('AWD复盘')
expect(wrapper.text()).toContain('进入复盘')
expect(wrapper.find('[data-testid=\"awd-review-export-report\"]').attributes('disabled')).toBeDefined()
```

- [ ] **Step 3: 先更新路由和侧边栏测试为新预期**

在 `sharedRoutes.test.ts` 和 `guards.test.ts` 里改成：

```ts
expect(findChild('academy/awd-reviews')?.name).toBe('TeacherAWDReviewIndex')
expect(findChild('academy/reports')).toBeUndefined()
```

在 `Sidebar.test.ts` 里新增 `AWD复盘`，并断言不再出现 `报告导出`。

- [ ] **Step 4: 更新教师 surface 测试的样本来源**

把 `ReportExport.vue` 的 raw import 换成两个新页面：

- `TeacherAWDReviewIndex.vue`
- `TeacherAWDReviewDetail.vue`

先让这些测试因为新文件不存在而 FAIL。

- [ ] **Step 5: 运行前端定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
npx vitest run src/api/__tests__/teacher.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts \
  src/router/__tests__/sharedRoutes.test.ts \
  src/router/__tests__/guards.test.ts \
  src/components/layout/__tests__/Sidebar.test.ts
```

预期：FAIL，失败点是新 API / 路由 / 页面 / 样式样本尚未实现。

- [ ] **Step 6: 提交前端红测**

```bash
git add code/frontend/src/api/__tests__/teacher.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts \
  code/frontend/src/router/__tests__/sharedRoutes.test.ts \
  code/frontend/src/router/__tests__/guards.test.ts \
  code/frontend/src/components/layout/__tests__/Sidebar.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
git commit -m "test(awd): 补充教师复盘页前端红测"
```

## Task 5: 实现教师端 AWD 复盘目录与详情页

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/teacher.ts`
- Create: `code/frontend/src/composables/useTeacherAwdReviewIndex.ts`
- Create: `code/frontend/src/composables/useTeacherAwdReviewDetail.ts`
- Create: `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
- Create: `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewIndexPage.vue`
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewDetailPage.vue`
- Create: `code/frontend/src/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue`
- Modify: `code/frontend/src/router/index.ts`
- Modify: `code/frontend/src/components/layout/Sidebar.vue`
- Modify: `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- Modify: `code/frontend/src/router/__tests__/guards.test.ts`
- Modify: `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`

- [ ] **Step 1: 先补前端契约和 teacher API**

在 `api/contracts.ts` 至少定义：

```ts
export interface TeacherAWDReviewContestItemData {
  id: ID
  title: string
  mode: ContestMode
  status: ContestStatus
  current_round?: number
  round_count: number
  team_count: number
  latest_evidence_at?: ISODateTime
  export_ready: boolean
}

export interface TeacherAWDReviewArchiveData {
  generated_at: ISODateTime
  scope: { contest_id: ID; round_number?: number; snapshot_type: 'live' | 'final' }
  contest: { id: ID; title: string; status: ContestStatus; mode: ContestMode; round_count: number; team_count: number }
  overview: { total_score: number; successful_attack_count: number; service_alert_count: number }
  rounds: TeacherAWDReviewRoundItemData[]
  selected_round?: TeacherAWDReviewSelectedRoundData
}
```

`teacher.ts` 里新增 4 个 API 方法，保持 `report_id` 统一转成字符串。

- [ ] **Step 2: 写两个 composable，先把数据流理顺**

`useTeacherAwdReviewIndex.ts` 负责：

- 加载赛事目录
- 状态和关键词筛选
- 跳详情

`useTeacherAwdReviewDetail.ts` 负责：

- 读取路由参数和 `round` query
- 加载详情
- 切换整场 / 单轮
- 发起 archive / report 导出并复用 `useReportStatusPolling`
- 控制队伍抽屉

- [ ] **Step 3: 实现目录页和详情页视图**

目录页先做“比赛卡片目录”，不要上来堆表格。

详情页先按这个骨架落：

- hero 区
- 总览指标区
- 轮次切换 rail
- 整场 or 单轮主体区
- 队伍下钻抽屉

运行中时按钮约束：

```ts
const canExportReport = computed(() => archive.value?.contest.status === 'ended')
```

- [ ] **Step 4: 接路由和侧边栏**

在 `router/index.ts` 增加：

```ts
{
  path: 'academy/awd-reviews',
  name: 'TeacherAWDReviewIndex',
  component: () => import('@/views/teacher/TeacherAWDReviewIndex.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: 'AWD复盘',
    icon: 'ScanEye',
    contentLayout: 'bleed',
  },
}
{
  path: 'academy/awd-reviews/:contestId',
  name: 'TeacherAWDReviewDetail',
  component: () => import('@/views/teacher/TeacherAWDReviewDetail.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: 'AWD赛事复盘',
    contentLayout: 'bleed',
  },
}
```

同时删掉 `academy/reports` 和 `teacher/reports`。

- [ ] **Step 5: 跑前端 AWD 复盘页定向测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
npx vitest run src/api/__tests__/teacher.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts \
  src/router/__tests__/sharedRoutes.test.ts \
  src/router/__tests__/guards.test.ts \
  src/components/layout/__tests__/Sidebar.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交教师 AWD 复盘页**

```bash
git add code/frontend/src/api/contracts.ts \
  code/frontend/src/api/teacher.ts \
  code/frontend/src/composables/useTeacherAwdReviewIndex.ts \
  code/frontend/src/composables/useTeacherAwdReviewDetail.ts \
  code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue \
  code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue \
  code/frontend/src/components/teacher/awd-review/TeacherAWDReviewIndexPage.vue \
  code/frontend/src/components/teacher/awd-review/TeacherAWDReviewDetailPage.vue \
  code/frontend/src/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue \
  code/frontend/src/router/index.ts \
  code/frontend/src/components/layout/Sidebar.vue
git commit -m "feat(awd): 新增教师复盘页面"
```

## Task 6: 移除 `/academy/reports` 并迁移班级报告导出为上下文对话框

**Files:**
- Create: `code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue`
- Create: `code/frontend/src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts`
- Create: `code/frontend/src/composables/useTeacherClassReportExport.ts`
- Modify: `code/frontend/src/views/teacher/ClassManagement.vue`
- Modify: `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- Modify: `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- Modify: `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- Delete: `code/frontend/src/views/teacher/ReportExport.vue`
- Delete: `code/frontend/src/composables/useTeacherReportExportPage.ts`
- Delete: `code/frontend/src/views/teacher/__tests__/ReportExport.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
- Modify: `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`

- [ ] **Step 1: 先把旧导出页逻辑抽成共享 composable**

把 `useTeacherReportExportPage.ts` 里真正有用的部分收进：

- 规范化班级名
- 预览数据加载
- 创建 class report 任务
- 轮询状态
- 下载文件

新 composable 只暴露对话框需要的状态，不再包含整页 hero 文案。

- [ ] **Step 2: 写共享导出对话框测试**

在 `TeacherClassReportExportDialog.test.ts` 先锁：

- 打开后会加载班级预览
- 点击创建导出任务会调用 `exportClassReport`
- `ready` 状态下可触发下载

核心断言示例：

```ts
expect(exportClassReportMock).toHaveBeenCalledWith({
  class_name: 'Class A',
  format: 'pdf',
})
expect(downloadReportMock).toHaveBeenCalledWith('102')
```

- [ ] **Step 3: 在四个教师视图里把旧跳转改成弹窗**

示例改法：

```ts
const reportDialogVisible = ref(false)

function openClassReportDialog() {
  reportDialogVisible.value = true
}
```

然后把原来的：

```ts
router.push({ name: 'ReportExport' })
```

全部替换成 `openClassReportDialog()`，并在模板底部挂对话框组件。

- [ ] **Step 4: 删除独立页面和旧测试**

删：

- `views/teacher/ReportExport.vue`
- `composables/useTeacherReportExportPage.ts`
- 旧 `ReportExport.test.ts`

再把 surface / eyebrow / dark-surface 测试里的 raw import 改成新 dialog 或 AWD review 页面样本。

- [ ] **Step 5: 运行迁移相关前端测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
npx vitest run src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts \
  src/views/teacher/__tests__/ClassManagement.test.ts \
  src/views/teacher/__tests__/TeacherClassStudents.test.ts \
  src/views/teacher/__tests__/TeacherStudentManagement.test.ts \
  src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts \
  src/views/teacher/__tests__/teacherSurface.test.ts \
  src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts \
  src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts \
  src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
```

预期：PASS。

- [ ] **Step 6: 提交 `/academy/reports` 迁移**

```bash
git add code/frontend/src/components/teacher/reports/TeacherClassReportExportDialog.vue \
  code/frontend/src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts \
  code/frontend/src/composables/useTeacherClassReportExport.ts \
  code/frontend/src/views/teacher/ClassManagement.vue \
  code/frontend/src/views/teacher/TeacherClassStudents.vue \
  code/frontend/src/views/teacher/TeacherStudentManagement.vue \
  code/frontend/src/views/teacher/TeacherStudentAnalysis.vue \
  code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts \
  code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts \
  code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
git rm code/frontend/src/views/teacher/ReportExport.vue \
  code/frontend/src/composables/useTeacherReportExportPage.ts \
  code/frontend/src/views/teacher/__tests__/ReportExport.test.ts
git commit -m "refactor(teacher): 移除独立报告导出页"
```

## Task 7: 端到端验证与收尾

**Files:**
- Modify: `docs/superpowers/plans/2026-04-12-awd-phase9-review-archive-implementation.md`
- Modify: any touched source files from Tasks 1-6 only if verification暴露问题

- [x] **Step 1: 运行后端完整定向验证**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/backend
go test ./internal/module/assessment/application/queries -count=1
go test ./internal/module/assessment/application/commands -count=1
go test ./internal/app -run 'TestNewRouter|TestFullRouter' -count=1
```

- [x] **Step 2: 运行前端定向验证**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
npx vitest run src/api/__tests__/teacher.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewIndex.test.ts \
  src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts \
  src/components/teacher/reports/__tests__/TeacherClassReportExportDialog.test.ts \
  src/router/__tests__/sharedRoutes.test.ts \
  src/router/__tests__/guards.test.ts \
  src/components/layout/__tests__/Sidebar.test.ts \
  src/views/teacher/__tests__/ClassManagement.test.ts \
  src/views/teacher/__tests__/TeacherClassStudents.test.ts \
  src/views/teacher/__tests__/TeacherStudentManagement.test.ts \
  src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts \
  src/views/teacher/__tests__/teacherSurface.test.ts \
  src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts \
  src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts \
  src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts
```

- [x] **Step 3: 跑前端类型检查**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/awd-phase9-review-archive-design/code/frontend
npx vue-tsc --noEmit
```

- [x] **Step 4: 记录计划勾选状态并整理 commit**

确认每个任务都已勾选、worktree 干净，再准备合并或后续执行方式。

- [x] **Step 5: 最终提交**

```bash
git status --short
git commit -m "feat(awd): 完成教师复盘归档与导出链路"
```
