# Instance Directory Contract Naming Neutralization 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-instance-directory-contract-naming-neutralization-implementation-plan.md`
  - files reviewed：
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/teaching/instances.ts`
    - `code/frontend/src/api/teacher/instances.ts`
    - `code/frontend/src/api/admin/teaching.ts`
    - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
    - `code/frontend/src/features/teacher-instances/model/useInstances.ts`
    - `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
    - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
    - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
    - `code/frontend/src/api/__tests__/teacher.test.ts`
    - `docs/contracts/api-contract-v1.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意本轮属于前端 P1 级结构收口，范围限定在共享实例目录 contract 的中性化命名。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `TeacherInstance*` 这组共享 contract 已同时服务 teacher / platform 两条实例目录流，本轮收口到 `InstanceDirectory*` 后，角色语义重新回到 public wrapper 和 HTTP path 层，不再停留在共享 DTO 上。
- `getTeacherInstances()`、`destroyTeacherInstance()` 与 `getPlatformInstances()` 的运行行为、请求参数和分页结构都没有变化，改动集中在 type surface 与 source-boundary 护栏，提交边界合理。
- teacher / platform 两侧实例目录 feature 以及 teacher page shell 已同步切到新命名，避免后续继续出现“共享 contract 已改，但某一侧页面还残留旧前缀”的半收口状态。

## Required re-validation

- `npm run test:run -- src/api/__tests__/teacher.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts`
- `npm run typecheck`
- `git diff --check -- code/frontend/src/api/contracts.ts code/frontend/src/api/teaching/instances.ts code/frontend/src/api/teacher/instances.ts code/frontend/src/api/admin/teaching.ts code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts code/frontend/src/features/teacher-instances/model/useInstances.ts code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue code/frontend/src/views/platform/__tests__/InstanceManage.test.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/api/__tests__/teacher.test.ts docs/contracts/api-contract-v1.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/instance-directory-contract-naming-neutralization.md docs/plan/impl-plan/2026-05-28-instance-directory-contract-naming-neutralization-implementation-plan.md docs/reviews/frontend/2026-05-28-instance-directory-contract-naming-neutralization-review.md`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- 这轮没有处理 `getTeacherInstances()` / `destroyTeacherInstance()` 这类 public wrapper 名和后端 `/api/v1/teacher/instances` transport path；如果后续要进一步去 teacher 语义，需要单独评估 public API owner 是否值得继续切。
- 其它仍带 teacher 前缀、但尚未进入本轮 touched surface 的共享 DTO 仍需后续单独切片。

## Touched known-debt status

- 本轮 touched 的已知结构债是实例目录共享 contract 仍保留 `TeacherInstance*` 前缀。
- 该债务在 touched surface 上已完成当前阶段收口；当前剩余 teacher 语义已进一步收敛到 public wrapper 命名和后端 transport path，而不是继续停留在 teacher / platform 共用的目录 DTO 上。
