> 状态：Current
> 事实源：`api/teaching/instances.ts`、`api/teacher/instances.ts`、`api/admin/teaching.ts`
> 替代：无

# Instance API Implementation Owner Neutralization Plan

## 目标

- 把实例目录共享实现层 `api/teaching/instances.ts` 的 teacher 命名函数收口成中性 owner。
- 保持 `api/teacher` / `api/admin` public API 不变，让 teacher 与 platform 继续以各自语义函数名暴露这组能力。
- 不修改后端 `/api/v1/teacher/instances*` 路径、权限语义和前端页面行为。

## 非目标

- 本轮不新增 `/api/v1/admin/instances*` 或其他平台专属实例目录 HTTP path。
- 本轮不改 `api/instances.ts` 的 role-aware facade 选择逻辑。
- 本轮不改实例目录 DTO、feature、页面交互和 teacher 班级目录读取链路。

## 输入依据

- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/instances.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `code/frontend/src/api/__tests__/instances.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `.harness/reuse-decisions/awd-review-api-implementation-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-awd-review-api-implementation-owner-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-28-awd-review-api-implementation-owner-neutralization-review.md`

## 当前结论

- route view、shared feature access owner、shared contract naming 都已经收口过，当前剩余 teacher 语义残片主要停在：
  - `api/teaching/instances.ts` 的共享实现函数名
  - `api/admin/teaching.ts` 对 teacher 命名函数的 alias re-export
- `api/teacher/instances.ts` 与 `api/admin/teaching.ts` 是更合适的 public owner 落点，因此角色语义应保留在 public wrapper，而不是继续渗进 `api/teaching` 共享实现层。

## 设计边界

### 本轮负责

- 中性化实例目录共享实现函数名
- 调整 teacher / admin public owner 的导出方式
- 补 raw-source 护栏、contract 记录和 backlog 进展

### 本轮不动

- 后端路由和 HTTP path
- 实例目录页面 / feature / facade 行为
- `getClasses()` 这类 teacher workspace 自有 owner

## 任务切片

### Slice 1：中性化 teaching 层实例目录实现符号

- 目标：
  - 在 `api/teaching/instances.ts` 提供中性实现函数，例如 `getInstanceDirectory()`、`destroyManagedInstance()`
  - 停止把 teacher 命名函数保留在共享实现层
- 预期改动：
  - `code/frontend/src/api/teaching/instances.ts`
- review focus：
  - HTTP path、normalize 行为、返回 DTO 完全不变

### Slice 2：收口 teacher / platform public owner

- 目标：
  - `api/teacher/instances.ts` 改成显式 teacher wrapper
  - `api/admin/teaching.ts` 改成显式 platform wrapper，不再 alias teacher 命名函数
- 预期改动：
  - `code/frontend/src/api/teacher/instances.ts`
  - `code/frontend/src/api/admin/teaching.ts`
- review focus：
  - teacher / admin public owner 继续稳定
  - 共享实现层不再暴露 teacher 语义符号

### Slice 3：测试与文档收尾

- 目标：
  - 补 API raw-source 护栏，记录实例目录本地 API owner 已进一步收口
- 预期改动：
  - `code/frontend/src/api/__tests__/teacher.test.ts`
  - `code/frontend/src/api/__tests__/admin.test.ts`
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-28-instance-api-implementation-owner-neutralization-review.md`

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/api/__tests__/admin.test.ts src/api/__tests__/instances.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 完成后，实例目录仍会保留 `/api/v1/teacher/instances*` 这组后端路径，以及 `getTeacherInstances()` / `destroyTeacherInstance()` 这组 teacher public wrapper；这是当前明确保留的 transport / public owner 语义，不属于共享实现 owner 漂移。
- `api/teaching` 目录内其他教学域函数是否仍有 teacher 命名残片，不在本轮范围；这刀只收实例目录 touched surface。
