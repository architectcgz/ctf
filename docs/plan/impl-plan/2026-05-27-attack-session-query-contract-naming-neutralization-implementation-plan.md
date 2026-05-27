> 状态：Current
> 事实源：`api/teacher/students.ts`、review workspace、student analysis workflow、fronted backlog
> 替代：无

# Attack Session Query Contract Naming Neutralization Implementation Plan

## 目标

- 把共享攻击会话筛选 contract `TeacherAttackSessionQuery` 收口成中性命名 `AttackSessionQuery`。
- 保持 teacher / platform 学员分析和复盘工作台的筛选、路由 query sync、请求参数与页面行为不变。

## 非目标

- 本轮不改 `TeacherAttackSessionData`、`TeacherAttackSessionResponseData` 等 response DTO。
- 本轮不改 `TeacherAWDReviewContestItemData`。
- 本轮不调整 `getStudentAttackSessions()` 的 API path、请求参数结构或 query sync 规则。

## 输入依据

- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherAttackSessionQuery` 当前已经穿过共享 review workspace，同时服务 teacher 和 platform 两边的复盘筛选流，继续挂 teacher 前缀不符合实际 owner 语义。
- 这组命名收口的真正变化面在 query type 和路由同步函数，不在 response 数据结构；因此可以保持切片较小。

## 任务切片

### Slice 1：收口共享 attack session query contract 名称

- 目标：
  - 在 `api/teacher/students.ts` 与 `api/teaching/students.ts` 提供中性 `AttackSessionQuery`。
- 预期改动：
  - `code/frontend/src/api/teacher/students.ts`
  - `code/frontend/src/api/teaching/students.ts`
  - `docs/contracts/api-contract-v1.md`
- review focus：
  - `getStudentAttackSessions()` 默认 query、透传字段和返回值不变

### Slice 2：同步 review workspace / student analysis 消费面

- 目标：
  - 让 student-analysis 与 review workspace 消费面切到 `AttackSessionQuery`。
- 预期改动：
  - `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
  - `code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue`
- review focus：
  - 复盘筛选 UI、URL query sync 和筛选参数透传不回归

### Slice 3：同步架构文档、backlog 与 review 证据

- 目标：
  - 对齐架构文档、backlog 和 review 对当前剩余命名债的描述。
- 预期改动：
  - `docs/contracts/api-contract-v1.md`
  - `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-attack-session-query-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按 query 命名切：撤销 `AttackSessionQuery` 与其消费面引用即可，不涉及响应 contract 或路由行为回退。

## 残余风险

- `TeacherAWDReviewContestItemData` 仍保留 teacher 前缀，后续需要独立切片。
