# Instance Role-Aware Access Owner Normalization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-instance-role-aware-access-owner-normalization-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/instance-role-aware-access-owner-normalization.md`
    - `docs/plan/impl-plan/2026-05-28-instance-role-aware-access-owner-normalization-plan.md`
    - `docs/reviews/frontend/2026-05-28-instance-role-aware-access-owner-normalization-review.md`
    - `code/frontend/src/api/instances.ts`
    - `code/frontend/src/api/__tests__/instances.test.ts`
    - `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
    - `code/frontend/src/features/teacher-instances/model/useInstances.ts`
    - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
    - `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
    - `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
    - `docs/contracts/api-contract-v1.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于实例目录共享 feature 的 role-aware access owner 收口。
- Gate verdict：Implemented and re-validated

## Findings

- 无新的未收口 findings。

## Material findings

- 无。

## Senior implementation assessment

- 已新增 `code/frontend/src/api/instances.ts` 作为实例目录共享 workflow 的中立 role-aware access owner，`list / destroy` 的角色分派不再分散在 teacher / platform 两个 feature model 里各自判断。
- `usePlatformInstanceManagementPage.ts` 与 `useInstances.ts` 已统一改为只依赖 `@/api/instances`，共享实例目录 workflow 不再继续直接双引 `@/api/admin` / `@/api/teacher` 的实例目录函数。
- 本轮把 owner 收口保持在 API facade 层，页面的筛选、分页、销毁确认、错误提示和跳转 owner 仍留在各自 feature model，不会把 workflow owner 反向抽空。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/instances.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- code/frontend/src/api/instances.ts code/frontend/src/api/__tests__/instances.test.ts code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts code/frontend/src/features/teacher-instances/model/useInstances.ts code/frontend/src/views/platform/__tests__/InstanceManage.test.ts code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts docs/contracts/api-contract-v1.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/instance-role-aware-access-owner-normalization.md docs/plan/impl-plan/2026-05-28-instance-role-aware-access-owner-normalization-plan.md docs/reviews/frontend/2026-05-28-instance-role-aware-access-owner-normalization-review.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这轮只收实例目录共享 workflow 的 access owner，不继续改 `getTeacherInstances()` / `destroyTeacherInstance()` / `getPlatformInstances()` 这类 public wrapper 名，也不改后端 `/api/v1/teacher/instances` path；如果后续还要继续去 teacher 语义，需要再单独切 transport/public owner。
- `useInstances.ts` 里的班级列表读取仍走 `getClasses()` teacher public owner；这属于 teacher workspace 自身 owner，不在本轮 access facade 收口边界里。

## Touched known-debt status

- 实例目录共享 workflow 在 touched surface 内，已从“teacher / platform feature 各自散落 role-specific API import”收口到单点 `api/instances.ts` owner。
- 这条 P1 当前已从 access owner 漂移收敛到 public wrapper 命名和 transport path，不再继续停留在 shared feature model 的 API 选择面。
