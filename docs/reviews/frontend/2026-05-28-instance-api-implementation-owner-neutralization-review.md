# Instance API Implementation Owner Neutralization 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-instance-api-implementation-owner-neutralization-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/instance-api-implementation-owner-neutralization.md`
  - `docs/plan/impl-plan/2026-05-28-instance-api-implementation-owner-neutralization-plan.md`
  - `docs/reviews/frontend/2026-05-28-instance-api-implementation-owner-neutralization-review.md`
  - `code/frontend/src/api/teaching/instances.ts`
  - `code/frontend/src/api/teacher/instances.ts`
  - `code/frontend/src/api/admin/teaching.ts`
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于实例目录本地 API owner 更深层收口，不涉及后端路径迁移。
- Gate verdict：same-context self-review passed，已重新验证；独立子代理 review gate 因当前工具委派限制未执行

## Review focus

- `api/teaching/instances.ts` 是否从 teacher 命名共享实现切到中性实现 owner
- `api/teacher/instances.ts` 与 `api/admin/teaching.ts` 是否把角色语义收口回 public owner
- teacher / admin public API 与 feature 调用面是否保持稳定

## Findings

- 无新的未收口 finding。

## Material findings

- 无。

## Senior implementation assessment

- `api/teaching/instances.ts` 已把实例目录共享实现层函数名收口到中性 `getInstanceDirectory()` / `destroyManagedInstance()`；共享实现层不再把 teacher 语义写死在 owner 上。
- `api/teacher/instances.ts` 与 `api/admin/teaching.ts` 已改为显式 public wrapper；teacher / platform owner 继续保留在 public API 层，而不是共享实现层。
- 这刀没有触碰后端 `/api/v1/teacher/instances*` path，也没有改实例目录页面和 feature 行为，边界合理。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/api/__tests__/admin.test.ts src/api/__tests__/instances.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 实例目录仍保留 `/api/v1/teacher/instances*` 这组后端路径，以及 teacher public wrapper 命名；这是当前刻意保留的 transport / public owner 语义，不属于前端本地共享 owner 漂移。
- `api/instances.ts` 仍依赖 teacher / platform public wrapper 做 role-aware 分派；这是当前设计的 access owner，不是本轮要继续下沉的残片。
- 按 pipeline 默认要求，这类 non-trivial 变更原本应补一份独立上下文 review；当前 session 的子代理工具只允许在用户明确要求委派时使用，因此本轮只完成了 same-context review 记录。

## Touched known-debt status

- 本轮 touched 的已知结构债是实例目录前端本地 API owner 还残留 teacher 命名共享实现与 admin alias teacher 函数。
- 在本轮 touched surface 上，这条债务已经完成当前阶段收口；当前实例目录线上前端本地残留的 teacher 语义，已进一步缩到后端 teacher HTTP path 和 teacher public wrapper 命名。
