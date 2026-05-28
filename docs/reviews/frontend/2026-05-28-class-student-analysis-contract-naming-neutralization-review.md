# Class Student Analysis Contract Naming Neutralization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-class-student-analysis-contract-naming-neutralization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/class-student-analysis-contract-naming-neutralization.md`
  - `docs/plan/impl-plan/2026-05-28-class-student-analysis-contract-naming-neutralization-plan.md`
  - `docs/reviews/frontend/2026-05-28-class-student-analysis-contract-naming-neutralization-review.md`
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
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于前端 `P1` 级结构收口，范围限定在 shared class/student analysis contract naming，不涉及 transport path 与 route name 迁移。
- Gate verdict：Implemented and re-validated

## Review focus

- shared class/student analysis contract 是否从 `Teacher*` 命名收口到中性 owner
- teacher / admin / shared feature 是否停止继续消费这组 teacher 前缀 DTO / query / payload
- 页面行为、请求路径和公共 wrapper 行为是否保持稳定

## Findings

- 无新的未收口 finding。

## Material findings

- 无。

## Senior implementation assessment

- 本轮如果完成，`class-students-workspace`、`student-analysis-workspace` 和其共享 widget 的 contract 层 teacher 前缀会明显缩小，能直接降低 `/platform/*` 共用 shared feature 时的角色语义漂移。
- 这刀不动 HTTP path 与 route name，边界合理；shared owner 收在 contract / API client / feature 消费层，比继续动 transport 语义更稳。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- same-context review only；当前 session 未执行独立子代理复核。
- teacher-only overview contract、teacher route name、teacher public wrapper 名与后端 teacher path 仍保留，这是当前明确保留的边界。

## Touched known-debt status

- 本轮 touched 的已知结构债是 shared class/student analysis contract 仍保留 `Teacher*` 前缀。
- 在本轮 touched surface 上，这条债务预期完成当前阶段收口；剩余 `P1` 将进一步压缩到显式 transport / public wrapper / route 语义。
