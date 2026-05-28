> 状态：Current
> 事实源：`api/contracts.ts`、class/student analysis API client、shared workspace features
> 替代：无

# Class Student Analysis Contract Naming Neutralization Plan

## 目标

- 把共享班级 / 学员分析链路里仍带 `Teacher*` 前缀的 contract、query 与 payload 收口成中性命名。
- 保持 teacher / platform 页面行为、HTTP path 和 public API 调用方式不变，只收命名 owner。
- 继续压缩剩余 `P1` 的前端本地 teacher 语义残片。

## 非目标

- 本轮不改 `/api/v1/teacher/*` HTTP path。
- 本轮不改 teacher / platform route path 或 route name。
- 本轮不改 teacher-only overview contract，例如 `TeacherOverviewData` 及其 teacher dashboard 业务语义。
- 本轮不顺手改 writeup / manual review / AWD review 等已独立收口过的 contract 线。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/features/class-students-workspace/**`
- `code/frontend/src/features/student-analysis-workspace/**`
- `code/frontend/src/features/student-directory/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-class-report-export/**`
- `code/frontend/src/features/teacher-dashboard/**`
- `code/frontend/src/features/skill-profile/**`
- `code/frontend/src/components/teacher/**`
- `code/frontend/src/widgets/teacher-student-review-workspace/**`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `/platform/*` 不再直接 import teacher route view，但共享 class/student analysis feature 仍在 contract 层保留一组 teacher 前缀语义。
- 当前最直接的残留包括：
  - `TeacherClassInsightQueryData`
  - `TeacherClassSummaryData`
  - `TeacherClassTrendPoint` / `TeacherClassTrendData`
  - `TeacherClassReviewItemData` / `TeacherClassReviewData`
  - `TeacherStudentItem`
  - `TeacherEvidenceData`
  - `TeacherAttackSession*`
  - `TeacherStudentDirectoryParams`
  - `TeacherEvidenceQuery`
  - `TeacherClassReportExportPayload`
- 这些符号已经同时被 teacher / platform route view、shared feature、shared widget、teacher class report export 和 teacher dashboard builder 消费，不应继续保留 teacher owner 命名。

## 设计边界

### 本轮负责

- 中性化 shared class insight / student analysis contract 命名
- 同步 classes / students teaching API client 与 teacher/admin wrapper 的类型出口
- 同步 shared feature、组件、widget 和测试护栏
- 更新 contract 文档、backlog 与 review 证据

### 本轮不动

- teacher-only overview contract 名称
- HTTP path 与权限语义
- route name
- 更深层 teacher public wrapper 命名

## 任务切片

### Slice 1：收口 shared contract / query / payload 命名

- 目标：
  - 在 `api/contracts.ts` 将 shared class/student analysis DTO 收口到中性命名
  - 同步 `TeacherStudentDirectoryParams`、`TeacherEvidenceQuery`、`TeacherClassReportExportPayload`
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teaching/classes.ts`
  - `code/frontend/src/api/teaching/students.ts`
  - `code/frontend/src/api/teacher/classes.ts`
  - `code/frontend/src/api/teacher/students.ts`
  - `code/frontend/src/api/admin/teaching.ts`
- review focus：
  - request / response shape、normalize 行为与 HTTP path 不变

### Slice 2：同步 shared feature / component / widget 消费面

- 目标：
  - 让 class students、student analysis、student directory、teacher class report export、teacher dashboard、skill profile 这批共享消费面切到中性 contract
- 预期改动：
  - `code/frontend/src/features/class-students-workspace/**`
  - `code/frontend/src/features/student-analysis-workspace/**`
  - `code/frontend/src/features/student-directory/**`
  - `code/frontend/src/features/teacher-student-analysis/**`
  - `code/frontend/src/features/teacher-class-report-export/**`
  - `code/frontend/src/features/teacher-dashboard/**`
  - `code/frontend/src/features/skill-profile/**`
  - `code/frontend/src/components/teacher/**`
  - `code/frontend/src/widgets/teacher-student-review-workspace/**`
- review focus：
  - platform / teacher 共享 feature 在 contract 层不再保留 teacher 前缀
  - 行为 owner 不被顺手打散

### Slice 3：测试与文档收尾

- 目标：
  - 补 raw-source / API contract 护栏，记录当前 P1 剩余面收缩
- 预期改动：
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-class-student-analysis-contract-naming-neutralization-review.md`

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 完成后，teacher-only overview contract、teacher public wrapper 命名、teacher route name 与后端 teacher path 仍会保留；这是当前刻意保留的 role / transport 语义，不属于 shared contract owner 漂移。
- `platform instance management` 的 server-owned query contract 问题属于架构 review 里的另一条 debt，不在本轮范围。
