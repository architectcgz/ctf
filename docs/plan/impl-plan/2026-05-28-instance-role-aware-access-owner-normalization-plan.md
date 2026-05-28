> 状态：Current
> 事实源：`usePlatformInstanceManagementPage.ts`、`useInstances.ts`、当前实例目录 public API owner
> 替代：无

# Instance Role-Aware Access Owner Normalization Plan

## 目标

- 为实例目录共享 workflow 增加中立 role-aware access owner。
- 把 `usePlatformInstanceManagementPage.ts` 与 `useInstances.ts` 里的 role-specific 实例目录 API 选择统一收口。
- 同步更新实例目录相关护栏与 backlog 记录。

## 非目标

- 本轮不改 `getTeacherInstances()`、`destroyTeacherInstance()`、`getPlatformInstances()`、`destroyPlatformInstance()` 的 public wrapper 名。
- 本轮不改 `/api/v1/teacher/instances` 的后端 HTTP path、请求参数、分页结构或响应字段。
- 本轮不继续改实例目录 DTO 命名；`InstanceDirectory*` 已在上一刀收口完成。

## 输入依据

- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/awd-reviews.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`

## 当前结论

- 实例目录共享 contract 已经中性化，但 teacher / platform 两侧 feature 仍分别直连各自 public wrapper。
- 当前更合适的下一刀不是再改 transport owner，而是先把 feature 层的 role-aware API 选择收口成单点 facade。
- `api/awd-reviews.ts` 已经证明这种“共享 feature + role-aware access owner”模式在本仓库可行。

## 设计边界

### `api/instances.ts` 本轮负责

- list instance directory
- destroy managed instance
- 按 role 选择 admin / teacher API owner

### 实例目录 feature model 本轮继续负责

- 路由、筛选、分页、销毁 workflow
- loading / error / duplicate-action owner

### 本轮不动

- 实例目录 route shell
- 实例目录 DTO / page data / summary / filter naming
- 实例销毁 public wrapper 命名和 transport path

## 任务切片

### Slice 1：role-aware access owner 收口

- 目标：
  - 新增 `api/instances.ts`
  - 统一提供 role-aware 的实例目录 list / destroy access owner
  - `usePlatformInstanceManagementPage.ts`、`useInstances.ts` 改为只依赖这层中立 owner
- 验证：
  - `npm run test:run -- src/api/__tests__/instances.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts`
- Review focus：
  - 实例目录 feature 内是否不再散落 role-specific API import
  - role-aware access owner 是否集中成单点，而不是换个文件继续复制分支逻辑

### Slice 2：护栏与 backlog 同步

- 目标：
  - 更新实例目录 teacher / platform 测试源码断言
  - backlog 记录实例 access owner 这条 P1 结构债的本轮进展
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 护栏是否明确要求 feature 走中立 access owner
  - 不把 transport owner / DTO naming 混进同一刀

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/instances.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/teacher/__tests__/InstanceManagement.test.ts src/features/challenge-detail/model/useChallengeInstance.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收实例目录共享 feature 的 access owner，不继续改 `getTeacherInstances()` / `destroyTeacherInstance()` 这类 public wrapper 与后端 path；如果后续还要继续去 teacher 语义，需要再单独切 transport/public owner。
