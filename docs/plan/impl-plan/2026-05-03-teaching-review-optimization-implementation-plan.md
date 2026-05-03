# Teaching Review Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将教师学生分析页的“攻防证据链”升级为围绕证据事件、攻击会话和教学观察点组织的复盘工作台，并让实时页面与归档复盘的事实口径保持一致。

**Architecture:** 后端以 `teaching_readmodel` 为读模型边界，先收敛证据事件构建，再基于事件聚合 `AttackSession`，避免新增事实写入链路。前端以 `teacher-student-analysis` feature 为页面状态入口，将超大展示组件拆成聚焦的复盘工作台组件和 presentation helpers，避免继续扩张 1000+ 行组件。

**Tech Stack:** Go 1.24, Gin, GORM, PostgreSQL, Vue 3, TypeScript, Pinia, Vue Router, Vitest.

---

## Plan Summary

### Objective

完成 `docs/architecture/features/teaching-review-optimization-design.md` 的第一阶段可交付版本：

- 教师实时证据纳入普通训练、AWD 攻击、AWD 流量、Writeup 和人工评审。
- 新增或整理 `GET /api/v1/teacher/students/:id/attack-sessions`，按会话展示解题和攻击过程。
- 学生分析页默认展示复盘工作台，而不是只展示平铺事件。
- 复盘归档与实时页面共享或对齐事件类型、字段名和统计口径。
- 增加最小必要筛选、观察点和测试，避免继续扩大已有大文件。

### Non-goals

- 不做完整流量回放、pcap 抓包、HAR 下载或命令级录制。
- 不新增持久化 `attack_sessions` 表；第一版按查询时聚合。
- 不自动判断漏洞利用类型。
- 不自动生成最终教师评价结论；观察点只作为提示。
- 不重构整个教师端页面体系，只处理学生分析页复盘区相关边界。

### Source architecture or design docs

- `docs/architecture/features/teaching-review-optimization-design.md`
- `docs/architecture/features/attack-session-replay-evolution-design.md`
- `AGENTS.md`

### Brainstorming Evidence

已读取的本地依据：

- 后端接口和 DTO：`code/backend/internal/dto/teacher.go`、`code/backend/internal/module/teaching_readmodel/api/http/handler.go`、`code/backend/internal/app/router_routes.go`
- 后端读模型：`code/backend/internal/module/teaching_readmodel/application/queries/service.go`、`code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`、`code/backend/internal/module/teaching_readmodel/ports/query.go`
- 前端 API 和状态：`code/frontend/src/api/teacher/students.ts`、`code/frontend/src/api/contracts.ts`、`code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- 前端展示：`code/frontend/src/components/teacher/StudentInsightPanel.vue`、`code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`、`code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- 当前工作区已有未提交改动已经包含 `attack-sessions` 路由、DTO、服务聚合、前端 API 和部分会话展示。

当前关键约束：

- `StudentInsightPanel.vue` 约 1424 行，`StudentAnalysisPage.vue` 约 725 行，后续不得继续直接堆功能；新增复盘区必须拆组件或 helpers。
- 当前 `teaching_readmodel.Repository.GetStudentEvidence` 已经承担多源事件查询，后续应先整理边界，再补筛选和归档对齐。
- 后端业务时间按仓库规则需要使用 UTC；本计划涉及返回和比较时间时必须检查 `.UTC()` 边界。

### Architecture Evaluation

这是本计划在开始编码前必须通过的评估关卡，不属于可选说明。

评估结论：

- 当前任务不只是“把教师复盘页做得更完整”，还包含一个明确的读模型收敛目标：实时证据、攻击会话和复盘归档不应长期各自维护一套相近但不相同的 evidence 构建逻辑。
- 如果只按功能顺序推进，很容易先完成“结果对齐”，再在收尾阶段才发现还需要重新设计 evidence builder 的共享边界，造成二次返工。
- 因此，这个功能的实现必须同时区分两类工作：
  - **结果对齐任务**：补筛选、补事件源、补前端工作台、补观察点。
  - **结构收敛任务**：明确统一 evidence event builder 的 owner、共享层落点、实时复盘与归档复盘的复用边界。

本计划要求的架构落点：

- 统一证据事件的事实语义应只有一套，不允许 `teaching_readmodel` 和 `assessment` 长期各自维护相近的事件转换规则。
- `AttackSession` 聚合继续属于 `teaching_readmodel` 应用层能力，不下沉到归档模块。
- 复盘归档应消费统一 evidence event，而不是自行重新理解底层表结构。
- 若第一阶段因为风险控制只做到“结果对齐”，必须把“共享 evidence builder 收敛”明确列成后续实施任务，而不是留在口头约定里。

本计划在编码前的风险判断：

- 高风险返工点不是 API 参数，而是 evidence builder 的共享边界。
- 如果实现过程中出现“功能已通，但归档和实时仍各查各拼”的状态，不能直接视为计划完成，只能视为阶段性对齐。
- 若出现这种情况，必须回写 impl-plan，补上单独的结构收敛任务和完成标准，再继续实现。

### Dependency order

1. 先冻结和评估现有未提交改动。
2. 先完成 evidence builder 的架构评估，明确“结果对齐”和“结构收敛”的任务边界。
3. 后端补齐证据查询契约、筛选、测试。
4. 后端整理会话聚合，保证 API 合同稳定。
5. 前端抽离复盘工作台，接入筛选和观察点。
6. 归档口径对齐。
7. 收敛共享 evidence builder 或明确记录其作为单独后续任务的完成标准。
8. 集成验证、独立 review、修复后重跑受影响验证。

### Expected specialist skills

- `backend-engineer`
- `frontend-engineer`
- `code-reviewer`
- `test-engineer`
- `doc-admin-agent`

## File Structure

### Backend

- Modify: `code/backend/internal/dto/teacher.go`
  - 扩展 `TeacherEvidenceQuery`、`TeacherEvidenceSummary`、`TeacherEvidenceEvent`，补齐筛选和事件字段。
- Modify: `code/backend/internal/module/teaching_readmodel/ports/query.go`
  - 将 evidence 查询参数从单一 `challengeID` 提升为读模型查询结构，避免继续扩张函数签名。
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
  - 整理 evidence DTO mapper、attack session 聚合、筛选、分页、summary。
- Modify: `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
  - 收敛多源证据查询，补齐 `contest_id`、`round_id`、`event_type`、`from`、`to`、分页边界。
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
  - 绑定新增查询参数，保持错误响应一致。
- Modify: `code/backend/internal/app/router_routes.go`
  - 仅当现有路由缺失或命名不一致时修改。
- Test: `code/backend/internal/app/router_test.go`
- Test: `code/backend/internal/app/practice_flow_integration_test.go`
- Test: add focused tests near `code/backend/internal/module/teaching_readmodel/...` if existing harness supports repository/service unit tests.
- Modify later: assessment report archive files after identifying exact reuse point:
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/infrastructure/report_repository.go`

### Frontend

- Modify: `code/frontend/src/api/contracts.ts`
  - 稳定 `TeacherEvidence*`、`TeacherAttackSession*`、filter query types。
- Modify: `code/frontend/src/api/teacher/students.ts`
  - 给 `getStudentEvidence`、`getStudentAttackSessions` 增加 query 参数和 normalization。
- Modify: `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
  - 页面状态只负责编排，筛选和复盘派生逻辑拆出。
- Create: `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
  - 负责筛选状态、加载会话、观察点派生、路由 query 同步。
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue`
  - 复盘工作台容器。
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewSessionList.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewEventTimeline.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewObservationStrip.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/model/presentation.ts`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/index.ts`
- Modify: `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - 移除内联复盘工作台模板和 helpers，只挂载新 widget。
- Modify: `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - 仅传递必要 props/events；不新增大段模板逻辑。
- Test: `code/frontend/src/api/__tests__/teacher.test.ts`
- Test: `code/frontend/src/widgets/teacher-student-review-workspace/*.test.ts`
- Test: `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Task 1: Baseline Audit and Current Diff Cleanup

**Goal:** 先确认现有未提交改动哪些属于本功能、哪些是前一轮遗留问题，避免在错误基础上继续写代码。

**Files:**
- Inspect: all files in `git status --short`
- Modify only if needed: no production code in this task unless removing accidental duplication
- Test: none

- [ ] **Step 1: Capture current status**

Run:

```bash
git -C /home/azhi/workspace/projects/ctf status --short
git -C /home/azhi/workspace/projects/ctf diff --stat
```

Expected: list current design docs and implementation files already touched.

- [ ] **Step 2: Classify existing changes**

Record in working notes:

- belongs to teaching review optimization
- belongs to attack session replay evolution
- unrelated or accidental

- [ ] **Step 3: Check large-file risk**

Run:

```bash
wc -l code/frontend/src/components/teacher/StudentInsightPanel.vue \
  code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue \
  code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts
```

Expected: confirm `StudentInsightPanel.vue` and `StudentAnalysisPage.vue` require extraction before adding more UI.

- [ ] **Step 4: Commit boundary note**

Do not commit yet if previous changes are unreviewed. Use this task to decide the smallest reviewable first patch.

## Task 2: Backend Evidence Query Contract

**Goal:** 将教师证据接口从单一 `challenge_id` 查询扩展为可分页、可筛选的读模型契约。

**Files:**
- Modify: `code/backend/internal/dto/teacher.go`
- Modify: `code/backend/internal/module/teaching_readmodel/ports/query.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Test: `code/backend/internal/app/practice_flow_integration_test.go`

- [ ] **Step 1: Write or extend failing API contract test**

Add coverage for:

- `GET /api/v1/teacher/students/:id/evidence?challenge_id=:id`
- `event_type=instance_proxy_request`
- `from` / `to`
- `limit` / `offset`

Expected before implementation: unsupported filters either ignored or not represented in response.

- [ ] **Step 2: Extend DTO query**

Add fields:

```go
type TeacherEvidenceQuery struct {
    ChallengeID *int64     `form:"challenge_id" binding:"omitempty,min=1"`
    ContestID   *int64     `form:"contest_id" binding:"omitempty,min=1"`
    RoundID     *int64     `form:"round_id" binding:"omitempty,min=1"`
    EventType   string     `form:"event_type" binding:"omitempty,max=64"`
    From        *time.Time `form:"from" binding:"omitempty"`
    To          *time.Time `form:"to" binding:"omitempty"`
    Limit       int        `form:"limit" binding:"omitempty,min=1,max=100"`
    Offset      int        `form:"offset" binding:"omitempty,min=0"`
}
```

Use UTC normalization before repository calls.

- [ ] **Step 3: Replace repository signature**

Prefer a repository query struct over adding many parameters:

```go
type EvidenceQuery struct {
    ChallengeID *int64
    ContestID   *int64
    RoundID     *int64
    EventType   string
    From        *time.Time
    To          *time.Time
    Limit       int
    Offset      int
}
```

- [ ] **Step 4: Run backend focused tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/app -run 'TestPracticeFlow|TestRoute' -count=1
```

Expected: PASS.

**Review focus:** API compatibility, UTC time handling, pagination defaults, forbidden cross-class access unchanged.

**Risk notes:** If repository-level pagination is applied before all event sources are merged, ordering can become incorrect. First version may merge then paginate in application/repository memory with a documented cap.

## Task 3: Backend Evidence Builder and AWD Personal Scope

**Goal:** 整理 `GetStudentEvidence` 多源构建逻辑，保证 AWD、Writeup、人工评审、普通提交和代理请求字段一致。

**Files:**
- Modify: `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Test: add focused repository/service tests if feasible, otherwise extend integration flow.

- [ ] **Step 1: Add tests for event source coverage**

Cover at least:

- `instance_proxy_request` has `request_method`, `target_path`, `status_code`, `payload_preview`.
- `awd_attack_submission` appears for `submitted_by_user_id`.
- team-level AWD fallback uses `meta.scope = "team"`.
- `writeup` and `manual_review` appear with review metadata.

- [ ] **Step 2: Extract small builder helpers**

Keep helpers private to repository or move to a focused file only if it reduces file size materially:

- `appendInstanceAccessEvidence`
- `appendProxyRequestEvidence`
- `appendSubmissionEvidence`
- `appendManualReviewEvidence`
- `appendWriteupEvidence`
- `appendAWDAttackEvidence`
- `appendAWDTrafficEvidence`

- [ ] **Step 3: Normalize meta**

Required canonical keys:

- `request_method`
- `target_path`
- `target_query`
- `status_code`
- `payload_preview`
- `payload_truncated`
- `is_correct`
- `is_success`
- `points`
- `score_gained`
- `scope`

- [ ] **Step 4: Run backend package tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/teaching_readmodel/... -count=1
go test ./internal/app -run 'TestPracticeFlow' -count=1
```

Expected: PASS.

**Review focus:** no N+1 query introduced for common page loads, no accidental exposure of full payloads, team-level AWD evidence clearly marked.

**Risk notes:** `awd_traffic_events` can be large. Keep explicit limit or query window; do not silently return unbounded traffic.

## Task 4: Backend Attack Session Aggregation

**Goal:** 稳定 `GET /api/v1/teacher/students/:id/attack-sessions` 的聚合规则、过滤和 summary。

**Files:**
- Modify: `code/backend/internal/dto/teacher.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Test: `code/backend/internal/app/router_test.go`
- Test: focused service tests for session grouping if possible.

- [ ] **Step 1: Test grouping rules**

Cover:

- practice: `student_id + challenge_id`
- jeopardy: `student/team + contest_id + challenge_id`
- awd: `student/team + contest_id + service_id + victim_team_id`
- split when event gap exceeds `1h`

- [ ] **Step 2: Test result derivation**

Cases:

- successful normal submission -> `success`
- successful AWD attack -> `success`
- failed submission only -> `failed`
- access/proxy only -> `in_progress`
- writeup/manual review only -> `unknown`

- [ ] **Step 3: Ensure `with_events=false` works**

Expected: response includes summary and session cards, omits event arrays.

- [ ] **Step 4: Run focused backend tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/teaching_readmodel/... -count=1
go test ./internal/app -run 'TestRoute|TestPracticeFlow' -count=1
```

Expected: PASS.

**Review focus:** deterministic session IDs, stable sort order, pagination after filtering, no mutation bug from slicing shared arrays.

**Risk notes:** Query-time aggregation is acceptable for phase one, but must keep limits explicit.

## Task 5: Frontend API and State Extraction

**Goal:** 前端先把数据契约、筛选状态和加载编排整理清楚，再改 UI。

**Files:**
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/teacher/students.ts`
- Modify: `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- Create: `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
- Test: `code/frontend/src/api/__tests__/teacher.test.ts`
- Test: create `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.test.ts` if project test harness supports composable tests.

- [ ] **Step 1: Add API normalization tests**

Cover numeric ID normalization for sessions/events and query serialization for:

- `mode`
- `challenge_id`
- `contest_id`
- `round_id`
- `result`
- `with_events`
- `limit`
- `offset`

- [ ] **Step 2: Implement query-aware API functions**

Use typed query objects:

```ts
export interface TeacherAttackSessionQuery {
  mode?: 'practice' | 'jeopardy' | 'awd'
  challenge_id?: string
  contest_id?: string
  round_id?: string
  result?: 'success' | 'failed' | 'in_progress' | 'unknown'
  with_events?: boolean
  limit?: number
  offset?: number
}
```

- [ ] **Step 3: Extract review workspace state**

`useTeacherReviewWorkspace` owns:

- filter refs
- loading/error for sessions
- `loadAttackSessions`
- derived summary
- observation items
- optional route query sync

- [ ] **Step 4: Keep page composable as orchestrator**

`useTeacherStudentAnalysisPage` should load core student data and delegate review workspace behavior instead of accumulating more state.

- [ ] **Step 5: Run frontend tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run test:run -- src/api/__tests__/teacher.test.ts src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.test.ts
npm run typecheck
```

Expected: PASS.

**Review focus:** no duplicated source of truth between page and workspace composable, no stale response race when switching students, query values stay strings where IDs are used by UI.

**Risk notes:** If current tests do not mock router cleanly, keep route query sync out of the first patch and add it in Task 7.

## Task 6: Frontend Review Workspace Components

**Goal:** 从 `StudentInsightPanel.vue` 中抽离复盘工作台，控制组件尺寸并提升可测试性。

**Files:**
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewSessionList.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewEventTimeline.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewObservationStrip.vue`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/model/presentation.ts`
- Create: `code/frontend/src/widgets/teacher-student-review-workspace/index.ts`
- Modify: `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- Test: `code/frontend/src/widgets/teacher-student-review-workspace/*.test.ts`

- [ ] **Step 1: Move presentation helpers first**

Move and test:

- `sessionResultLabel`
- `sessionModeLabel`
- `eventTypeLabel`
- `eventMetaItems`
- `sessionPathSummary`

- [ ] **Step 2: Create session list and event timeline**

Keep props narrow:

```ts
defineProps<{
  sessions: TeacherAttackSessionData[]
  evidenceSummary?: TeacherEvidenceSummaryData
}>()
```

- [ ] **Step 3: Replace inline section in `StudentInsightPanel.vue`**

`StudentInsightPanel.vue` should render one workspace component inside the existing tab section. Do not add more long template branches.

- [ ] **Step 4: Add component tests**

Cover:

- empty state
- summary metrics
- session result label
- event meta pills for proxy/AWD/writeup/manual review

- [ ] **Step 5: Run targeted frontend tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run test:run -- src/widgets/teacher-student-review-workspace src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
npm run typecheck
```

Expected: PASS.

**Review focus:** large-file reduction, accessible list semantics, no visible implementation/design explanation copy, text overflow on pills.

**Risk notes:** Avoid introducing a card-inside-card visual pattern; reuse existing teacher panel surfaces.

## Task 7: Filters and Teaching Observation Points

**Goal:** 增加最小必要筛选与教师观察点，让复盘工作台可用于课堂判断。

**Files:**
- Modify: `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
- Modify/Create: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewWorkspaceFilters.vue`
- Modify: `code/frontend/src/widgets/teacher-student-review-workspace/TeacherReviewObservationStrip.vue`
- Test: presentation/composable tests.

- [ ] **Step 1: Implement filters**

First version supports:

- mode
- result
- event type
- challenge id if current page already has challenge context

Defer complex date picker unless the existing design system has a matching compact control.

- [ ] **Step 2: Implement observation derivation**

Observation keys:

- `has_practice_request`
- `repeated_failed_submission`
- `visited_without_submission`
- `success_without_writeup`
- `awd_success_without_review_material`

- [ ] **Step 3: Add tests for observation derivation**

Use fixture sessions/events and assert stable observation output.

- [ ] **Step 4: Run tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run test:run -- src/widgets/teacher-student-review-workspace src/features/teacher-student-analysis/model
npm run typecheck
```

Expected: PASS.

**Review focus:** observations must not pretend to be final teacher judgment, filters must not trigger duplicate requests on mount.

**Risk notes:** Route query sync can create loops; add it only with tests.

## Task 8: Review Archive Alignment

**Goal:** 让复盘归档与实时复盘页面的事件类型、字段名和 summary 统计口径一致。

**Files:**
- Inspect/Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
- Inspect/Modify: `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- Modify tests: `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
- Modify tests: `code/backend/internal/module/assessment/application/commands/report_service_test.go`

- [ ] **Step 1: Identify current archive evidence builder**

Run:

```bash
rg -n "BuildStudentReviewArchive|GetStudentEvidence|ReviewArchiveEvidence" code/backend/internal/module/assessment
```

- [ ] **Step 2: Decide reuse level**

Preferred:

- shared mapper or shared contract-level event builder if dependency direction stays clean.

Acceptable phase-one fallback:

- duplicate query remains, but event types, canonical meta keys, summary counts and AWD inclusion match `teaching_readmodel`.

- [ ] **Step 3: Add archive alignment tests**

Cover:

- AWD attack evidence exists in archive.
- proxy meta uses `request_method` / `target_path` / `status_code`.
- submit/success counts match realtime evidence for same fixture.

- [ ] **Step 4: Run assessment tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/assessment/... -count=1
```

Expected: PASS.

**Review focus:** no dependency cycle between `assessment` and `teaching_readmodel`, no contract drift between archive JSON and realtime UI.

**Risk notes:** Full reuse may require a shared read-model package; do not introduce broad architectural movement unless the diff stays small.

## Task 9: Integration Checks and Documentation Sync

**Goal:** 完成端到端验证、review 证据归档和必要文档同步。

**Files:**
- Modify if needed: `docs/architecture/features/teaching-review-optimization-design.md`
- Create review evidence after review: `docs/reviews/{frontend|backend|general}/2026-05-03-teaching-review-optimization-review-*.md`
- Update this plan checklist as tasks complete.

- [ ] **Step 1: Backend full targeted verification**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/module/teaching_readmodel/... ./internal/module/assessment/... ./internal/app -count=1
```

Expected: PASS.

- [ ] **Step 2: Frontend targeted verification**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run test:run -- src/api/__tests__/teacher.test.ts src/widgets/teacher-student-review-workspace src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
npm run typecheck
```

Expected: PASS.

- [ ] **Step 3: Independent review**

Use `code-reviewer` with focus:

- correctness and access control
- backend query bounds and UTC handling
- frontend file size and responsibility boundaries
- test quality and missing negative cases
- archive/realtime contract alignment

Archive review result under:

```text
docs/reviews/general/2026-05-03-teaching-review-optimization-review-implementation.md
```

- [ ] **Step 4: Fix material findings**

For each P0/P1/P2 review finding:

- apply smallest correct fix
- rerun impacted verification
- update review evidence or add follow-up note

- [ ] **Step 5: Final status**

Report:

- implemented task slices
- verification commands and outcomes
- review archive path
- residual risks

## Integration Checks

Validate these paths together after all tasks land:

- `GET /api/v1/teacher/students/:id/evidence`
- `GET /api/v1/teacher/students/:id/attack-sessions`
- student analysis page data load and tab switch
- review archive generation/export

Most likely integration failure points:

- IDs returned as numbers in backend but expected as strings in frontend.
- time filters interpreted in local time instead of UTC.
- evidence pagination applied before source merge.
- `with_events=false` breaking frontend assumptions.
- AWD team-level evidence shown without `scope=team`.
- `StudentInsightPanel.vue` continuing to grow instead of shrinking.

## Rollback / Recovery Notes

- No database migration is planned; rollback is code-only.
- Backend route addition can be reverted independently if frontend keeps old evidence list fallback.
- Frontend workspace component can be hidden behind existing evidence tab if attack session API fails.
- Archive alignment should be reverted separately from realtime evidence if it introduces dependency-cycle risk.

## Residual Risks

- Query-time session aggregation may not scale for large AWD traffic datasets; keep limits and time filters explicit.
- Full archive reuse may require a cleaner shared read-model boundary in a later refactor.
- Existing uncommitted code already contains partial implementation; review must treat it as suspect until tests and independent review pass.
