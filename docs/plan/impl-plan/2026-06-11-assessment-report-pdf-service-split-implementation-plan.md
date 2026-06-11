<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 拆分 assessment report PDF service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or equivalent step-by-step execution. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `assessment/application/commands/report_service.go` 中的报表渲染实现拆到 `assessment/application/reporting`，让 `ReportService` 保留用例编排、权限、任务生命周期和导出任务状态职责。

**Architecture:** 这次不改 HTTP / application API，不改 repository / SQL，不改 PDF 输出语义。`commands` owner 报表用例编排、权限校验、数据读取和文件生命周期；`application/reporting` owner 报表输出数据结构、PDF / Excel / JSON / ZIP 写入、PDF 字体和渲染 helper。`commands` 只保留 AWD renderer 的薄兼容入口。

**Tech Stack:** Go, `gofpdf`, `excelize`, CTF `code-workflow`, package-local Go tests.

---

## Task Metadata

- Task Slug: `2026-06-11-assessment-report-pdf-service-split`
- Started At: `2026-06-11T01:54:12Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-assessment-report-pdf-service-split`
- Branch: `task/2026-06-11-assessment-report-pdf-service-split`

## Objective And Non-Goals

- Objective:
  - 将 `report_service.go` 从 2426 行的混合职责文件拆成更小的应用层文件。
  - `ReportService` 继续 owner 报表创建、异步任务、下载状态、数据读取和导出文件路径。
  - `application/reporting` owner PDF / Excel / JSON / ZIP 写入、报表输出结构和共享渲染 helper。
  - 保持现有导出格式、文件名、MIME、PDF 字体、AWD PDF 复用和测试行为不变。
- Non-Goals:
  - 不调整 API、DTO、OpenAPI、数据库 schema、repository SQL 或权限策略。
  - 不重新设计报表内容或 AWD 复盘报告语义。
  - 不把 renderer 提成跨模块公共包；当前复用只在 `assessment` bounded context 内成立。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `code/backend/tests/README.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-05/2026-05-31-report-status-polling-reporting-owner-tightening-plan.md`
  - `docs/todos/2026-06-02-security-review-findings.md` 中的 “报表 SQL 拼接模式收敛”只作为触达范围判断，本次不修改 repository / SQL。

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 当前任务是结构性拆分，触达受保护后端实现面和已有 oversized service。
  - 需要 task slug、implementation plan、startup gate、TDD red/green、completion-full 和 review gate。
  - 变更不改变外部行为，但会移动多个 renderer helper，存在编译、隐式复用和测试覆盖风险。

## Files

- Create:
  - `code/backend/internal/module/assessment/application/commands/report_renderer_structure_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_data.go`
  - `code/backend/internal/module/assessment/application/commands/report_generation.go`
  - `code/backend/internal/module/assessment/application/commands/report_review_archive_builder.go`
  - `code/backend/internal/module/assessment/application/commands/report_file_output.go`
  - `code/backend/internal/module/assessment/application/reporting/data.go`
  - `code/backend/internal/module/assessment/application/reporting/common.go`
  - `code/backend/internal/module/assessment/application/reporting/json_renderer.go`
  - `code/backend/internal/module/assessment/application/reporting/pdf_fonts.go`
  - `code/backend/internal/module/assessment/application/reporting/pdf_helpers.go`
  - `code/backend/internal/module/assessment/application/reporting/standard_pdf_renderer.go`
  - `code/backend/internal/module/assessment/application/reporting/spreadsheet_renderer.go`
  - `code/backend/internal/module/assessment/application/reporting/awd_review_renderer.go`
  - `code/backend/internal/module/assessment/application/reporting/pdf_helpers_test.go`
- Modify:
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
  - `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
  - `code/backend/internal/module/assessment/application/commands/report_writer_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_awd_review_builder_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_awd_review_render_test.go`
  - `docs/plan/impl-plan/2026-06-11-assessment-report-pdf-service-split-implementation-plan.md`
- Review:
  - `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go` 是否只保留薄兼容入口。
  - `code/backend/internal/module/assessment/application/reporting/*` 是否只承载输出结构与 renderer，不读取 repository。
  - `code/backend/internal/module/assessment/application/commands/report_generation.go` / `report_review_archive_builder.go` 是否继续承接应用数据读取。
- Test:
  - `go test ./internal/module/assessment/application/commands ./internal/module/assessment/application/reporting -run 'TestReportServiceKeepsRendererImplementationsOutOfServiceFile|TestReportServiceFileStaysFocused|TestCommandsDoNotOwnReportRenderingAdapters|TestWritePersonalPDFCreatesPDFFile|TestWritePersonalExcelCreatesWorkbook|TestWriteJSONReportCreatesJSONFile|TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary|TestNewReportPDFRegistersBoldFont|TestTeacherAWDReviewExportBuilderSelectsFocusRoundWhenRoundMissing|TestHottestRoundPrefersAttackDenseRound|TestTopRiskyServicePrefersCompromisedService|TestBuildAWDReviewSuggestionsIncludesTrafficOnlyHint' -count=1`
  - `go test ./internal/module/assessment/application/commands ./internal/module/assessment/application/reporting -count=1`
  - `go test ./internal/module/assessment/... -count=1`
  - `bash scripts/run-workflow-stage.sh completion-full`

## 复用与 Owner 决策

- Existing patterns searched:
  - `report_pdf_fonts.go` 已经把 PDF 字体注册单独放置。
  - `awd_review_export_renderer.go` 已经是 renderer 专属文件，并复用 `newReportPDF` / table helper / term localization。
  - `report_writer_test.go` 已有 PDF / Excel / JSON writer 行为测试。
- Reuse / extend / split / create-new decision:
  - 将 renderer 复用范围收口在 `assessment/application/reporting`，不引入跨 bounded context 的 shared abstraction。
  - 新增 package-local 结构护栏测试，避免 renderer 实现回流到 `commands` 或 `report_service.go`。
  - 按职责拆文件，而不是只按行数机械拆块。
- Owner boundary:
  - `report_service.go`：用例编排、生命周期、权限、异步任务和错误标记。
  - `report_data.go`：`commands` 对 `reporting` 输出结构的包内兼容别名和 `reportNow`。
  - `report_generation.go`：各报表类型的数据读取和生成入口。
  - `report_review_archive_builder.go`：学生复盘归档聚合与观察建议构建。
  - `report_file_output.go`：格式规范化、文件路径、下载文件名、响应 DTO 组装。
  - `application/reporting/data.go`：报表输出数据结构。
  - `application/reporting/pdf_helpers.go` / `pdf_fonts.go`：共享 PDF document、section、table、text sanitize、term localization 和字体资源消费。
  - `application/reporting/standard_pdf_renderer.go`：个人/班级训练报告 PDF 输出。
  - `application/reporting/spreadsheet_renderer.go`：个人/班级训练报告 Excel 输出。
  - `application/reporting/json_renderer.go`：JSON 导出写入。
  - `application/reporting/awd_review_renderer.go`：教师 AWD 复盘 ZIP / PDF 输出和报告建议。
- Why this is the narrowest safe surface:
  - 新包仍在 `assessment/application` 下，不跨模块、不触碰 HTTP、runtime、repository 或 SQL。
  - 不触碰 repository / SQL，因此不扩大到待办中的 SQL 安全收敛。
  - AWD builder 继续留在 `commands` 做用例侧选轮编排，renderer 只消费 archive DTO 并输出文件。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming` + `architect-agent` + `go-backend`
- Why this pass fits:
  - 这是结构性拆分，必须先确认模块 owner、调用路径和已有 renderer 复用点。
- grill-with-docs findings:
  - 后端架构文档明确 `assessment/application/commands/report_service.go` owner 报表用例编排，未要求 renderer 成为跨模块能力。
  - `awd_review_export_renderer.go` 已经证明 PDF helper 是 assessment report renderer 内部共享能力，不能被放进个人/班级专属文件。
  - `docs/todos/2026-06-02-security-review-findings.md` 的 SQL 收敛待办不在本次 touched surface；本次不改 `infrastructure/report_repository.go`。
- Plan adjustments after challenge:
  - 放弃“renderer 仍放 commands 同包”的方案，改为 `commands` 编排、`application/reporting` 渲染。
  - 数据读取 / repository 聚合仍留在 `commands` focused 文件，避免把用例构建误放进 renderer 包。
  - 增加结构测试而不是只靠人工 review 记住拆分边界。

## Validation

- Commands:
  - `bash scripts/check-startup-gate.sh`
  - TDD red: `go test ./internal/module/assessment/application/commands -run TestReportServiceKeepsRendererImplementationsOutOfServiceFile -count=1`
  - Focused green: `go test ./internal/module/assessment/application/commands ./internal/module/assessment/application/reporting -run 'TestReportServiceKeepsRendererImplementationsOutOfServiceFile|TestReportServiceFileStaysFocused|TestCommandsDoNotOwnReportRenderingAdapters|TestWritePersonalPDFCreatesPDFFile|TestWritePersonalExcelCreatesWorkbook|TestWriteJSONReportCreatesJSONFile|TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary|TestNewReportPDFRegistersBoldFont|TestTeacherAWDReviewExportBuilderSelectsFocusRoundWhenRoundMissing|TestHottestRoundPrefersAttackDenseRound|TestTopRiskyServicePrefersCompromisedService|TestBuildAWDReviewSuggestionsIncludesTrafficOnlyHint' -count=1`
  - Package: `go test ./internal/module/assessment/application/commands ./internal/module/assessment/application/reporting -count=1`
  - Module: `go test ./internal/module/assessment/... -count=1`
  - Completion: `bash scripts/run-workflow-stage.sh completion-full`
- Manual checks:
  - `wc -l` 确认 `report_service.go` 明显下降且 renderer 文件保持可读大小。
  - `rg` 确认 `commands` production code 不再导入 `archive/zip`、`gofpdf`、`excelize` 或 `reportassets`。
  - `rg` 确认 `application/reporting` 不读取 repository 或 lifecycle dependency。
- Review focus:
  - 移动代码是否保持行为等价。
  - `report_service.go` 是否仍然混入 renderer implementation。
  - `commands` 到 `reporting` 的依赖是否保持应用层内、同 bounded context 内。
  - AWD PDF renderer 是否仍能使用共享 helper，并保持建议文案等价。
  - import / package visibility 是否保持最小。

## Execution Checklist

### Task 1: TDD structure guard

**Files:**
- Create: `code/backend/internal/module/assessment/application/commands/report_renderer_structure_test.go`

- [x] **Step 1: Add a package-local test that parses `report_service.go` and fails when renderer implementation functions are declared there.**
- [x] **Step 2: Run the test and confirm it fails against the current 2426-line service file.**

### Task 2: Move renderer implementations out of `report_service.go`

**Files:**
- Create: `application/reporting/standard_pdf_renderer.go`, `pdf_helpers.go`, `spreadsheet_renderer.go`, `json_renderer.go`, `common.go`, `data.go`
- Modify: `report_service.go`

- [x] **Step 3: Move shared PDF helpers and term localization into `application/reporting/pdf_helpers.go`.**
- [x] **Step 4: Move personal/class PDF writers into `application/reporting/standard_pdf_renderer.go`.**
- [x] **Step 5: Move Excel writers and spreadsheet helpers into `application/reporting/spreadsheet_renderer.go`.**
- [x] **Step 6: Move JSON writer into `application/reporting/json_renderer.go`.**
- [x] **Step 7: Run `gofmt` on touched Go files.**

### Task 3: Verify behavior and workflow gates

**Files:**
- Review: all touched Go files and this plan.

- [x] **Step 8: Run focused green tests for structure, PDF, Excel, JSON, and AWD PDF helpers.**
- [x] **Step 9: Run full package tests for assessment commands after renderer split.**

### Task 4: Close remaining oversized service surface

**Files:**
- Create: `report_data.go`, `report_generation.go`, `report_review_archive_builder.go`, `report_file_output.go`
- Modify: `report_service.go`, `report_renderer_structure_test.go`

- [x] **Step 10: Add a failing focused-size guard for `report_service.go`.**
- [x] **Step 11: Move internal report data aliases and `reportNow` to `report_data.go`; output data structs live in `application/reporting/data.go`.**
- [x] **Step 12: Move report generation/data builder methods to `report_generation.go`.**
- [x] **Step 13: Move student review archive aggregation helpers to `report_review_archive_builder.go`.**
- [x] **Step 14: Move format/path/download response helpers to `report_file_output.go`.**
- [x] **Step 15: Run `gofmt` and focused green tests for both structure guards.**

### Task 5: Workflow validation and review

**Files:**
- Review: all touched Go files and this plan.

- [x] **Step 16: Run full package tests for assessment commands, reporting, and assessment module.**
- [x] **Step 17: Run `completion-full` workflow stage.**
- [x] **Step 18: Perform same-context code-reviewer pass after Claude review attempt timed out; do not treat this as an independent gate.**
